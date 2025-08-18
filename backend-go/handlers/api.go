package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"keycloak-logout-backend-go/models"
	"keycloak-logout-backend-go/services"
)

// APIHandler handles API requests
type APIHandler struct {
	sessionService *services.SessionService
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(sessionSvc *services.SessionService) *APIHandler {
	return &APIHandler{
		sessionService: sessionSvc,
	}
}

// HandleGetUser returns current user information
func (h *APIHandler) HandleGetUser(c *gin.Context) {
	userID := c.GetString("user_id")

	sessionData, exists := h.sessionService.GetSession(userID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})
		return
	}

	log.Printf("User profile: %+v", sessionData.User)

	profile := sessionData.User
	name := profile.DisplayName
	if name == "" && profile.Name.GivenName != "" {
		name = fmt.Sprintf("%s %s", profile.Name.GivenName, profile.Name.FamilyName)
	}
	if name == "" {
		name = profile.Username
	}

	email := "No email"
	if len(profile.Emails) > 0 {
		email = profile.Emails[0].Value
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    profile.ID,
			"name":  name,
			"email": email,
		},
	})
}

// HandleGetSessions returns all active sessions
func (h *APIHandler) HandleGetSessions(c *gin.Context) {
	sessions := h.sessionService.GetAllSessions()

	sessionList := make([]gin.H, 0, len(sessions))
	for _, session := range sessions {
		name := session.User.DisplayName
		if name == "" {
			name = session.User.Username
		}
		if name == "" {
			name = "Unknown"
		}

		sessionList = append(sessionList, gin.H{
			"sessionId": session.SessionID,
			"userId":    session.User.ID,
			"userName":  name,
			"loginTime": session.LoginTime.Format(time.RFC3339),
		})
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessionList})
}

// HandleSessionStatus returns current session status
func (h *APIHandler) HandleSessionStatus(c *gin.Context) {
	log.Println("Session status check called")

	session := sessions.Default(c)
	userID := session.Get("user_id")

	isAuthenticated := userID != nil
	log.Printf("Is authenticated: %v", isAuthenticated)

	var sessionActive bool
	if isAuthenticated {
		log.Printf("User ID: %s", userID)
		_, sessionActive = h.sessionService.GetSession(userID.(string))
		log.Printf("Session active: %v", sessionActive)
	}

	sessions := h.sessionService.GetAllSessions()
	log.Printf("Active sessions: %+v", sessions)

	status := models.SessionStatus{
		Authenticated: isAuthenticated,
		SessionActive: sessionActive,
	}

	c.JSON(http.StatusOK, status)
}

// HandleSSE handles Server-Sent Events connection
func (h *APIHandler) HandleSSE(c *gin.Context) {
	userID := c.GetString("user_id")

	log.Printf("SSE client connected: %s", userID)

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Credentials", "true")

	// Create SSE client
	client := &models.SSEClient{
		UserID: userID,
		C:      make(chan string, 10),
		Done:   make(chan bool),
	}

	h.sessionService.AddSSEClient(userID, client)
	defer h.sessionService.RemoveSSEClient(userID)

	// Send initial connection message
	if _, err := c.Writer.WriteString("data: connected\n\n"); err != nil {
		log.Printf("Error writing SSE initial message: %v", err)
		return
	}
	c.Writer.Flush()

	// Create keepalive ticker to prevent connection timeout (every 3 seconds)
	keepalive := time.NewTicker(3 * time.Second)
	defer keepalive.Stop()

	// Handle client messages
	for {
		select {
		case message := <-client.C:
			log.Printf("📨 Sending SSE message to %s: %s", userID, message)
			if _, err := c.Writer.WriteString(fmt.Sprintf("data: %s\n\n", message)); err != nil {
				log.Printf("Error writing SSE message: %v", err)
				return
			}
			c.Writer.Flush()
		case <-keepalive.C:
			// Send keepalive ping to prevent browser timeout
			if _, err := c.Writer.WriteString(": keepalive\n\n"); err != nil {
				log.Printf("Error writing SSE keepalive: %v", err)
				return
			}
			c.Writer.Flush()
		case <-client.Done:
			log.Printf("SSE client disconnected: %s", userID)
			return
		case <-c.Request.Context().Done():
			log.Printf("SSE client connection closed: %s", userID)
			return
		}
	}
}

