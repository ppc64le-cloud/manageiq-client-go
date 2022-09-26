package manageiq

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v11"
)

var _ Authenticator = &KeycloakAuthenticator{}

const (
	defaultKeycloakBaseURL = "http://localhost:8080"
)

type KeycloakAuthenticator struct {
	Client *http.Client

	// Debug flag to the keycloak
	Debug bool

	// KeycloakBaseURL is the keycloak server URL used for all the authentication
	KeycloakBaseURL string

	// Realm used for generating the token
	Realm string

	// ClientID used for generating the token
	ClientID string

	// ClientSecret used for generating the token
	ClientSecret string

	// UserName used for generating the token
	UserName string

	// Password used for generating the token
	Password string

	// BaseURL for the ManageIQ
	BaseURL string

	token *gocloak.JWT

	keycloakClient gocloak.GoCloak
}

func (k *KeycloakAuthenticator) Authenticate(request *http.Request) error {
	if err := k.Validate(); err != nil {
		return err
	}

	//var err error
	if err := k.generateToken(); err != nil {
		return err
	}

	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", k.token.AccessToken))
	return nil
}

func (k *KeycloakAuthenticator) Validate() error {
	if k.KeycloakBaseURL == "" {
		k.KeycloakBaseURL = defaultKeycloakBaseURL
	}
	if k.Realm == "" || k.ClientID == "" || k.ClientSecret == "" || k.UserName == "" || k.Password == "" {
		return fmt.Errorf("Realm or ClientID or ClientSecret or UserName or Password can't be empty")
	}

	return nil
}

func (k *KeycloakAuthenticator) GetBaseURL() string {
	if k.BaseURL != "" {
		return k.BaseURL
	}
	return defaultBaseURL
}

func (k *KeycloakAuthenticator) generateToken() error {
	if k.keycloakClient == nil {
		k.keycloakClient = gocloak.NewClient(k.KeycloakBaseURL, gocloak.SetAuthAdminRealms("admin/realms"), gocloak.SetAuthRealms("realms"))
		restyClient := k.keycloakClient.RestyClient()
		restyClient.SetDebug(k.Debug)
	}

	if k.token == nil {
		return k.generateAccessToken()
	}

	if isvalid(k.token.AccessToken) {
		return nil
	}

	if isvalid(k.token.RefreshToken) {
		token, err := k.keycloakClient.RefreshToken(context.Background(), k.token.RefreshToken, k.ClientID, k.ClientSecret, k.Realm)
		if err != nil {
			return err
		}

		k.token = token
		return nil
	}

	return k.generateAccessToken()
}

func (k *KeycloakAuthenticator) generateAccessToken() error {
	token, err := k.keycloakClient.Login(context.Background(), k.ClientID, k.ClientSecret, k.Realm, k.UserName, k.Password)
	if err != nil {
		return err
	}

	k.token = token
	return nil
}

func isvalid(token string) bool {
	claims, err := parseJWT(token)
	if err != nil {
		return false
	}
	// Invalidate the token before 60 seconds to avoid any failures during the calls.
	if time.Now().UTC().Unix() > (claims.ExpiresAt - 60) {
		return false
	}
	return true
}

type claim struct {
	ExpiresAt int64 `json:"exp,omitempty"`
	IssuedAt  int64 `json:"iat,omitempty"`
}

// parseJWT parses the specified JWT token string and returns an instance of the coreJWTClaims struct.
func parseJWT(tokenString string) (claims *claim, err error) {
	// A JWT consists of three .-separated segments
	segments := strings.Split(tokenString, ".")
	if len(segments) != 3 {
		err = fmt.Errorf("token contains an invalid number of segments")
		return
	}

	// Parse Claims segment.
	var claimBytes []byte
	claimBytes, err = decodeSegment(segments[1])
	if err != nil {
		err = fmt.Errorf("error decoding claims segment: %s", err.Error())
		return
	}

	// Now deserialize the claims segment into our coreClaims struct.
	claims = &claim{}
	err = json.Unmarshal(claimBytes, claims)
	if err != nil {
		err = fmt.Errorf("error unmarshalling token: %s", err.Error())
		return
	}

	return
}

// Decode JWT specific base64url encoding with padding stripped
// Copied from https://github.com/golang-jwt/jwt/blob/main/token.go
func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}
