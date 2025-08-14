package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"keycloak-logout-backend-go/config"
	"keycloak-logout-backend-go/models"
	"keycloak-logout-backend-go/services"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	config         *config.Config
	authService    *services.AuthService
	sessionService *services.SessionService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config, authSvc *services.AuthService, sessionSvc *services.SessionService) *AuthHandler {
	return &AuthHandler{
		config:         cfg,
		authService:    authSvc,
		sessionService: sessionSvc,
	}
}

// HandleLogin initiates OAuth2 login flow
func (h *AuthHandler) HandleLogin(c *gin.Context) {
	session := sessions.Default(c)

	// Clear existing session
	session.Clear()

	state := h.authService.GenerateState()

	log.Printf("Login: session ID = %s", session.ID())
	log.Printf("Login: generated state = %s", state)

	session.Set("state", state)
	if err := session.Save(); err != nil {
		log.Printf("Login: session save error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session save failed"})
		return
	}

	log.Printf("Login: state saved to session")
	savedState := session.Get("state")
	log.Printf("Login: verified saved state = %v", savedState)

	authURL := h.authService.GetAuthURL(state)
	log.Printf("Login: redirecting to %s", authURL)

	c.Redirect(http.StatusFound, authURL)
}

// HandleCallback handles OAuth2 callback
func (h *AuthHandler) HandleCallback(c *gin.Context) {
	session := sessions.Default(c)
	storedState := session.Get("state")
	receivedState := c.Query("state")

	log.Printf("Callback: session ID = %s", session.ID())
	log.Printf("Callback: stored state = %v", storedState)
	log.Printf("Callback: received state = %s", receivedState)

	if storedState != receivedState {
		log.Printf("State mismatch: stored='%v', received='%s'", storedState, receivedState)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state"})
		return
	}

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code"})
		return
	}

	ctx := context.Background()
	token, err := h.authService.ExchangeCode(ctx, code)
	if err != nil {
		log.Printf("Token exchange error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token exchange failed"})
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Missing ID token"})
		return
	}

	claims, err := h.authService.VerifyIDToken(ctx, rawIDToken)
	if err != nil {
		log.Printf("ID token verification error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ID token verification failed"})
		return
	}

	// Extract user information
	userID := claims["sub"].(string)
	profile := h.authService.ExtractUserProfile(claims)

	// Create session data
	sessionID := uuid.New().String()
	sessionData := &models.SessionData{
		SessionID: sessionID,
		User:      profile,
		LoginTime: time.Now(),
	}

	// Store in active sessions
	h.sessionService.AddSession(userID, sessionData)

	// Save user ID to session
	session.Set("user_id", userID)
	if err := session.Save(); err != nil {
		log.Printf("Session save error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Session save failed"})
		return
	}

	log.Printf("User logged in: %s", userID)
	log.Printf("Session saved for user: %s", userID)

	c.Redirect(http.StatusFound, h.config.FrontendURL)
}

// HandleLogout handles user logout
func (h *AuthHandler) HandleLogout(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")

	if userID != nil {
		h.sessionService.RemoveSession(userID.(string))
		h.sessionService.RemoveSSEClient(userID.(string))
		log.Printf("User logged out: %s", userID)
	}

	session.Clear()
	session.Save()

	logoutURL := h.config.GetLogoutURL()
	c.JSON(http.StatusOK, gin.H{"logoutUrl": logoutURL})
}

// HandleBackchannelLogout handles Keycloak backchannel logout
func (h *AuthHandler) HandleBackchannelLogout(c *gin.Context) {
	log.Println("=== Backchannel Logout Called ===")
	log.Printf("Request body: %+v", c.Request.Form)
	log.Printf("Content-Type: %s", c.GetHeader("Content-Type"))

	logoutToken := c.PostForm("logout_token")
	if logoutToken == "" {
		log.Println("ERROR: Missing logout_token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing logout_token"})
		return
	}

	log.Printf("Logout token received: %s", logoutToken)

	claims, err := h.authService.ParseLogoutToken(logoutToken)
	if err != nil {
		log.Printf("ERROR: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid logout token"})
		return
	}

	log.Printf("Decoded payload: %+v", claims)

	userID, ok := claims["sub"].(string)
	if !ok {
		log.Println("ERROR: Missing 'sub' claim in logout token")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid logout token: missing sub"})
		return
	}

	log.Printf("Extracted user ID: %s", userID)

	// Check and remove active session
	sessions := h.sessionService.GetAllSessions()
	log.Printf("Active sessions before: %+v", sessions)

	if _, exists := h.sessionService.GetSession(userID); exists {
		h.sessionService.RemoveSession(userID)
		h.sessionService.NotifySessionInvalidated(userID)
		log.Printf("✅ User session invalidated: %s", userID)
	} else {
		log.Printf("⚠ User session not found in active sessions: %s", userID)
	}

	sessions = h.sessionService.GetAllSessions()
	log.Printf("Active sessions after: %+v", sessions)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// HandleBackchannelLogoutTest handles test endpoint
func (h *AuthHandler) HandleBackchannelLogoutTest(c *gin.Context) {
	log.Println("=== Backchannel Logout Test (GET) Called ===")
	c.JSON(http.StatusOK, gin.H{"message": "Backchannel logout endpoint is working"})
}