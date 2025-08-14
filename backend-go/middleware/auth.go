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
		
		// userID가 존재하는지, 그리고 string 타입이 맞는지 확인합니다.
		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			log.Printf("requireAuth: No valid user_id in session. userID: %v", userID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		log.Printf("requireAuth: Authenticated user: %s", userIDStr)
		c.Set("user_id", userIDStr)
		c.Next()
	}
}