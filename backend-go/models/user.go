package models

import "time"

// UserProfile represents user information from OIDC token
type UserProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	Name        struct {
		GivenName  string `json:"givenName"`
		FamilyName string `json:"familyName"`
	} `json:"name"`
	Emails []struct {
		Value string `json:"value"`
	} `json:"emails"`
}

// SessionData represents an active user session
type SessionData struct {
	SessionID string      `json:"sessionId"`
	User      UserProfile `json:"user"`
	LoginTime time.Time   `json:"loginTime"`
}

// SSEClient represents a Server-Sent Events client connection
type SSEClient struct {
	UserID string
	C      chan string
	Done   chan bool
}

// SessionStatus represents the current authentication status
type SessionStatus struct {
	Authenticated bool `json:"authenticated"`
	SessionActive bool `json:"sessionActive"`
}