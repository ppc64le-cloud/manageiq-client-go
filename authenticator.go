package manageiq

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultBaseURL = "https://127.0.0.1:8443/api"

	POST = "POST"
	GET  = "GET"

	ERRORMSG_SERVICE_URL_MISSING = "service GetBaseURL is empty"
	ERRORMSG_SERVICE_URL_INVALID = "error parsing service GetBaseURL: %s"
	ERRORMSG_PATH_PARAM_EMPTY    = "path parameter '%s' is empty"
)

type Authenticator interface {
	Authenticate(request *http.Request) error
	Validate() error
	GetBaseURL() string
}

type BasicAuthenticator struct {
	UserName string
	Password string
	BaseURL  string
	Client   *http.Client
	Insecure bool
}

func (a *BasicAuthenticator) Validate() error {
	if a.UserName == "" || a.Password == "" {
		return fmt.Errorf("username or password can't be empty")
	}
	return nil
}

func (a *BasicAuthenticator) GetBaseURL() string {
	if a.BaseURL != "" {
		return a.BaseURL
	}
	return defaultBaseURL
}

func (a *BasicAuthenticator) Authenticate(request *http.Request) error {
	if err := a.Validate(); err != nil {
		return err
	}
	if a.BaseURL == "" {
		a.BaseURL = defaultBaseURL
	}
	token, err := a.GetToken()
	if err != nil {
		return err
	}

	request.Header.Set("X-Auth-Token", token.AccessToken)
	return nil
}

// DetailedResponse holds the response information received from the server.
type DetailedResponse struct {

	// The HTTP status code associated with the response.
	StatusCode int

	// The HTTP headers contained in the response.
	Headers http.Header

	// Result - this field will contain the result of the operation (obtained from the response body).
	//
	// If the operation was successful and the response body contains a JSON response, it is un-marshalled
	// into an object of the appropriate type (defined by the particular operation), and the Result field will contain
	// this response object.  If there was an error while un-marshalling the JSON response body, then the RawResult field
	// will be set to the byte array containing the response body.
	//
	// Alternatively, if the generated SDK code passes in a result object which is an io.ReadCloser instance,
	// the JSON un-marshalling step is bypassed and the response body is simply returned in the Result field.
	// This scenario would occur in a situation where the SDK would like to provide a streaming model for large JSON
	// objects.
	//
	// If the operation was successful and the response body contains a non-JSON response,
	// the Result field will be an instance of io.ReadCloser that can be used by generated SDK code
	// (or the application) to read the response data.
	//
	// If the operation was unsuccessful and the response body contains a JSON error response,
	// this field will contain an instance of map[string]interface{} which is the result of un-marshalling the
	// response body as a "generic" JSON object.
	// If the JSON response for an unsuccessful operation could not be properly un-marshalled, then the
	// RawResult field will contain the raw response body.
	Result interface{}

	// This field will contain the raw response body as a byte array under these conditions:
	// 1) there was a problem un-marshalling a JSON response body -
	// either for a successful or unsuccessful operation.
	// 2) the operation was unsuccessful, and the response body contains a non-JSON response.
	RawResult []byte
}

type TokenResponse struct {
	AccessToken string `json:"auth_token"`
	TokenTTL    int64  `json:"token_ttl"`
	ExpiresOn   string `json:"expires_on"`
}

func (a *BasicAuthenticator) GetToken() (*TokenResponse, error) {
	builder := NewRequestBuilder(GET)
	_, err := builder.ResolveRequestURL(a.GetBaseURL(), "/auth", nil)
	if err != nil {
		return nil, err
	}
	req, err := builder.Build()
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(a.UserName, a.Password)

	resp, err := a.client().Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buff := new(bytes.Buffer)
		_, _ = buff.ReadFrom(resp.Body)

		// Create a DetailedResponse to be included in the error below.
		detailedResponse := &DetailedResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			RawResult:  buff.Bytes(),
		}

		iamErrorMsg := string(detailedResponse.RawResult)
		if iamErrorMsg == "" {
			iamErrorMsg =
				fmt.Sprintf("unexpected status code %d received from IAM token server %s", detailedResponse.StatusCode, builder.URL)
		}
		return nil, NewAuthenticationError(detailedResponse, fmt.Errorf(iamErrorMsg))
	}

	tokenResponse := &TokenResponse{}
	_ = json.NewDecoder(resp.Body).Decode(tokenResponse)
	defer resp.Body.Close()

	return tokenResponse, nil
}

// AuthenticationError describes the error returned when authentication fails
type AuthenticationError struct {
	Response *DetailedResponse
	Err      error
}

func (e *AuthenticationError) Error() string {
	return e.Err.Error()
}

func NewAuthenticationError(response *DetailedResponse, err error) *AuthenticationError {
	return &AuthenticationError{
		Response: response,
		Err:      err,
	}
}

