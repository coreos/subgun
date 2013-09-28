package mailgun

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
)

type Mail interface {
	From() string
	To() []string
	Cc() []string
	Bcc() []string
	Subject() string
	Html() string
	Text() string
	Headers() map[string]string
	Options() map[string]string
	Variables() map[string]string
}

var EMAIL_DOMAIN_RE = regexp.MustCompile(`[^<>]+<?.+@([^<>]+)>?`)

func (c *Client) Send(m Mail) (msgId string, err error) {
	match := EMAIL_DOMAIN_RE.FindStringSubmatch(m.From())
	if len(match) != 2 {
		err = fmt.Errorf("invalid From address: %s", m.From())
		return
	}
	domain := match[1]
	v := url.Values{}
	v.Set("from", m.From())
	for _, to := range m.To() {
		v.Add("to", to)
	}
	for _, cc := range m.Cc() {
		v.Add("cc", cc)
	}
	for _, bcc := range m.Bcc() {
		v.Add("bcc", bcc)
	}
	v.Set("subject", m.Subject())
	v.Set("html", m.Html())
	v.Set("text", m.Text())

	for k, e := range m.Headers() {
		v.Add("h:"+k, e)
	}
	for k, e := range m.Options() {
		v.Add("o:"+k, e)
	}
	for k, e := range m.Variables() {
		v.Add("v:"+k, e)
	}

	rsp, err := c.api("POST", "/"+domain+"/messages", v)
	if err != nil {
		return
	}
	var res struct {
		Message string `json:"message"`
		Id      string `json:"id"`
	}
	err = json.Unmarshal(rsp, &res)
	msgId = res.Id
	return
}
