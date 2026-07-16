package resilient

import (
	"alongtu/backend/internal/navigation"
	"alongtu/backend/internal/places"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"sync"
	"time"
)

type Upstream interface {
	places.Provider
	navigation.Provider
}
type Metrics interface {
	ObserveProvider(operation, outcome string, duration time.Duration)
	ObserveProviderCache(operation, state string)
	SetProviderCircuit(state string)
}
type Provider struct {
	up        Upstream
	redis     *redis.Client
	mu        sync.Mutex
	failures  int
	openUntil time.Time
	halfOpen  bool
	metrics   Metrics
	timeout   time.Duration
}
type cacheEntry struct {
	FreshUntil time.Time       `json:"freshUntil"`
	Value      json.RawMessage `json:"value"`
}

var ErrCircuitOpen = errors.New("provider circuit open")

const defaultTotalTimeout = 2200 * time.Millisecond

func New(up Upstream, r *redis.Client, metrics ...Metrics) *Provider {
	return NewWithTimeout(up, r, defaultTotalTimeout, metrics...)
}
func NewWithTimeout(up Upstream, r *redis.Client, timeout time.Duration, metrics ...Metrics) *Provider {
	if timeout <= 0 {
		timeout = defaultTotalTimeout
	}
	provider := &Provider{up: up, redis: r, timeout: timeout}
	if len(metrics) > 0 {
		provider.metrics = metrics[0]
		provider.metrics.SetProviderCircuit("closed")
	}
	return provider
}
func (p *Provider) Nearby(ctx context.Context, q places.NearbyQuery) (places.Page, error) {
	key := key("nearby", q)
	var cached places.Page
	fresh, stale := p.read(ctx, "nearby", key, &cached)
	if fresh {
		cached.DataFreshness = "cached"
		return cached, nil
	}
	var out places.Page
	err := p.call(ctx, "nearby", func(c context.Context) error { var e error; out, e = p.up.Nearby(c, q); return e })
	if err == nil {
		p.write(ctx, key, out, 2*time.Minute, 30*time.Minute)
		return out, nil
	}
	if stale {
		cached.Degraded = true
		cached.DataFreshness = "degraded"
		return cached, nil
	}
	return places.Page{}, err
}
func (p *Provider) Coverage(category places.Category) string {
	if provider, ok := p.up.(interface{ Coverage(places.Category) string }); ok && provider.Coverage(category) == "not_covered" {
		return "not_covered"
	}
	return "covered"
}
func (p *Provider) Search(ctx context.Context, q places.SearchQuery) (places.Page, error) {
	key := key("search", q)
	var cached places.Page
	fresh, stale := p.read(ctx, "search", key, &cached)
	if fresh {
		cached.DataFreshness = "cached"
		return cached, nil
	}
	var out places.Page
	err := p.call(ctx, "search", func(c context.Context) error { var e error; out, e = p.up.Search(c, q); return e })
	if err == nil {
		p.write(ctx, key, out, 2*time.Minute, 30*time.Minute)
		return out, nil
	}
	if stale {
		cached.Degraded = true
		cached.DataFreshness = "degraded"
		return cached, nil
	}
	return places.Page{}, err
}
func (p *Provider) Route(ctx context.Context, q navigation.RouteRequest) (navigation.Route, error) {
	key := key("route", q)
	var cached navigation.Route
	fresh, _ := p.read(ctx, "route", key, &cached)
	if fresh {
		cached.DataFreshness = "cached"
		return cached, nil
	}
	var out navigation.Route
	err := p.call(ctx, "route", func(c context.Context) error { var e error; out, e = p.up.Route(c, q); return e })
	if err == nil {
		p.write(ctx, key, out, time.Minute, time.Minute)
		return out, nil
	}
	return navigation.Route{}, err
}
func (p *Provider) call(ctx context.Context, operation string, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	started := time.Now()
	if !p.allow() {
		p.observeProvider(operation, "circuit_open", started)
		return ErrCircuitOpen
	}
	err := fn(ctx)
	if err != nil && temporary(err) {
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case <-time.After(100 * time.Millisecond):
			err = fn(ctx)
		}
	}
	p.record(err == nil || !retryable(err))
	outcome := "success"
	if err != nil {
		outcome = "error"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			outcome = "timeout"
		}
	}
	p.observeProvider(operation, outcome, started)
	return err
}
func temporary(err error) bool {
	type temporaryError interface{ Temporary() bool }
	var target temporaryError
	return errors.As(err, &target) && target.Temporary()
}

func retryable(err error) bool {
	return temporary(err) || errors.Is(err, context.DeadlineExceeded)
}
func (p *Provider) allow() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	if now.Before(p.openUntil) {
		p.setCircuit("open")
		return false
	}
	if !p.openUntil.IsZero() {
		if p.halfOpen {
			return false
		}
		p.halfOpen = true
		p.setCircuit("half_open")
	}
	return true
}
func (p *Provider) record(success bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.halfOpen = false
	if success {
		p.failures = 0
		p.openUntil = time.Time{}
		p.setCircuit("closed")
		return
	}
	p.failures++
	if p.failures >= 5 {
		p.openUntil = time.Now().Add(30 * time.Second)
		p.setCircuit("open")
	}
}
func (p *Provider) read(ctx context.Context, operation, k string, dst any) (bool, bool) {
	if p.redis == nil {
		p.observeCache(operation, "miss")
		return false, false
	}
	raw, err := p.redis.Get(ctx, k).Bytes()
	if err != nil {
		p.observeCache(operation, "miss")
		return false, false
	}
	var e cacheEntry
	if json.Unmarshal(raw, &e) != nil || json.Unmarshal(e.Value, dst) != nil {
		p.observeCache(operation, "invalid")
		return false, false
	}
	if time.Now().Before(e.FreshUntil) {
		p.observeCache(operation, "fresh")
		return true, true
	}
	p.observeCache(operation, "stale")
	return false, true
}
func (p *Provider) write(ctx context.Context, k string, v any, fresh, ttl time.Duration) {
	if p.redis == nil {
		return
	}
	value, _ := json.Marshal(v)
	raw, _ := json.Marshal(cacheEntry{FreshUntil: time.Now().Add(fresh), Value: value})
	_ = p.redis.Set(ctx, k, raw, ttl).Err()
}
func (p *Provider) observeProvider(operation, outcome string, started time.Time) {
	if p.metrics != nil {
		p.metrics.ObserveProvider(operation, outcome, time.Since(started))
	}
}
func (p *Provider) observeCache(operation, state string) {
	if p.metrics != nil {
		p.metrics.ObserveProviderCache(operation, state)
	}
}
func (p *Provider) setCircuit(state string) {
	if p.metrics != nil {
		p.metrics.SetProviderCircuit(state)
	}
}
func key(kind string, v any) string {
	raw, _ := json.Marshal(v)
	sum := sha256.Sum256(raw)
	return "provider:" + kind + ":" + base64.RawURLEncoding.EncodeToString(sum[:])
}
