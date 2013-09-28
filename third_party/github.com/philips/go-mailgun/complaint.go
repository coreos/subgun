package mailgun

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"
)

type Complaint struct {
	Count     int    `json:"count"`
	CreatedAt string `json:"created_at"`
	Address   string `json:"address"`
}

func (c *Complaint) Time() time.Time {
	t, _ := time.Parse(time.RFC1123, c.CreatedAt)
	return t
}

func (c *Client) Complaints(domain string, limit, skip int) (total int, res []Complaint, err error) {
	v := url.Values{}
	v.Set("limit", strconv.Itoa(limit))
	v.Set("skip", strconv.Itoa(skip))
	body, err := c.api("GET", "/"+domain+"/complaints", v)
	if err != nil {
		return
	}

	var j struct {
		Total int         `json:"total_count"`
		Items []Complaint `json:"items"`
	}

	err = json.Unmarshal(body, &j)
	if err != nil {
		return
	}
	total, res = j.Total, j.Items
	return
}
