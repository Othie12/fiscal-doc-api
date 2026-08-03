package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/othie12/scanner-api/utils"
)

// Function to verify JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := extractTokenStringFromRequest(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ResponseWrapper(false, "Unauthorized access denied", nil))
			c.Abort()
			return
		}

		userID, userLevel, err := utils.VerifyToken(tokenString)
		if err != nil {
			log.Println("Veryfy token error: ", err.Error())
			c.JSON(http.StatusUnauthorized, utils.ResponseWrapper(false, "Unauthorized access denied", nil))
			c.Abort()
			return
		}
		log.Printf("RECOVERD USER ID: %d, ROLE: %s FROM TOKEN: ", userID, userLevel)

		// Attach user ID to the context
		c.Set("userID", userID)
		c.Set("userLevel", userLevel)
		c.Next()
	}
}

// Function to restrict unauthorized access to some resources
func UserLevelMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		userLevel := c.GetString("userLevel")

		if userLevel != "admin" {
			c.JSON(http.StatusUnauthorized, utils.ResponseWrapper(false, "You don't have permission to access this route", nil))
			c.Abort()
			return
		}
		c.Next()
	}
}

func extractTokenStringFromRequest(c *gin.Context) (string, bool) {
	tokenString := c.GetHeader("Authorization")
	log.Println("Tokenstring: ", tokenString)
	// check if tokenstring doesn't only contain "Bearer "
	if len(tokenString) > 7 {
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}
		log.Printf("Token from Header: %s", tokenString)
		return tokenString, true
	}

	var err error
	tokenString, err = c.Cookie("authToken")
	if err == nil {
		log.Printf("Token from cookie: %s", tokenString)
		return tokenString, true
	}

	log.Printf("Token from cookie: %s", tokenString)
	log.Println("Veryfy token error: ", err.Error())
	return err.Error(), false
}

func GetAuthUser(ctx *gin.Context) (uint, string) {
	userID := ctx.GetUint("userID")
	userLevel := ctx.GetString("userLevel")
	return userID, userLevel
}
