package manageiq

import (
	"encoding/json"
)

type MangeIQListResource struct {
	Name      string     `json:"name"`
	Count     float64    `json:"count"`
	SubCount  float64    `json:"subcount"`
	Pages     float64    `json:"pages"`
	Resources []Resource `json:"resources"`
	Actions   []Action   `json:"actions"`
	Links     Links      `json:"links"`
}

type Group struct {
	CreatedOn           string   `json:"created_on"`
	DetailedDescription string   `json:"detailed_description"`
	Actions             []Action `json:"actions"`
	GroupType           string   `json:"group_type"`
	Sequence            float64  `json:"sequence"`
	UpdatedOn           string   `json:"updated_on"`
	Settings            string   `json:"settings"`
	TenantID            string   `json:"tenant_id"`
	Href                string   `json:"href"`
	ID                  string   `json:"id"`
	Description         string   `json:"description"`
}

type Groups struct {
	MangeIQListResource
}

type Links struct {
	Self  string `json:"self"`
	First string `json:"first"`
	Last  string `json:"last"`
}

type Action struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	Href   string `json:"href"`
}

type Resource struct {
	Href string `json:"href"`
}

func (c *Client) GetGroups() (*Groups, error) {
	builder := NewRequestBuilder(GET)
	_, err := builder.ResolveRequestURL(c.Authenticator.GetBaseURL(), "/groups", nil)
	if err != nil {
		return nil, err
	}
	req, err := builder.Build()
	if err != nil {
		return nil, err
	}

	resp, err := c.sendRequest(req, nil)
	if err != nil {
		return nil, err
	}
	g := &Groups{}
	if err := json.Unmarshal(resp.RawResult, &g); err != nil {
		return nil, err
	}
	return g, nil
}

func (c *Client) GetGroup(id string) (*Group, error) {
	builder := NewRequestBuilder(GET)
	_, err := builder.ResolveRequestURL(c.Authenticator.GetBaseURL(), "/groups/"+id, nil)
	if err != nil {
		return nil, err
	}
	req, err := builder.Build()
	if err != nil {
		return nil, err
	}

	resp, err := c.sendRequest(req, nil)
	if err != nil {
		return nil, err
	}
	g := &Group{}
	if err := json.Unmarshal(resp.RawResult, &g); err != nil {
		return nil, err
	}
	return g, nil
}
