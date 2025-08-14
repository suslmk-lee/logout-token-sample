package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	KeycloakURL    string
	KeycloakRealm  string
	ClientID       string
	ClientSecret   string
	SessionSecret  string
	Port           string
	FrontendURL    string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Try to load .env file (optional)
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	config := &Config{
		KeycloakURL:   getEnv("KEYCLOAK_URL", "https://registry-keycloak.k-paas.org"),
		KeycloakRealm: getEnv("KEYCLOAK_REALM", "cp-realm"),
		ClientID:      getEnv("CLIENT_ID", "cp-client"),
		ClientSecret:  getEnv("CLIENT_SECRET", ""),
		SessionSecret: getEnv("SESSION_SECRET", "default-session-secret-for-development"),
		Port:          getEnv("PORT", "3001"),
		FrontendURL:   getEnv("FRONTEND_URL", "http://localhost:3000"),
	}

	// Validate required fields
	if config.ClientSecret == "" {
		log.Fatal("CLIENT_SECRET environment variable is required")
	}

	return config
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// IsHTTPS checks if the frontend URL uses HTTPS
func (c *Config) IsHTTPS() bool {
	return len(c.FrontendURL) >= 8 && c.FrontendURL[:8] == "https://"
}

// GetIssuerURL returns the Keycloak issuer URL
func (c *Config) GetIssuerURL() string {
	return c.KeycloakURL + "/realms/" + c.KeycloakRealm
}

// GetLogoutURL returns the Keycloak logout URL
func (c *Config) GetLogoutURL() string {
	return c.KeycloakURL + "/realms/" + c.KeycloakRealm + "/protocol/openid-connect/logout?redirect_uri=" + c.FrontendURL
}

// IsLocalDevelopment checks if running in local development environment
func (c *Config) IsLocalDevelopment() bool {
	return c.FrontendURL == "http://localhost:3000"
}

// GetRedirectURL returns the appropriate redirect URL based on environment
func (c *Config) GetRedirectURL() string {
	if c.IsLocalDevelopment() {
		return "http://localhost:" + c.Port + "/auth/callback"
	}
	return c.FrontendURL + "/auth/callback"
}