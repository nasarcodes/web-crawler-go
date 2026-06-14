package ratelimiter

import (
	"sync"
	"time"
)

type Limiter struct {
	lastRequest map[string]time.Time
	mu          sync.Mutex
	interval    time.Duration
}

func New(interval time.Duration) *Limiter {
	return &Limiter{
		lastRequest: make(map[string]time.Time),
		interval:    interval,
	}
}

func (l *Limiter) Wait(host string) {
	l.mu.Lock()

	now := time.Now()

	nextAllowed := l.lastRequest[host]

	if nextAllowed.IsZero() {
		nextAllowed = now
	}

	if nextAllowed.Before(now) {
		nextAllowed = now
	}

	waitTime := time.Duration(0)

	if nextAllowed.After(now) {
		waitTime = nextAllowed.Sub(now)
	}

	l.lastRequest[host] = nextAllowed.Add(l.interval)

	l.mu.Unlock()

	if waitTime > 0 {
		time.Sleep(waitTime)
	}
}
