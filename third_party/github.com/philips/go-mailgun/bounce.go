package mailgun

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"
)

type Bounce struct {
	Code      int    `json:"code"`
	CreatedAt string `json:"created_at"`
	Error     string `json:"error"`
	Address   string `json:"address"`
}

func (b *Bounce) Time() time.Time {
	t, _ := time.Parse(time.RFC1123, b.CreatedAt)
	return t
}

func (c *Client) Bounces(domain string, limit, skip int) (total int, res []Bounce, err error) {
	v := url.Values{}
	v.Set("limit", strconv.Itoa(limit))
	v.Set("skip", strconv.Itoa(skip))
	body, err := c.api("GET", "/"+domain+"/bounces", v)
	if err != nil {
		return
	}

	var j struct {
		Total int      `json:"total_count"`
		Items []Bounce `json:"items"`
	}

	err = json.Unmarshal(body, &j)
	if err != nil {
		return
	}
	total, res = j.Total, j.Items
	return
}
