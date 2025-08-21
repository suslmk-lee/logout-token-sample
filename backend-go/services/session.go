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
	sseClients      map[string]map[string]*models.SSEClient // userID -> clientID -> SSEClient
	sseClientsMutex sync.RWMutex
}

// NewSessionService creates a new session service
func NewSessionService() *SessionService {
	return &SessionService{
		activeSessions: make(map[string]*models.SessionData),
		sseClients:     make(map[string]map[string]*models.SSEClient),
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

// AddSSEClient adds a new SSE client and returns the clientID
func (s *SessionService) AddSSEClient(userID string, client *models.SSEClient) string {
	s.sseClientsMutex.Lock()
	defer s.sseClientsMutex.Unlock()
	
	// 사용자별 클라이언트 맵이 없으면 생성
	if s.sseClients[userID] == nil {
		s.sseClients[userID] = make(map[string]*models.SSEClient)
	}
	
	// 클라이언트 ID로 고유하게 저장
	clientID := client.UserID + "-" + time.Now().Format("20060102150405")
	s.sseClients[userID][clientID] = client
	
	totalClients := 0
	for _, userClients := range s.sseClients {
		totalClients += len(userClients)
	}
	log.Printf("➕ SSE client added for user: %s (clientID: %s, total: %d)", userID, clientID, totalClients)
	return clientID
}

// RemoveSSEClient removes a specific SSE client
func (s *SessionService) RemoveSSEClient(userID string, clientID string) {
	s.sseClientsMutex.Lock()
	defer s.sseClientsMutex.Unlock()
	
	if userClients, exists := s.sseClients[userID]; exists {
		if client, exists := userClients[clientID]; exists {
			close(client.Done)
			delete(userClients, clientID)
			
			// 사용자의 모든 클라이언트가 제거되면 사용자 맵도 제거
			if len(userClients) == 0 {
				delete(s.sseClients, userID)
			}
			
			totalClients := 0
			for _, uc := range s.sseClients {
				totalClients += len(uc)
			}
			log.Printf("➖ SSE client removed for user: %s (clientID: %s, remaining: %d)", userID, clientID, totalClients)
		}
	}
}

// RemoveAllSSEClientsForUser removes all SSE clients for a user
func (s *SessionService) RemoveAllSSEClientsForUser(userID string) {
	s.sseClientsMutex.Lock()
	defer s.sseClientsMutex.Unlock()
	
	if userClients, exists := s.sseClients[userID]; exists {
		for clientID, client := range userClients {
			close(client.Done)
			log.Printf("➖ SSE client removed for user: %s (clientID: %s)", userID, clientID)
		}
		delete(s.sseClients, userID)
	}
}

// NotifySessionInvalidated notifies all SSE clients for a user about session invalidation
func (s *SessionService) NotifySessionInvalidated(userID string) {
	s.sseClientsMutex.RLock()
	userClients, exists := s.sseClients[userID]
	s.sseClientsMutex.RUnlock()

	log.Printf("🔔 NotifySessionInvalidated called for user: %s", userID)
	log.Printf("📋 SSE clients exists: %v", exists)

	if exists && len(userClients) > 0 {
		log.Printf("📤 Sending session_invalidated message to %d SSE clients for user: %s", len(userClients), userID)
		
		// 모든 클라이언트에게 메시지 전송
		var clientsToRemove []string
		for clientID, client := range userClients {
			select {
			case client.C <- "session_invalidated":
				log.Printf("✅ Message sent successfully to SSE client: %s (clientID: %s)", userID, clientID)
			case <-time.After(1 * time.Second):
				log.Printf("⏰ SSE client not responding, marking for removal: %s (clientID: %s)", userID, clientID)
				clientsToRemove = append(clientsToRemove, clientID)
			}
		}
		
		// 응답하지 않는 클라이언트들 제거
		for _, clientID := range clientsToRemove {
			s.RemoveSSEClient(userID, clientID)
		}
	} else {
		log.Printf("❌ No SSE clients found for user: %s", userID)
		// Show all current SSE clients for debugging
		s.sseClientsMutex.RLock()
		totalClients := 0
		for id, uc := range s.sseClients {
			clientCount := len(uc)
			totalClients += clientCount
			log.Printf("  - User ID: %s, Clients: %d", id, clientCount)
		}
		log.Printf("📊 Total SSE clients: %d", totalClients)
		s.sseClientsMutex.RUnlock()
	}
}
