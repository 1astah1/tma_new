package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type LoginAttempt struct {
	Count      int
	LastAttempt time.Time
	LockedUntil time.Time
}

type BruteForceProtector struct {
	attempts   map[string]*LoginAttempt
	mu         sync.Mutex
	maxAttempts int
	lockDuration time.Duration
	ttl        time.Duration
}

func NewBruteForceProtector(maxAttempts int, lockDuration time.Duration) *BruteForceProtector {
	bfp := &BruteForceProtector{
		attempts:     make(map[string]*LoginAttempt),
		maxAttempts:  maxAttempts,
		lockDuration: lockDuration,
		ttl:          1 * time.Hour,
	}

	go bfp.cleanup()

	return bfp
}

func (bfp *BruteForceProtector) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		bfp.mu.Lock()
		now := time.Now()
		for ip, attempt := range bfp.attempts {
			if now.Sub(attempt.LastAttempt) > bfp.ttl {
				delete(bfp.attempts, ip)
			}
		}
		bfp.mu.Unlock()
	}
}

func (bfp *BruteForceProtector) IsLocked(ip string) bool {
	bfp.mu.Lock()
	defer bfp.mu.Unlock()

	attempt, exists := bfp.attempts[ip]
	if !exists {
		return false
	}

	if time.Now().Before(attempt.LockedUntil) {
		return true
	}

	// Reset if lock period has expired
	if attempt.Count >= bfp.maxAttempts && time.Now().After(attempt.LockedUntil) {
		attempt.Count = 0
		attempt.LockedUntil = time.Time{}
	}

	return false
}

func (bfp *BruteForceProtector) RecordFailure(ip string) {
	bfp.mu.Lock()
	defer bfp.mu.Unlock()

	attempt, exists := bfp.attempts[ip]
	if !exists {
		attempt = &LoginAttempt{}
		bfp.attempts[ip] = attempt
	}

	attempt.Count++
	attempt.LastAttempt = time.Now()

	if attempt.Count >= bfp.maxAttempts {
		attempt.LockedUntil = time.Now().Add(bfp.lockDuration)
	}
}

func (bfp *BruteForceProtector) RecordSuccess(ip string) {
	bfp.mu.Lock()
	defer bfp.mu.Unlock()

	delete(bfp.attempts, ip)
}

func (bfp *BruteForceProtector) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		if bfp.IsLocked(ip) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"code":    "LOGIN_LOCKED",
					"message": "Too many failed login attempts. Please try again later.",
				},
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
