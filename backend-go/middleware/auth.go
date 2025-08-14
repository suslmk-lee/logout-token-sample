package middleware

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// RequireAuth middleware ensures user is authenticated
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")

		log.Printf("requireAuth: session userID = %v", userID)
		log.Printf("requireAuth: session ID = %s", session.ID())

		if userID == nil {
			log.Printf("requireAuth: No user_id in session")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		log.Printf("requireAuth: Authenticated user: %s", userID)
		c.Set("user_id", userID)
		c.Next()
	}
}