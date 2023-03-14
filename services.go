package manageiq

import (
	"encoding/json"
	"net/url"
)

type Services struct {
	MangeIQListResource
}

type Service struct {
	MangeIQListResource
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	Retired        bool   `json:"retired"`
	RetiresOn      string `json:"retires_on"`
	LifecycleState string `json:"lifecycle_state"`
	VMs            []VM   `json:"vms"`
}

type VM struct {
	CreatedOn    string   `json:"created_on"`
	Description  string   `json:"description"`
	EMSCreatedOn string   `json:"ems_created_on"`
	EMSID        string   `json:"ems_id"`
	EMSRef       string   `json:"ems_ref"`
	ID           string   `json:"id"`
	IPAddresses  []string `json:"ipaddresses"`
	Name         string   `json:"name"`
	Vendor       string   `json:"vendor"`
	// ID for the vm in the powervs(backend)
	UIDEMS string `json:"uid_ems"`
	// Power state of the vm
	RawPowerState string `json:"raw_power_state"`
}

func (c *Client) ListServices(queries url.Values) (*Services, error) {
	builder := NewRequestBuilder(GET)
	_, err := builder.ResolveRequestURL(c.Authenticator.GetBaseURL(), "/services", nil, queries)
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
	s := &Services{}
	if err := json.Unmarshal(resp.RawResult, &s); err != nil {
		return nil, err
	}
	return s, nil
}

func (c *Client) GetService(id string, queries url.Values) (*Service, error) {
	builder := NewRequestBuilder(GET)
	_, err := builder.ResolveRequestURL(c.Authenticator.GetBaseURL(), "/services/"+id, nil, queries)
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
	s := &Service{}
	if err := json.Unmarshal(resp.RawResult, &s); err != nil {
		return nil, err
	}
	return s, nil
}

func (c *Client) UpdateService(id string, body interface{}) (*Service, error) {
	builder := NewRequestBuilder(POST)
	_, err := builder.ResolveRequestURL(c.Authenticator.GetBaseURL(), "/services/"+id, nil, nil)
	if err != nil {
		return nil, err
	}

	if _, err := builder.SetBodyContentJSON(body); err != nil {
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
	s := &Service{}
	if err := json.Unmarshal(resp.RawResult, &s); err != nil {
		return nil, err
	}
	return s, nil
}
