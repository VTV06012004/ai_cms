package security

import (
	"errors"
	"testing"
	"time"
)

func TestNewCredentialsRequiresJWTSecret(t *testing.T) {
	for _, secret := range []string{"", " ", "\t"} {
		if _, err := NewCredentials(secret); !errors.Is(err, ErrMissingJWTSecret) {
			t.Errorf("NewCredentials(%q) error = %v, want ErrMissingJWTSecret", secret, err)
		}
	}
}

func TestGeneratedTokenExpiresAfterFifteenMinutes(t *testing.T) {
	credentials, err := NewCredentials("test-secret")
	if err != nil {
		t.Fatalf("NewCredentials: %v", err)
	}

	token, err := credentials.GenerateToken(1, "author@example.com", "author")
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	claims, err := credentials.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining < 14*time.Minute+59*time.Second || remaining > tokenLifetime {
		t.Fatalf("token lifetime remaining = %s, want approximately 15 minutes", remaining)
	}
}
