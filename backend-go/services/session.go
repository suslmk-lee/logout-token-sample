package services

import (
	"log"
	"sync"
	"time"

	"keycloak-logout-backend-go/models"
)

// SessionService manages user sessions and SSE connections
type SessionService struct {
	activeSessions  map[string]*models.SessionData
	sessionsMutex   sync.RWMutex
	sseClients      map[string]*models.SSEClient
	sseClientsMutex sync.RWMutex
}

// NewSessionService creates a new session service
func NewSessionService() *SessionService {
	return &SessionService{
		activeSessions: make(map[string]*models.SessionData),
		sseClients:     make(map[string]*models.SSEClient),
	}
}

// AddSession adds a new user session
func (s *SessionService) AddSession(userID string, sessionData *models.SessionData) {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()
	s.activeSessions[userID] = sessionData
}

// GetSession retrieves a user session
func (s *SessionService) GetSession(userID string) (*models.SessionData, bool) {
	s.sessionsMutex.RLock()
	defer s.sessionsMutex.RUnlock()
	session, exists := s.activeSessions[userID]
	return session, exists
}

// RemoveSession removes a user session
func (s *SessionService) RemoveSession(userID string) {
	s.sessionsMutex.Lock()
	defer s.sessionsMutex.Unlock()
	delete(s.activeSessions, userID)
}

// GetAllSessions returns all active sessions
func (s *SessionService) GetAllSessions() []*models.SessionData {
	s.sessionsMutex.RLock()
	defer s.sessionsMutex.RUnlock()

	sessions := make([]*models.SessionData, 0, len(s.activeSessions))
	for _, session := range s.activeSessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// AddSSEClient adds a new SSE client
func (s *SessionService) AddSSEClient(userID string, client *models.SSEClient) {
	s.sseClientsMutex.Lock()
	defer s.sseClientsMutex.Unlock()
	s.sseClients[userID] = client
	log.Printf("➕ SSE client added for user: %s (total: %d)", userID, len(s.sseClients))
}

// RemoveSSEClient removes an SSE client
func (s *SessionService) RemoveSSEClient(userID string) {
	s.sseClientsMutex.Lock()
	defer s.sseClientsMutex.Unlock()
	if client, exists := s.sseClients[userID]; exists {
		close(client.Done)
		delete(s.sseClients, userID)
		log.Printf("➖ SSE client removed for user: %s (remaining: %d)", userID, len(s.sseClients))
	}
}

// NotifySessionInvalidated notifies SSE clients about session invalidation
func (s *SessionService) NotifySessionInvalidated(userID string) {
	s.sseClientsMutex.RLock()
	client, exists := s.sseClients[userID]
	s.sseClientsMutex.RUnlock()

	log.Printf("🔔 NotifySessionInvalidated called for user: %s", userID)
	log.Printf("📋 SSE client exists: %v", exists)

	if exists {
		log.Printf("📤 Sending session_invalidated message to SSE client: %s", userID)
		select {
		case client.C <- "session_invalidated":
			log.Printf("✅ Message sent successfully to SSE client: %s", userID)
		case <-time.After(1 * time.Second):
			log.Printf("⏰ SSE client not responding, removing: %s", userID)
			// Client not responding, remove it
			s.RemoveSSEClient(userID)
		}
	} else {
		log.Printf("❌ No SSE client found for user: %s", userID)
		// Show all current SSE clients for debugging
		s.sseClientsMutex.RLock()
		log.Printf("📊 Current SSE clients: %d", len(s.sseClients))
		for id := range s.sseClients {
			log.Printf("  - Client ID: %s", id)
		}
		s.sseClientsMutex.RUnlock()
	}
}