func (a *BasicAuthenticator) client() *http.Client {
	if a.Client == nil {
		a.Client = http.DefaultClient
	}
	if a.Insecure {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return a.Client
}

// NewRequestBuilder initiates a new request.
func NewRequestBuilder(method string) *RequestBuilder {
	return &RequestBuilder{
		Method: method,
		Header: make(http.Header),
		Query:  make(map[string][]string),
		Form:   make(map[string][]FormData),
	}
}

// RequestBuilder is used to build an HTTP Request instance.
type RequestBuilder struct {
	Method string
	URL    *url.URL
	Header http.Header
	Body   io.Reader
	Query  map[string][]string
	Form   map[string][]FormData

	// EnableGzipCompression indicates whether or not request bodies
	// should be gzip-compressed.
	// This field has no effect on response bodies.
	// If enabled, the Body field will be gzip-compressed and
	// the "Content-Encoding" header will be added to the request with the
	// value "gzip".
	EnableGzipCompression bool

	// RequestContext is an optional Context instance to be associated with the
	// http.Request that is constructed by the Build() method.
	ctx context.Context
}

// FormData stores information for form data.
type FormData struct {
	fileName    string
	contentType string
	contents    interface{}
}

// WithContext sets "ctx" as the Context to be associated with
// the http.Request instance that will be constructed by the Build() method.
func (requestBuilder *RequestBuilder) WithContext(ctx context.Context) *RequestBuilder {
	requestBuilder.ctx = ctx
	return requestBuilder
}

// AddQuery adds a query parameter name and value to the request.
func (requestBuilder *RequestBuilder) AddQuery(name string, value string) *RequestBuilder {
	requestBuilder.Query[name] = append(requestBuilder.Query[name], value)
	return requestBuilder
}

// AddHeader adds a header name and value to the request.
func (requestBuilder *RequestBuilder) AddHeader(name string, value string) *RequestBuilder {
	requestBuilder.Header[name] = []string{value}
	return requestBuilder
}

// AddFormData adds a new mime part (constructed from the input parameters)
// to the request's multi-part form.
func (requestBuilder *RequestBuilder) AddFormData(fieldName string, fileName string, contentType string,
	contents interface{}) *RequestBuilder {
	if fileName == "" {
		if file, ok := contents.(*os.File); ok {
			if !((os.File{}) == *file) { // if file is not empty
				name := filepath.Base(file.Name())
				fileName = name
			}
		}
	}
	requestBuilder.Form[fieldName] = append(requestBuilder.Form[fieldName], FormData{
		fileName:    fileName,
		contentType: contentType,
		contents:    contents,
	})
	return requestBuilder
}

// SetBodyContentJSON sets the body content from a JSON structure.
func (requestBuilder *RequestBuilder) SetBodyContentJSON(bodyContent interface{}) (*RequestBuilder, error) {
	requestBuilder.Body = new(bytes.Buffer)
	err := json.NewEncoder(requestBuilder.Body.(io.Writer)).Encode(bodyContent)
	return requestBuilder, err
}

func (requestBuilder *RequestBuilder) ResolveRequestURL(serviceURL string, path string, pathParams map[string]string) (*RequestBuilder, error) {
	if serviceURL == "" {
		return requestBuilder, fmt.Errorf(ERRORMSG_SERVICE_URL_MISSING)
	}

	urlString := serviceURL

	// If we have a non-empty "path" input parameter, then process it for possible path param references.
	if path != "" {

		// If path parameter values were passed in, then for each one, replace any references to it
		// within "path" with the path parameter's encoded value.
		if len(pathParams) > 0 {
			for k, v := range pathParams {
				if v == "" {
					return requestBuilder, fmt.Errorf(ERRORMSG_PATH_PARAM_EMPTY, k)
				}
				encodedValue := url.PathEscape(v)
				ref := fmt.Sprintf("{%s}", k)
				path = strings.ReplaceAll(path, ref, encodedValue)
			}
		}

		// Next, we need to append "path" to "urlString".
		// We need to pay particular attention to any trailing slash on "urlString" and
		// a leading slash on "path".  Ultimately, we do not want a double slash.
		if strings.HasSuffix(urlString, "/") {
			// If urlString has a trailing slash, then make sure path does not have a leading slash.
			path = strings.TrimPrefix(path, "/")
		} else {
			// If urlString does not have a trailing slash and path does not have a
			// leading slash, then append a slash to urlString.
			if !strings.HasPrefix(path, "/") {
				urlString += "/"
			}
		}

		urlString += path
	}

	var URL *url.URL

	URL, err := url.Parse(urlString)
	if err != nil {
		return requestBuilder, fmt.Errorf(ERRORMSG_SERVICE_URL_INVALID, err.Error())
	}

	requestBuilder.URL = URL
	return requestBuilder, nil
}

func (requestBuilder *RequestBuilder) Build() (req *http.Request, err error) {
	// Create the request
	req, err = http.NewRequest(requestBuilder.Method, requestBuilder.URL.String(), requestBuilder.Body)
	if err != nil {
		return
	}

	// Headers
	req.Header = requestBuilder.Header

	query := req.URL.Query()
	for k, l := range requestBuilder.Query {
		for _, v := range l {
			query.Add(k, v)
		}
	}

	// Encode query
	req.URL.RawQuery = query.Encode()

	// Finally, if a Context should be associated with the new Request instance, then set it.
	if requestBuilder.ctx != nil {
		req = req.WithContext(requestBuilder.ctx)
	}

	return
}
