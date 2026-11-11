package persistence

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/VTV06012004/ai_cms/internal/usecase"
)

func TestRefreshTokenRotationIsSingleUse(t *testing.T) {
	db, userID, _ := newPostSlugMySQLTestDB(t)
	repository := NewRefreshTokenRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	oldHash := strings.Repeat("a", 64)
	newHashes := []string{strings.Repeat("b", 64), strings.Repeat("c", 64)}
	if err := repository.Create(ctx, userID, oldHash, now.Add(time.Hour)); err != nil {
		t.Fatalf("create refresh token: %v", err)
	}

	type result struct {
		err error
	}
	results := make(chan result, len(newHashes))
	var workers sync.WaitGroup
	for _, newHash := range newHashes {
		workers.Add(1)
		go func(newHash string) {
			defer workers.Done()
			_, err := repository.Rotate(ctx, oldHash, newHash, now, now.Add(2*time.Hour))
			results <- result{err: err}
		}(newHash)
	}
	workers.Wait()
	close(results)

	successes, rejected := 0, 0
	for result := range results {
		switch {
		case result.err == nil:
			successes++
		case errors.Is(result.err, usecase.ErrInvalidRefreshToken):
			rejected++
		default:
			t.Fatalf("rotate returned unexpected error: %v", result.err)
		}
	}
	if successes != 1 || rejected != 1 {
		t.Fatalf("successful rotations = %d, rejected replays = %d; want 1 each", successes, rejected)
	}

	var activeTokens int64
	if err := db.Model(&RefreshTokenModel{}).Count(&activeTokens).Error; err != nil {
		t.Fatalf("count active refresh tokens: %v", err)
	}
	if activeTokens != 1 {
		t.Fatalf("active token count = %d, want 1", activeTokens)
	}
}

func TestRefreshTokenCreateRemovesExpiredTokensOfTheSameUser(t *testing.T) {
	db, userID, _ := newPostSlugMySQLTestDB(t)
	repository := NewRefreshTokenRepository(db)
	ctx := context.Background()
	now := time.Now().UTC()
	expired := strings.Repeat("d", 64)
	active := strings.Repeat("e", 64)
	fresh := strings.Repeat("f", 64)
	if err := repository.Create(ctx, userID, expired, now.Add(-time.Hour)); err != nil {
		t.Fatalf("create expired token: %v", err)
	}
	if err := repository.Create(ctx, userID, active, now.Add(time.Hour)); err != nil {
		t.Fatalf("create active token: %v", err)
	}
	if err := repository.Create(ctx, userID, fresh, now.Add(2*time.Hour)); err != nil {
		t.Fatalf("create fresh token: %v", err)
	}

	var hashes []string
	if err := db.Model(&RefreshTokenModel{}).Where("user_id = ?", userID).Pluck("token_hash", &hashes).Error; err != nil {
		t.Fatalf("list tokens: %v", err)
	}
	if len(hashes) != 2 {
		t.Fatalf("tokens after cleanup = %v, want only the active and fresh ones", hashes)
	}
	for _, hash := range hashes {
		if hash == expired {
			t.Fatal("expired token was not removed")
		}
	}
}
