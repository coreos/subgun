package mailgun

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const MAX_FORM_SIZE = 2 * 1024 * 1024

type Event struct {
	name   string
	time   time.Time
	rcpt   string
	msgid  string
	reason string
}

func (e *Event) Name() string    { return e.name }
func (e *Event) Time() time.Time { return e.time }
func (e *Event) Rcpt() string    { return e.rcpt }
func (e *Event) MsgId() string   { return e.msgid }
func (e *Event) Reason() string  { return e.reason }

var (
	ErrMethodNotAllowed = fmt.Errorf("method not allowed")
	ErrBadSignature     = fmt.Errorf("bad signature")
	ErrInvalidTimestamp = fmt.Errorf("invalid timestamp")
	ErrInvalidForm      = fmt.Errorf("invalid form data")
)

type Webhook struct {
	key string
}

func NewWebhook(key string) *Webhook {
	return &Webhook{key}
}

func (wh *Webhook) Handle(w http.ResponseWriter, req *http.Request) (evt *Event, err error) {
	if req.Method != "POST" {
		err = ErrMethodNotAllowed
		w.Header().Set("Allow", "POST")
		http.Error(w, "only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	if err = req.ParseMultipartForm(MAX_FORM_SIZE); err != nil {
		err = ErrInvalidForm
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	ts := req.Form.Get("timestamp")
	if !wh.Verify(ts, req.Form.Get("token"), req.Form.Get("signature")) {
		err = ErrBadSignature
		http.Error(w, "bad signature", http.StatusForbidden)
		return
	}

	unix, err := strconv.ParseInt(ts, 10, 64) // second since Unix epoch
	if err != nil {
		err = ErrInvalidTimestamp
		http.Error(w, "invalid timestamp", http.StatusBadRequest)
		return
	}
	evt = &Event{
		time:   time.Unix(unix, 0),
		name:   req.Form.Get("event"),
		rcpt:   req.Form.Get("recipient"),
		msgid:  req.Form.Get("Message-Id"),
		reason: req.Form.Get("reason"),
	}
	switch evt.name {
	case "bounced":
		evt.reason = fmt.Sprintf("%s %s %s", req.Form.Get("code"), req.Form.Get("error"), req.Form.Get("notification"))
	case "dropped":
		evt.reason = fmt.Sprintf("%s %s", req.Form.Get("reason"), req.Form.Get("description"))
	}
	return
}

func (wh *Webhook) Verify(timestamp, token, signature string) bool {
	h := hmac.New(sha256.New, []byte(wh.key))
	io.WriteString(h, timestamp)
	io.WriteString(h, token)
	calcSig := h.Sum(nil)
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	if len(sig) != len(calcSig) {
		return false
	}
	return subtle.ConstantTimeCompare(sig, calcSig) == 1
}
