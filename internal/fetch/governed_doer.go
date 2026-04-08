package fetch

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GovernanceConfig configures bounded 429 cooldown behavior for an HTTP transport.
type GovernanceConfig struct {
	CooldownBase      time.Duration
	CooldownMax       time.Duration
	RespectRetryAfter bool
}

// RateLimitCooldownError reports that the transport is currently cooling down after a 429.
type RateLimitCooldownError struct {
	Until     time.Time
	Remaining time.Duration
}

func (e *RateLimitCooldownError) Error() string {
	if e == nil {
		return "search transport cooldown active"
	}
	return fmt.Sprintf("search transport cooling down for %s after upstream rate limiting", e.Remaining.Round(time.Second))
}

// GovernedDoer applies bounded, fail-fast rate-limit governance on top of another Doer.
type GovernedDoer struct {
	next Doer
	cfg  GovernanceConfig
	now  func() time.Time

	mu             sync.Mutex
	consecutive429 int
	cooldownUntil  time.Time
}

// NewGovernedDoer wraps a Doer with bounded 429 governance.
func NewGovernedDoer(next Doer, cfg GovernanceConfig) *GovernedDoer {
	base := cfg.CooldownBase
	if base <= 0 {
		base = 5 * time.Second
	}
	maxCooldown := cfg.CooldownMax
	if maxCooldown <= 0 || maxCooldown < base {
		maxCooldown = base
	}
	return &GovernedDoer{
		next: resolveClient(next),
		cfg: GovernanceConfig{
			CooldownBase:      base,
			CooldownMax:       maxCooldown,
			RespectRetryAfter: cfg.RespectRetryAfter,
		},
		now: func() time.Time { return time.Now().UTC() },
	}
}

// Do executes the request or fails fast when an active cooldown exists.
func (g *GovernedDoer) Do(req *http.Request) (*http.Response, error) {
	if g == nil {
		return resolveClient(nil).Do(req)
	}
	if err := g.cooldownError(); err != nil {
		return nil, err
	}

	resp, err := g.next.Do(req)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		g.resetCooldown()
		return nil, nil
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		g.resetCooldown()
		return resp, nil
	}

	g.enterCooldown(resp)
	return resp, nil
}

func (g *GovernedDoer) cooldownError() error {
	now := g.now()
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.cooldownUntil.IsZero() || !now.Before(g.cooldownUntil) {
		return nil
	}
	return &RateLimitCooldownError{Until: g.cooldownUntil, Remaining: g.cooldownUntil.Sub(now)}
}

func (g *GovernedDoer) resetCooldown() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.consecutive429 = 0
	g.cooldownUntil = time.Time{}
}

func (g *GovernedDoer) enterCooldown(resp *http.Response) {
	now := g.now()
	g.mu.Lock()
	defer g.mu.Unlock()

	g.consecutive429++
	cooldown := g.backoffDuration(now, resp)
	g.cooldownUntil = now.Add(cooldown)
}

func (g *GovernedDoer) backoffDuration(now time.Time, resp *http.Response) time.Duration {
	if g.cfg.RespectRetryAfter {
		if retryAfter, ok := parseRetryAfter(now, resp); ok {
			return clampDuration(retryAfter, g.cfg.CooldownBase, g.cfg.CooldownMax)
		}
	}
	shift := g.consecutive429 - 1
	if shift < 0 {
		shift = 0
	}
	backoff := g.cfg.CooldownBase << shift
	return clampDuration(backoff, g.cfg.CooldownBase, g.cfg.CooldownMax)
}

func parseRetryAfter(now time.Time, resp *http.Response) (time.Duration, bool) {
	if resp == nil {
		return 0, false
	}
	raw := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if raw == "" {
		return 0, false
	}
	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second, true
	}
	if when, err := http.ParseTime(raw); err == nil {
		return when.Sub(now), true
	}
	return 0, false
}

func clampDuration(value, minimum, maximum time.Duration) time.Duration {
	if value < minimum {
		return minimum
	}
	if value > maximum {
		return maximum
	}
	return value
}
