package mailgun

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"
)

type Log struct {
	Hap       string `json:"hap"`
	CreatedAt string `json:"created_at"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	MessageId string `json:"message_id"`
}

func (l *Log) Time() time.Time {
	t, _ := time.Parse(time.RFC1123, l.CreatedAt)
	return t
}

func (c *Client) Logs(domain string, limit, skip int) (total int, res []Log, err error) {
	v := url.Values{}
	v.Set("limit", strconv.Itoa(limit))
	v.Set("skip", strconv.Itoa(skip))
	body, err := c.api("GET", "/"+domain+"/log", v)
	if err != nil {
		return
	}

	var j struct {
		Total int   `json:"total_count"`
		Items []Log `json:"items"`
	}

	err = json.Unmarshal(body, &j)
	if err != nil {
		return
	}
	total, res = j.Total, j.Items
	return
}
