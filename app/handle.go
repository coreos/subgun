package app

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/philips/go-mailgun"
)

func NewRouter(cfg *Config, mg *mailgun.Client) *mux.Router {
	hdlr := NewHandler(cfg, mg)

	r := mux.NewRouter()

	r.HandleFunc("/{action}/{list}", hdlr.initialHandler)
	r.HandleFunc("/{action}/{list}/confirm/{email}/{token}", hdlr.confirmationHandler)
	r.HandleFunc("/health", hdlr.healthHandler)

	return r
}

type Handler struct {
	cfg *Config
	mg  *mailgun.Client
}

func NewHandler(cfg *Config, mg *mailgun.Client) *Handler {
	return &Handler{cfg, mg}
}

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
func (h *Handler) listAllowed(list string) bool {
	for _, l := range h.cfg.Subscribegun.Lists {
		if l == list {
			return true
		}
	}
	return false
}

// confirmationHandler handles a confirmation link and changes the persons
// subscription state to "subscribed" or "unsubscribed" if the token matches.
func (h *Handler) confirmationHandler(w http.ResponseWriter, r *http.Request) {
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

	member, err := h.mg.GetListMember(listName, email)
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

	_, err = h.mg.UpdateListMember(listName, member)
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
func (h *Handler) initialHandler(w http.ResponseWriter, r *http.Request) {
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

	if h.listAllowed(listName) == false {
		http.Error(w, "Unknown list.", 403)
		return
	}

	email := r.FormValue("email")

	action := muxVars["action"]

	switch action {
	case "subscribe":
		h.handleSubscribe(w, listName, email)
	case "unsubscribe":
		h.handleUnsubscribe(w, listName, email)
	default:
		http.Error(w, fmt.Sprintf("Unknown action %s", action), 500)
	}
}

// confirmURL generates a url.URL for a confirmation link that the user will
// get via and must click to confirm a request.
func (h *Handler) confirmURL(action string, listName string, email string, key string) url.URL {
	u := url.URL{}
	u.Scheme = "http"
	u.Host = h.cfg.Subscribegun.Listen
	u.Path = path.Join("/", action, listName, "confirm", email, key)
	return u
}

func (h *Handler) addListMember(listName string, email string) (string, error) {
	// Generate the tokens for the user
	vars := map[string]string{
		"UnsubscribeToken": randomString(16),
		"SubscribeToken":   randomString(16),
	}
	member := mailgun.ListMember{email, false, vars, "", ""}
	key := vars["SubscribeToken"]

	_, err := h.mg.AddListMember(listName, member)
	return key, err
}

// handleSubscribe handles a subscribe request.
func (h *Handler) handleSubscribe(w http.ResponseWriter, listName string, email string) {
	var key string

	member, err := h.mg.GetListMember(listName, email)
	// Try to add the member if it doesn't exist
	if err != nil {
		key, err = h.addListMember(listName, email)
	} else {
		key = member.Vars["SubscribeToken"]
	}

	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}

	u := h.confirmURL("subscribe", listName, email, key)
	confirmMail := mail{
		from:    "no-reply@lists.coreos.com",
		to:      []string{email},
		subject: "confirm subscription to " + listName,
		text:    "Click below to confirm your subscription request to " + listName + ":\n\n" + u.String(),
	}
	_, err = h.mg.Send(&confirmMail)

	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}
}

// handleUnsubscribe handles an unsubscribe request.
func (h *Handler) handleUnsubscribe(w http.ResponseWriter, listName string, email string) {
	member, err := h.mg.GetListMember(listName, email)
	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}
	key := member.Vars["UnsubscribeToken"]

	u := h.confirmURL("unsubscribe", listName, email, key)
	confirmMail := mail{
		from:    "no-reply@lists.coreos.com",
		to:      []string{email},
		subject: "confirm unsubscribe to " + listName,
		text:    "Click below to confirm your unsubscribe request to " + listName + ":\n\n" + u.String(),
	}
	_, err = h.mg.Send(&confirmMail)

	if err != nil {
		http.Error(w, "Internal error", 500)
		fmt.Println(err)
		return
	}
}

func (h *Handler) healthHandler(w http.ResponseWriter, r *http.Request) {
	for _, addr := range h.cfg.Subscribegun.Lists {
		parts := strings.Split(addr, "@")
		domain := parts[len(parts)-1]
		_, _, err := h.mg.Stats(domain, 0, 0, []string{}, time.Now())
		if err != nil {
			http.Error(w, "FAIL", 500)
			return
		}
	}

	fmt.Fprint(w, "OK")
}
