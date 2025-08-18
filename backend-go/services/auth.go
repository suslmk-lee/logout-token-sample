package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"

	"keycloak-logout-backend-go/config"
	"keycloak-logout-backend-go/models"
)

// AuthService handles OIDC authentication
type AuthService struct {
	config       *config.Config
	oauth2Config *oauth2.Config
	oidcVerifier *oidc.IDTokenVerifier
}

// NewAuthService creates a new authentication service
func NewAuthService(cfg *config.Config) (*AuthService, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(ctx, cfg.GetIssuerURL())
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC provider: %w", err)
	}

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.GetRedirectURL(),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	oidcVerifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})

	return &AuthService{
		config:       cfg,
		oauth2Config: oauth2Config,
		oidcVerifier: oidcVerifier,
	}, nil
}

// GenerateState generates a random state for OAuth2
func (a *AuthService) GenerateState() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Error generating random state: %v", err)
		// Fallback to timestamp-based state
		return base64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("state_%d", time.Now().UnixNano())))
	}
	return base64.URLEncoding.EncodeToString(b)
}

// GetAuthURL returns the OAuth2 authorization URL
func (a *AuthService) GetAuthURL(state string) string {
	return a.oauth2Config.AuthCodeURL(state)
}

// ExchangeCode exchanges authorization code for tokens
func (a *AuthService) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	return a.oauth2Config.Exchange(ctx, code)
}

// VerifyIDToken verifies and returns ID token claims
func (a *AuthService) VerifyIDToken(ctx context.Context, rawIDToken string) (map[string]interface{}, error) {
	idToken, err := a.oidcVerifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("ID token verification failed: %w", err)
	}

	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("claims extraction failed: %w", err)
	}

	return claims, nil
}

// ExtractUserProfile extracts user profile from claims
func (a *AuthService) ExtractUserProfile(claims map[string]interface{}) models.UserProfile {
	profile := models.UserProfile{
		ID: claims["sub"].(string),
	}

	if name, ok := claims["name"].(string); ok {
		profile.DisplayName = name
	}

	if username, ok := claims["preferred_username"].(string); ok {
		profile.Username = username
	}

	if givenName, ok := claims["given_name"].(string); ok {
		profile.Name.GivenName = givenName
	}

	if familyName, ok := claims["family_name"].(string); ok {
		profile.Name.FamilyName = familyName
	}

	if email, ok := claims["email"].(string); ok {
		profile.Emails = []struct {
			Value string `json:"value"`
		}{{Value: email}}
	}

	return profile
}

// ParseLogoutToken parses and validates logout token
func (a *AuthService) ParseLogoutToken(logoutToken string) (map[string]interface{}, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(logoutToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("invalid token format: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	return map[string]interface{}(claims), nil
}
