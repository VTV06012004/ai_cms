package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"
)

var jwtSecret = []byte("super_secret_jwt_key_cms_2026")

type Claims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
}

// Mã hóa mật khẩu
func HashPassword(password string) (string, error) {
	const iterations = 600000
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := derivePasswordKey([]byte(password), salt, iterations)
	return "$pbkdf2-sha256$" + strconv.Itoa(iterations) + "$" +
		base64.RawStdEncoding.EncodeToString(salt) + "$" +
		base64.RawStdEncoding.EncodeToString(hash), nil
}

// Kiểm tra mật khẩu
func CheckPasswordHash(password, hash string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 5 || parts[1] != "pbkdf2-sha256" {
		return false
	}
	iterations, err := strconv.Atoi(parts[2])
	if err != nil || iterations < 1 || iterations > 10000000 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil || len(salt) == 0 {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(want) != sha256.Size {
		return false
	}
	got := derivePasswordKey([]byte(password), salt, iterations)
	return subtle.ConstantTimeCompare(got, want) == 1
}

func derivePasswordKey(password, salt []byte, iterations int) []byte {
	mac := hmac.New(sha256.New, password)
	block := append(append([]byte(nil), salt...), 0, 0, 0, 1)
	mac.Write(block)
	u := mac.Sum(nil)
	result := append([]byte(nil), u...)
	for i := 1; i < iterations; i++ {
		mac.Reset()
		mac.Write(u)
		u = mac.Sum(nil)
		for j := range result {
			result[j] ^= u[j]
		}
	}
	return result
}

// Tạo Token JWT
func GenerateToken(userID uint, email, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		ExpiresAt: expirationTime.Unix(),
	}
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	unsigned := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, jwtSecret)
	_, _ = mac.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// Giải mã & Kiểm tra Token JWT
func ValidateToken(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("Token không hợp lệ hoặc đã hết hạn")
	}
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errors.New("Token không hợp lệ hoặc đã hết hạn")
	}
	var header struct {
		Algorithm string `json:"alg"`
	}
	if json.Unmarshal(headerJSON, &header) != nil || header.Algorithm != "HS256" {
		return nil, errors.New("Token không hợp lệ hoặc đã hết hạn")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, errors.New("Token không hợp lệ hoặc đã hết hạn")
	}
	mac := hmac.New(sha256.New, jwtSecret)
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return nil, errors.New("Token không hợp lệ hoặc đã hết hạn")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errors.New("Token không hợp lệ hoặc đã hết hạn")
	}
	claims := &Claims{}
	if json.Unmarshal(payload, claims) != nil || claims.ExpiresAt == 0 || time.Now().Unix() >= claims.ExpiresAt {
		return nil, errors.New("Token không hợp lệ hoặc đã hết hạn")
	}
	return claims, nil
}
