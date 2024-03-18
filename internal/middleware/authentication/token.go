package authentication

import (
	"github.com/golang-jwt/jwt"
	"os"
	"time"
)

type TokenService interface {
	GenerateToken(userID string) (string, error)
}

type JWTTokenService struct{}

type CustomClaims struct {
	UserID string `json:"userId"`
	jwt.StandardClaims
}

func NewJWTTokenService() *JWTTokenService {
	return &JWTTokenService{}
}

// GenerateToken generates a new JWT token for a given user ID.
func (service *JWTTokenService) GenerateToken(userID string) (string, error) {
	jwtkey := os.Getenv("JWT_KEY")
	if jwtkey == "" {
		jwtkey = "quick_match"
	}
	var jk = []byte(jwtkey)
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &CustomClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jk)
	if err != nil {
		return "", err
	}

	return tokenString, err
}
