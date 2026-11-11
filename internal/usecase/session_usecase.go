package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"api/internal/domain"
)

const RefreshTokenLifetime = 7 * 24 * time.Hour

type AccessTokenGenerator interface {
	GenerateToken(userID uint, email, role string) (string, error)
}

type SessionUseCase struct {
	users  domain.UserRepository
	tokens domain.RefreshTokenRepository
	signer AccessTokenGenerator
	now    func() time.Time
}

type RefreshedSession struct {
	AccessToken  string
	RefreshToken string
	User         domain.User
}

func NewSessionUseCase(users domain.UserRepository, tokens domain.RefreshTokenRepository, signer AccessTokenGenerator) *SessionUseCase {
	return &SessionUseCase{users: users, tokens: tokens, signer: signer, now: time.Now}
}

func (uc *SessionUseCase) Issue(ctx context.Context, userID uint) (string, error) {
	rawToken, tokenHash, err := generateRefreshToken()
	if err != nil {
		return "", err
	}
	if err := uc.tokens.Create(ctx, userID, tokenHash, uc.now().Add(RefreshTokenLifetime)); err != nil {
		return "", fmt.Errorf("không thể tạo refresh token: %w", err)
	}
	return rawToken, nil
}

func (uc *SessionUseCase) Refresh(ctx context.Context, rawToken string) (*RefreshedSession, error) {
	if rawToken == "" || len(rawToken) > 128 {
		return nil, ErrInvalidRefreshToken
	}
	now := uc.now()
	newToken, newHash, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("không thể tạo refresh token mới: %w", err)
	}
	userID, err := uc.tokens.Rotate(ctx, hashRefreshToken(rawToken), newHash, now, now.Add(RefreshTokenLifetime))
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("không thể xoay refresh token: %w", err)
	}

	user, err := uc.users.GetByID(ctx, userID)
	if err != nil {
		if revokeErr := uc.tokens.Revoke(ctx, newHash); revokeErr != nil {
			return nil, fmt.Errorf("không thể tải tài khoản khi làm mới phiên: %v; không thể thu hồi token mới: %w", err, revokeErr)
		}
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, fmt.Errorf("không thể tải tài khoản khi làm mới phiên: %w", err)
	}
	data := user.Data()
	accessToken, err := uc.signer.GenerateToken(data.ID, string(data.Email), string(data.Role))
	if err != nil {
		if revokeErr := uc.tokens.Revoke(ctx, newHash); revokeErr != nil {
			return nil, fmt.Errorf("không thể tạo access token mới: %v; không thể thu hồi token mới: %w", err, revokeErr)
		}
		return nil, fmt.Errorf("không thể tạo access token mới: %w", err)
	}
	return &RefreshedSession{
		AccessToken:  accessToken,
		RefreshToken: newToken,
		User:         *user,
	}, nil
}

func (uc *SessionUseCase) Revoke(ctx context.Context, rawToken string) error {
	if rawToken == "" || len(rawToken) > 128 {
		return nil
	}
	return uc.tokens.Revoke(ctx, hashRefreshToken(rawToken))
}

func generateRefreshToken() (string, string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", "", fmt.Errorf("không thể tạo token ngẫu nhiên: %w", err)
	}
	raw := base64.RawURLEncoding.EncodeToString(value)
	return raw, hashRefreshToken(raw), nil
}

func hashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
