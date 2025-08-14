package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"keycloak-logout-backend-go/config"
	"keycloak-logout-backend-go/handlers"
	"keycloak-logout-backend-go/middleware"
	"keycloak-logout-backend-go/services"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Printf("Starting server with config - Keycloak: %s, Realm: %s", cfg.KeycloakURL, cfg.KeycloakRealm)

	// Initialize services
	authService, err := services.NewAuthService(cfg)
	if err != nil {
		log.Fatal("Failed to initialize auth service: ", err)
	}

	sessionService := services.NewSessionService()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(cfg, authService, sessionService)
	apiHandler := handlers.NewAPIHandler(sessionService)

	// Setup Gin
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Session configuration
	store := cookie.NewStore([]byte(cfg.SessionSecret))
	
	// 환경에 따른 세션 설정
	var sameSiteMode http.SameSite
	if cfg.IsLocalDevelopment() {
		sameSiteMode = http.SameSiteLaxMode  // 로컬 환경 (HTTP)
	} else {
		sameSiteMode = http.SameSiteNoneMode // 프로덕션 환경 (HTTPS, Cross-site)
	}
	
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   24 * 60 * 60,    // 24시간
		HttpOnly: false,           // 디버깅을 위해 false (CORS 환경)
		Secure:   cfg.IsHTTPS(),  // HTTPS 환경에서는 true
		SameSite: sameSiteMode,   // 환경에 따라 다르게 설정
		Domain:   "",             // 도메인 지정하지 않음
	})
	r.Use(sessions.Sessions("keycloak-session", store))

	// Setup routes
	setupRoutes(r, authHandler, apiHandler)

	// Start server
	log.Printf("Go Backend server running on http://localhost:%s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}

func setupRoutes(r *gin.Engine, authHandler *handlers.AuthHandler, apiHandler *handlers.APIHandler) {
	// Authentication routes
	r.GET("/auth/login", authHandler.HandleLogin)
	r.GET("/auth/callback", authHandler.HandleCallback)
	r.GET("/auth/logout", authHandler.HandleLogout)
	r.POST("/auth/backchannel-logout", authHandler.HandleBackchannelLogout)

	// API routes (with authentication)
	api := r.Group("/api")
	api.Use(middleware.RequireAuth())
	{
		api.GET("/user", apiHandler.HandleGetUser)
		api.GET("/events", apiHandler.HandleSSE)
	}

	// Public API routes
	r.GET("/api/sessions", apiHandler.HandleGetSessions)
	r.GET("/api/session-status", apiHandler.HandleSessionStatus)

	// Test endpoint
	r.GET("/auth/backchannel-logout", authHandler.HandleBackchannelLogoutTest)
}