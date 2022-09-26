package manageiq

import (
	"encoding/json"
)

type ServiceCatalogs struct {
	MangeIQListResource
}

func (c *Client) GetServiceCatalogs() (*ServiceCatalogs, error) {
	builder := NewRequestBuilder(GET)
	_, err := builder.ResolveRequestURL(c.Authenticator.GetBaseURL(), "/service_catalogs", nil)
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
	s := &ServiceCatalogs{}
	if err := json.Unmarshal(resp.RawResult, &s); err != nil {
		return nil, err
	}
	return s, nil
}
