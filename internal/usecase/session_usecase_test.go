package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VTV06012004/ai_cms/internal/domain"
)

type refreshTokenRepositoryStub struct {
	tokens map[string]storedRefreshToken
}

type storedRefreshToken struct {
	userID    uint
	expiresAt time.Time
}

func newRefreshTokenRepositoryStub() *refreshTokenRepositoryStub {
	return &refreshTokenRepositoryStub{tokens: make(map[string]storedRefreshToken)}
}

func (r *refreshTokenRepositoryStub) Create(_ context.Context, userID uint, tokenHash string, expiresAt time.Time) error {
	r.tokens[tokenHash] = storedRefreshToken{userID: userID, expiresAt: expiresAt}
	return nil
}

func (r *refreshTokenRepositoryStub) Rotate(_ context.Context, oldHash, newHash string, now, expiresAt time.Time) (uint, error) {
	current, ok := r.tokens[oldHash]
	if !ok || !current.expiresAt.After(now) {
		return 0, ErrInvalidRefreshToken
	}
	delete(r.tokens, oldHash)
	r.tokens[newHash] = storedRefreshToken{userID: current.userID, expiresAt: expiresAt}
	return current.userID, nil
}

func (r *refreshTokenRepositoryStub) Revoke(_ context.Context, tokenHash string) error {
	delete(r.tokens, tokenHash)
	return nil
}

type accessTokenGeneratorStub struct {
	err error
}

func (s accessTokenGeneratorStub) GenerateToken(userID uint, email, role string) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return email + ":" + role, nil
}

func TestSessionUseCaseIssuesAndRotatesRefreshTokens(t *testing.T) {
	users := newUserRepositoryStub()
	user := domain.ReconstituteUser(domain.UserData{
		ID: 7, Name: "Author", Email: "author@example.com", Role: domain.UserRoleAuthor,
	})
	users.users[7] = user
	tokens := newRefreshTokenRepositoryStub()
	useCase := NewSessionUseCase(users, tokens, accessTokenGeneratorStub{})
	now := time.Date(2026, 11, 10, 10, 0, 0, 0, time.UTC)
	useCase.now = func() time.Time { return now }

	original, err := useCase.Issue(context.Background(), 7)
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if len(tokens.tokens) != 1 {
		t.Fatalf("stored token count = %d, want 1", len(tokens.tokens))
	}
	for hash := range tokens.tokens {
		if hash == original {
			t.Fatal("repository stored the raw refresh token")
		}
		if got := tokens.tokens[hash].expiresAt; !got.Equal(now.Add(RefreshTokenLifetime)) {
			t.Fatalf("expires at %v, want %v", got, now.Add(RefreshTokenLifetime))
		}
	}

	if err := user.ChangeRole(domain.UserRoleEditor); err != nil {
		t.Fatalf("ChangeRole returned error: %v", err)
	}
	refreshed, err := useCase.Refresh(context.Background(), original)
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if refreshed.AccessToken != "author@example.com:editor" {
		t.Fatalf("access token = %q, want current editor role", refreshed.AccessToken)
	}
	if refreshed.RefreshToken == original {
		t.Fatal("refresh token was not rotated")
	}
	if len(tokens.tokens) != 1 {
		t.Fatalf("stored token count after rotation = %d, want 1", len(tokens.tokens))
	}
	if _, err := useCase.Refresh(context.Background(), original); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("reusing rotated refresh token returned %v, want ErrInvalidRefreshToken", err)
	}
	if err := useCase.Revoke(context.Background(), refreshed.RefreshToken); err != nil {
		t.Fatalf("Revoke returned error: %v", err)
	}
	if len(tokens.tokens) != 0 {
		t.Fatalf("stored token count after revoke = %d, want 0", len(tokens.tokens))
	}
}

func TestSessionUseCaseRejectsExpiredAndMalformedRefreshTokens(t *testing.T) {
	tokens := newRefreshTokenRepositoryStub()
	users := newUserRepositoryStub()
	useCase := NewSessionUseCase(users, tokens, accessTokenGeneratorStub{})
	now := time.Now()
	useCase.now = func() time.Time { return now }
	tokens.tokens[hashRefreshToken("expired")] = storedRefreshToken{
		userID: 1, expiresAt: now,
	}

	for _, token := range []string{"expired", "", string(make([]byte, 129))} {
		if _, err := useCase.Refresh(context.Background(), token); !errors.Is(err, ErrInvalidRefreshToken) {
			t.Errorf("Refresh(%q) error = %v, want ErrInvalidRefreshToken", token, err)
		}
	}
}

func TestSessionUseCaseRevokesNewTokenWhenAccessTokenCannotBeGenerated(t *testing.T) {
	users := newUserRepositoryStub()
	users.users[1] = domain.ReconstituteUser(domain.UserData{
		ID: 1, Name: "Author", Email: "author@example.com", Role: domain.UserRoleAuthor,
	})
	tokens := newRefreshTokenRepositoryStub()
	useCase := NewSessionUseCase(users, tokens, accessTokenGeneratorStub{err: errors.New("signing failed")})
	now := time.Now()
	useCase.now = func() time.Time { return now }
	raw, err := useCase.Issue(context.Background(), 1)
	if err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if _, err := useCase.Refresh(context.Background(), raw); err == nil {
		t.Fatal("Refresh succeeded with a failing access token generator")
	}
	if len(tokens.tokens) != 0 {
		t.Fatalf("stored token count = %d, want no active session", len(tokens.tokens))
	}
}
