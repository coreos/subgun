package mailgun

import (
	"encoding/json"
	"net/url"
)

type ListMemberResponse struct {
	Member  ListMember `json:member`
	Message string     `json:message`
}

type ListMember struct {
	Address     string            `json:address`
	Subscribed  bool              `json:subscribed`
	Vars        map[string]string `json:vars`
	Name        string            `json:name`
	Description string            `json:description`
}

func (m *ListMember) setURLValues(v *url.Values) {
	if m.Subscribed == false {
		v.Set("subscribed", "False")
	}
	v.Set("address", m.Address)
	v.Set("name", m.Name)
	v.Set("description", m.Description)
	vars, _ := json.Marshal(m.Vars)
	v.Set("vars", string(vars))
}

func (c *Client) AddListMember(list string, m ListMember) (message string, err error) {
	v := url.Values{}
	m.setURLValues(&v)

	rsp, err := c.api("POST", "/lists/"+list+"/members", v)
	if err != nil {
		return
	}

	response := ListMemberResponse{}
	err = json.Unmarshal(rsp, &response)
	message = response.Message
	return
}

func (c *Client) UpdateListMember(list string, m ListMember) (message string, err error) {
	v := url.Values{}
	m.setURLValues(&v)


	rsp, err := c.api("PUT", "/lists/"+list+"/members"+m.Address, v)
	if err != nil {
		return
	}

	response := ListMemberResponse{}
	err = json.Unmarshal(rsp, &response)
	message = response.Message
	return
}
