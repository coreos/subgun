package mailgun

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"
)

type Stat struct {
	Count     int            `json:"total_count"`
	CreatedAt string         `json:"created_at"`
	Tags      map[string]int `json:"tags"`
	Id        string         `json:"id"`
	Event     string         `json:"event"`
}

func (s *Stat) Time() time.Time {
	t, _ := time.Parse(time.RFC1123, s.CreatedAt)
	return t
}

func (c *Client) Stats(domain string, limit, skip int, events []string, startDate time.Time) (total int, res []Stat, err error) {
	v := url.Values{}
	v.Set("limit", strconv.Itoa(limit))
	v.Set("skip", strconv.Itoa(skip))
	for _, evt := range events {
		v.Add("event", evt)
	}
	// Gotcha: If startDate is specified, the result is ordered by date ascendingly (earliest first). Otherwise the result is ordered by date descendingly (latest first)
	if !startDate.IsZero() {
		v.Set("start-date", startDate.Format("2006-01-02")) // ISO 8601 date format
	}
	body, err := c.api("GET", "/"+domain+"/stats", v)
	if err != nil {
		return
	}

	var j struct {
		Total int    `json:"total_count"`
		Items []Stat `json:"items"`
	}

	err = json.Unmarshal(body, &j)
	if err != nil {
		return
	}
	total, res = j.Total, j.Items
	return
}
