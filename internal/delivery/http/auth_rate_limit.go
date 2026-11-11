package http

import (
	"strings"
	"sync"
	"time"
)

const (
	authRateLimit = 5
	// Làm mới phiên và đăng xuất xảy ra thường xuyên hơn đăng nhập (mỗi tab, mỗi
	// 15 phút) và nhiều người có thể dùng chung một IP, nên có ngưỡng riêng.
	sessionRateLimit = 60
	authRateWindow   = time.Minute
	maxRateLimitKeys = 100_000
	cleanupEveryHits = 256
)

type authRateLimitEntry struct {
	count   int
	resetAt time.Time
}

type authRateLimiter struct {
	limit    int
	mu       sync.Mutex
	entries  map[string]authRateLimitEntry
	requests uint64
}

func newAuthRateLimiter() *authRateLimiter {
	return newRateLimiter(authRateLimit)
}

func newRateLimiter(limit int) *authRateLimiter {
	return &authRateLimiter{limit: limit, entries: make(map[string]authRateLimitEntry)}
}

func (l *authRateLimiter) allow(ip, email string, now time.Time) bool {
	return l.allowKeys([]string{
		"ip:" + ip,
		"email:" + strings.ToLower(strings.TrimSpace(email)),
	}, now)
}

func (l *authRateLimiter) allowIP(ip string, now time.Time) bool {
	return l.allowKeys([]string{"ip:" + ip}, now)
}

func (l *authRateLimiter) allowEmail(email string, now time.Time) bool {
	return l.allowKeys([]string{"email:" + strings.ToLower(strings.TrimSpace(email))}, now)
}

func (l *authRateLimiter) allowKeys(keys []string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.requests++
	if l.requests%cleanupEveryHits == 0 {
		l.removeExpired(now)
	}

	newKeys := 0
	for _, key := range keys {
		entry, ok := l.entries[key]
		if !ok || !now.Before(entry.resetAt) {
			newKeys++
			continue
		}
		if entry.count >= l.limit {
			return false
		}
	}

	if len(l.entries)+newKeys > maxRateLimitKeys {
		l.removeExpired(now)
		if len(l.entries)+newKeys > maxRateLimitKeys {
			return false
		}
	}

	for _, key := range keys {
		entry := l.entries[key]
		if !now.Before(entry.resetAt) {
			entry = authRateLimitEntry{resetAt: now.Add(authRateWindow)}
		}
		entry.count++
		l.entries[key] = entry
	}
	return true
}

func (l *authRateLimiter) removeExpired(now time.Time) {
	for key, entry := range l.entries {
		if !now.Before(entry.resetAt) {
			delete(l.entries, key)
		}
	}
}
