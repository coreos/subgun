package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
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

// subscribeConfirmHandler handles a confirmation link and changes the persons
// subscription state to "subscribed" if the token matches.
func subscribeConfirmHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, "No subscription token!", 400)
		return
	}

	member, err := mg.GetListMember(listName, email)
	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}

	// Success! Finish up the subscription.
	if token != member.Vars["SubscribeToken"] {
		http.Error(w, "Bad confirmation token", 400)
		return
	}

	member.Subscribed = true
	_, err = mg.UpdateListMember(listName, member)
	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}

	fmt.Fprintf(w, "Success! You are now subscribed to %s", listName)

	return
}

// subscribeHandler adds the requested email to the list as unsubscribed and stores
// a Subscribe token. The token is sent to the email address to ensure the owner
// actually subscribed.
func subscribeHandler(w http.ResponseWriter, r *http.Request) {
	muxVars := mux.Vars(r)

	listName := muxVars["list"]
	if len(listName) == 0 {
		http.Error(w, "No list specified!", 404)
		return
	}

	if listAllowed(listName) == false {
		// TODO: check the right return code for not-allowed
		http.Error(w, "Unknown list.", 404)
		return
	}

	email := r.FormValue("email")

	// Generate the tokens for the user
	vars := map[string]string{
		"UnsubscribeToken": randomString(16),
		"SubscribeToken":   randomString(16),
	}
	member := mailgun.ListMember{email, false, vars, "", ""}

	key := vars["SubscribeToken"]

	_, err := mg.AddListMember(listName, member)
	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}

	confirmMail := mail{
		from:    "no-reply@lists.coreos.com",
		to:      []string{email},
		subject: "confirm subscription to " + listName,
		text:    "click here to confirm http://" + cfg.Subscribegun.Hostname + "/subscribe/" + listName + "/confirm/" + email + "/" + key,
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

	r := mux.NewRouter()

	// TODO: add a secret seed in here
	rand.Seed(time.Now().UTC().UnixNano())

	// subscription handling
	r.HandleFunc("/subscribe/{list}", subscribeHandler)
	r.HandleFunc("/subscribe/{list}/confirm/{email}/{token}", subscribeConfirmHandler)

	// TODO: unsubscribe handling
	// http.HandleFunc("/unsubscribe", handler)
	// http.HandleFunc("/unsubscribe/confirm", handler)

	http.ListenAndServe(":8080", r)
}
