package manageiq

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const ()

// LogLevel defines a type that can set the desired level of logging the SDK will generate.
type LogLevel uint

const (
	// LogOff will disable all SDK logging. This is the default log level
	LogOff LogLevel = iota * (1 << 8)
	// LogDebug will enable detailed SDK debug logs. It will log requests (including arguments),
	// response and body contents.
	LogDebug
	// LogInfo will log SDK request (not including arguments) and responses.
	LogInfo
)

func (l LogLevel) shouldLog(v LogLevel) bool {
	return l > v || l&v == v
}

// Client contains parameters for configuring the SDK.
type Client struct {
	// Authenticator for the client
	Authenticator Authenticator

	// apiKey used to talk ManageIQ
	apiKey string

	// Logging level for SDK generated logs
	LogLevel LogLevel

	// No need to set -- for testing only
	HTTPClient *http.Client
}

type ClientParams struct {
	BaseURL  string
	LogLevel LogLevel
}

func NewClient(authenticator Authenticator, param ClientParams) *Client {
	return &Client{
		Authenticator: authenticator,
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
	}
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{}
}

func (c *Client) GetServices() (*DetailedResponse, error) {
	builder := NewRequestBuilder(GET)
	_, err := builder.ResolveRequestURL(c.Authenticator.GetBaseURL(), "/services", nil)
	if err != nil {
		return nil, err
	}
	req, err := builder.Build()
	if err != nil {
		return nil, err
	}

	return c.sendRequest(req, nil)
}

func (c *Client) sendRequest(req *http.Request, v interface{}) (*DetailedResponse, error) {
	if err := c.Authenticator.Authenticate(req); err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json; charset=utf-8")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes errorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return nil, errors.New(errRes.Message)
		}

		return nil, fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	detailedResponse := &DetailedResponse{
		StatusCode: res.StatusCode,
		Headers:    res.Header,
		Result:     v,
		RawResult:  body,
	}
	if err = json.NewDecoder(bytes.NewReader(body)).Decode(&detailedResponse.Result); err != nil {
		return nil, err
	}

	return detailedResponse, nil
}
