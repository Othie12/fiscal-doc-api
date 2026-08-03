package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/othie12/scanner-api/config"
	"golang.org/x/crypto/bcrypt"
)

/**########################## PASSWORD HASHING ###############################**/
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

/**########################## JWT TOKEN SIGNING AND VERIFICATION ###############################**/
var secretKey = []byte(config.ServerConfig.Secret)

type Claims struct {
	UserID    uint   `json:"user_id"`
	UserLevel string `json:"user_level"`
	jwt.RegisteredClaims
}

func CreateToken(userID uint, userLevel string) (string, error) {
	claims := Claims{
		UserID:    userID,
		UserLevel: userLevel,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // Token valid for 7 days
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// verifyToken parses and verifies the JWT token, returning the user ID or an error
func VerifyToken(tokenString string) (uint, string, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return 0, "", err
	}

	// Check if the token is valid
	if !token.Valid {
		return 0, "", fmt.Errorf("invalid token")
	}

	// Extract claims and userID
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return 0, "", fmt.Errorf("error extracting claims")
	}

	return claims.UserID, claims.UserLevel, nil
}
