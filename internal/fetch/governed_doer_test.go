package fetch

import (
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

const firstDoErrFmt = "first Do() error = %v"

func TestGovernedDoerEntersCooldownOnFirst429AndFailsFast(t *testing.T) {
	now := time.Date(2026, time.April, 8, 19, 0, 0, 0, time.UTC)
	upstream := &countingDoer{response: httpResponse(http.StatusTooManyRequests, "limited")}
	governed := NewGovernedDoer(upstream, GovernanceConfig{CooldownBase: 5 * time.Second, CooldownMax: 30 * time.Second})
	governed.now = func() time.Time { return now }

	firstResp, err := governed.Do(mustRequest(t))
	if err != nil {
		t.Fatalf(firstDoErrFmt, err)
	}
	defer closeResponse(firstResp)
	if firstResp == nil || firstResp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("first Do() status = %v, want 429", statusCode(firstResp))
	}
	if upstream.calls != 1 {
		t.Fatalf("upstream calls after first 429 = %d, want 1", upstream.calls)
	}

	secondResp, err := governed.Do(mustRequest(t))
	var cooldownErr *RateLimitCooldownError
	if !errors.As(err, &cooldownErr) {
		t.Fatalf("second Do() error = %T, want *RateLimitCooldownError", err)
	}
	defer closeResponse(secondResp)
	if upstream.calls != 1 {
		t.Fatalf("upstream calls during cooldown = %d, want 1", upstream.calls)
	}
}

func TestGovernedDoerUsesRetryAfterBeforeFallbackBackoff(t *testing.T) {
	now := time.Date(2026, time.April, 8, 19, 0, 0, 0, time.UTC)
	upstream := &countingDoer{response: func(req *http.Request) (*http.Response, error) {
		resp, err := httpResponse(http.StatusTooManyRequests, "limited")(req)
		if err != nil {
			return nil, err
		}
		if resp.Header == nil {
			resp.Header = make(http.Header)
		}
		resp.Header.Set("Retry-After", "17")
		return resp, nil
	}}
	governed := NewGovernedDoer(upstream, GovernanceConfig{CooldownBase: 5 * time.Second, CooldownMax: 30 * time.Second, RespectRetryAfter: true})
	governed.now = func() time.Time { return now }

	resp, err := governed.Do(mustRequest(t))
	if err != nil {
		t.Fatalf(firstDoErrFmt, err)
	}
	defer closeResponse(resp)
	now = now.Add(10 * time.Second)
	secondResp, err := governed.Do(mustRequest(t))
	defer closeResponse(secondResp)
	var cooldownErr *RateLimitCooldownError
	if !errors.As(err, &cooldownErr) {
		t.Fatalf("second Do() error = %T, want *RateLimitCooldownError", err)
	}
	if got := cooldownErr.Remaining.Round(time.Second); got != 7*time.Second {
		t.Fatalf("cooldown remaining = %s, want 7s", got)
	}
}

func TestGovernedDoerFallsBackToBoundedExponentialBackoff(t *testing.T) {
	now := time.Date(2026, time.April, 8, 19, 0, 0, 0, time.UTC)
	upstream := &countingDoer{response: httpResponse(http.StatusTooManyRequests, "limited")}
	governed := NewGovernedDoer(upstream, GovernanceConfig{CooldownBase: 5 * time.Second, CooldownMax: 12 * time.Second})
	governed.now = func() time.Time { return now }

	firstResp, err := governed.Do(mustRequest(t))
	defer closeResponse(firstResp)
	if err != nil {
		t.Fatalf(firstDoErrFmt, err)
	}
	now = now.Add(6 * time.Second)
	secondResp, err := governed.Do(mustRequest(t))
	defer closeResponse(secondResp)
	if err != nil {
		t.Fatalf("second upstream Do() error = %v", err)
	}
	now = now.Add(9 * time.Second)
	thirdResp, err := governed.Do(mustRequest(t))
	defer closeResponse(thirdResp)
	var cooldownErr *RateLimitCooldownError
	if !errors.As(err, &cooldownErr) {
		t.Fatalf("cooldown error = %T, want *RateLimitCooldownError", err)
	}
	if got := cooldownErr.Remaining.Round(time.Second); got != time.Second {
		t.Fatalf("cooldown remaining = %s, want 1s", got)
	}
}

func TestGovernedDoerLeavesPlainTransportErrorsUngoverned(t *testing.T) {
	upstream := &countingDoer{response: func(*http.Request) (*http.Response, error) {
		return nil, errors.New("dial tcp refused")
	}}
	governed := NewGovernedDoer(upstream, GovernanceConfig{CooldownBase: 5 * time.Second, CooldownMax: 30 * time.Second})

	for range 2 {
		resp, err := governed.Do(mustRequest(t))
		closeResponse(resp)
		if err == nil {
			t.Fatal("Do() error = nil, want transport error")
		}
	}
	if upstream.calls != 2 {
		t.Fatalf("upstream calls = %d, want 2 without cooldown", upstream.calls)
	}
}

type countingDoer struct {
	response func(*http.Request) (*http.Response, error)
	calls    int
}

func (d *countingDoer) Do(req *http.Request) (*http.Response, error) {
	d.calls++
	return d.response(req)
}

func mustRequest(t *testing.T) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "https://example.invalid/search/node?keys=tilde", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}
	return req
}

func statusCode(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	defer closeResponse(resp)
	_, _ = io.ReadAll(resp.Body)
	return resp.StatusCode
}

func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}
