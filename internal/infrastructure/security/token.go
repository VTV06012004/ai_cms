package security

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenLifetime = 15 * time.Minute

var ErrMissingJWTSecret = errors.New("JWT_SECRET is required")

type Credentials struct {
	secret []byte
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewCredentials(secret string) (*Credentials, error) {
	if strings.TrimSpace(secret) == "" {
		return nil, ErrMissingJWTSecret
	}
	return &Credentials{secret: []byte(secret)}, nil
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
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenLifetime)),
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
