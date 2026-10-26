package security

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const legacyJWTSecret = "super_secret_jwt_key_cms_2026"

type Credentials struct {
	secret []byte
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewCredentials() *Credentials {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = legacyJWTSecret
	}
	return &Credentials{secret: []byte(secret)}
}

func (c *Credentials) HashPassword(password string) (string, error) {
	return HashPassword(password)
}

func (c *Credentials) CheckPasswordHash(password, encoded string) bool {
	return CheckPasswordHash(password, encoded)
}

func (c *Credentials) GenerateToken(userID uint, email, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(c.secret)
}

func (c *Credentials) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return c.secret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
