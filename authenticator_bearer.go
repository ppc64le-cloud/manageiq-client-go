package manageiq

import (
	"fmt"
	"net/http"
)

var _ Authenticator = &BearerAuthenticator{}

type BearerAuthenticator struct {
	Token   string
	BaseURL string
	Client  *http.Client
}

func (b *BearerAuthenticator) Authenticate(request *http.Request) error {
	if err := b.Validate(); err != nil {
		return err
	}
	if b.BaseURL == "" {
		b.BaseURL = defaultBaseURL
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", b.Token))
	return nil
}

func (b *BearerAuthenticator) Validate() error {
	if b.Token == "" {
		return fmt.Errorf("token can't be empty")
	}
	return nil
}

func (b *BearerAuthenticator) GetBaseURL() string {
	if b.BaseURL != "" {
		return b.BaseURL
	}
	return defaultBaseURL
}
