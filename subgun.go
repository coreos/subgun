package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/philips/go-mailgun"
)

type config struct {
	Subscribegun struct {
		Hostname string
		Lists    []string
	}
	Mailgun struct {
		Key string
	}
}

var mg *mailgun.Client
var cfg config

type mail struct {
	from      string
	to        []string
	cc        []string
	bcc       []string
	subject   string
	html      string
	text      string
	headers   map[string]string
	options   map[string]string
	variables map[string]string
}

func (m *mail) From() string                 { return m.from }
func (m *mail) To() []string                 { return m.to }
func (m *mail) Cc() []string                 { return m.cc }
func (m *mail) Bcc() []string                { return m.bcc }
func (m *mail) Subject() string              { return m.subject }
func (m *mail) Html() string                 { return m.html }
func (m *mail) Text() string                 { return m.text }
func (m *mail) Headers() map[string]string   { return m.headers }
func (m *mail) Options() map[string]string   { return m.options }
func (m *mail) Variables() map[string]string { return m.variables }

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

// listAllowed checks to ensure that the list is in the configuration file as
// a publicly subscribable list.
func listAllowed(list string) bool {
	for _, l := range cfg.Subscribegun.Lists {
		if l == list {
			return true
		}
	}
	return false
}

// confirmationHandler handles a confirmation link and changes the persons
// subscription state to "subscribed" or "unsubscribed" if the token matches.
func confirmationHandler(w http.ResponseWriter, r *http.Request) {
	muxVars := mux.Vars(r)

	listName := muxVars["list"]
	if len(listName) == 0 {
		http.Error(w, "No list specified!", 404)
		return
	}

	email := muxVars["email"]
	if len(email) == 0 {
		http.Error(w, "No email address!", 400)
		return
	}

	token := muxVars["token"]
	if len(token) == 0 {
		http.Error(w, "No token!", 400)
		return
	}

	member, err := mg.GetListMember(listName, email)
	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}

	action := muxVars["action"]
	if token != member.Vars[strings.Title(action)+"Token"] {
		http.Error(w, "Bad confirmation token", 400)
		return
	}

	if action == "subscribe" {
		member.Subscribed = true
		fmt.Fprintf(w, "Success! You are now subscribed to %s", listName)
	} else if action == "unsubscribe" {
		member.Subscribed = false
		fmt.Fprintf(w, "Success! You are now unsubscribed from %s", listName)
	} else {
		http.Error(w, fmt.Sprintf("Unknown action %s", action), 500)
		return
	}

	_, err = mg.UpdateListMember(listName, member)
	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}
	return
}

// initialHandler subscribes or unsubscribes the requested email from the list
// and sends a confirmation email with a token to ensure the email owner actually
// requested the action.
func initialHandler(w http.ResponseWriter, r *http.Request) {
	muxVars := mux.Vars(r)

	// Deal with CORS stuff
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Add("Access-Control-Allow-Headers", "Origin")
	if w.Header().Get("access-control-request-headers") != "" {
		return
	}

	listName := muxVars["list"]
	if len(listName) == 0 {
		http.Error(w, "No list specified!", 404)
		return
	}

	if listAllowed(listName) == false {
		http.Error(w, "Unknown list.", 403)
		return
	}

	email := r.FormValue("email")

	action := muxVars["action"]

	switch action {
	case "subscribe":
		handleSubscribe(w, listName, email)
	case "unsubscribe":
		handleUnsubscribe(w, listName, email)
	default:
		http.Error(w, fmt.Sprintf("Unknown action %s", action), 500)
	}
}

// confirmURL generates a url.URL for a confirmation link that the user will
// get via and must click to confirm a request.
func confirmURL(action string, listName string, email string, key string) url.URL {
	u := url.URL{}
	u.Scheme = "http"
	u.Host = cfg.Subscribegun.Hostname
	u.Path = path.Join("/", action, listName, "confirm", email, key)
	return u
}

func addListMember(listName string, email string) (string, error) {
	// Generate the tokens for the user
	vars := map[string]string{
		"UnsubscribeToken": randomString(16),
		"SubscribeToken":   randomString(16),
	}
	member := mailgun.ListMember{email, false, vars, "", ""}
	key := vars["SubscribeToken"]

	_, err := mg.AddListMember(listName, member)
	return key, err
}

// handleSubscribe handles a subscribe request.
func handleSubscribe(w http.ResponseWriter, listName string, email string) {
	var key string

	member, err := mg.GetListMember(listName, email)
	// Try to add the member if it doesn't exist
	if err != nil {
		key, err = addListMember(listName, email)
	} else {
		key = member.Vars["SubscribeToken"]
	}

	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}

	u := confirmURL("subscribe", listName, email, key)
	confirmMail := mail{
		from:    "no-reply@lists.coreos.com",
		to:      []string{email},
		subject: "confirm subscription to " + listName,
		text:    "click here to confirm your subscription request to " + listName + ":\n" + u.String(),
	}
	_, err = mg.Send(&confirmMail)

	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}
}

// handleUnsubscribe handles an unsubscribe request.
func handleUnsubscribe(w http.ResponseWriter, listName string, email string) {
	member, err := mg.GetListMember(listName, email)
	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}
	key := member.Vars["UnsubscribeToken"]

	u := confirmURL("unsubscribe", listName, email, key)
	confirmMail := mail{
		from:    "no-reply@lists.coreos.com",
		to:      []string{email},
		subject: "confirm unsubscribe to " + listName,
		text:    "click here to confirm your unsubscribe request to " + listName + ":\n" + u.String(),
	}
	_, err = mg.Send(&confirmMail)

	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}
}

func main() {
	configBytes, err := ioutil.ReadFile(os.Args[1])

	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(configBytes, &cfg)
	if err != nil {
		panic(err)
	}

	mg = mailgun.New(cfg.Mailgun.Key)

	// TODO: add a secret seed in here
	rand.Seed(time.Now().UTC().UnixNano())

	r := mux.NewRouter()

	r.HandleFunc("/{action}/{list}", initialHandler)
	r.HandleFunc("/{action}/{list}/confirm/{email}/{token}", confirmationHandler)

	http.ListenAndServe(":8080", r)
}
