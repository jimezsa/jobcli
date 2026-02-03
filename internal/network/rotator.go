package network

import (
	"errors"
	"net/url"
	"sync"
	"time"
)

var ErrNoProxies = errors.New("no proxies available")

type Rotator struct {
	proxies     []*url.URL
	banDuration time.Duration
	bannedUntil map[string]time.Time
	index       int
	mu          sync.Mutex
}

func NewRotator(raw []string, banDuration time.Duration) (*Rotator, error) {
	rotator := &Rotator{
		banDuration: banDuration,
		bannedUntil: map[string]time.Time{},
	}

	for _, proxy := range raw {
		u, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		rotator.proxies = append(rotator.proxies, u)
	}

	return rotator, nil
}

func (r *Rotator) Next() (*url.URL, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.proxies) == 0 {
		return nil, ErrNoProxies
	}

	start := r.index
	for {
		proxy := r.proxies[r.index]
		r.index = (r.index + 1) % len(r.proxies)

		if !r.isBanned(proxy) {
			return proxy, nil
		}

		if r.index == start {
			return nil, ErrNoProxies
		}
	}
}

func (r *Rotator) Report(proxy *url.URL, status int) {
	if proxy == nil {
		return
	}
	if status != 403 && status != 429 {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.bannedUntil[proxy.String()] = time.Now().Add(r.banDuration)
}

func (r *Rotator) isBanned(proxy *url.URL) bool {
	until, ok := r.bannedUntil[proxy.String()]
	if !ok {
		return false
	}
	if time.Now().After(until) {
		delete(r.bannedUntil, proxy.String())
		return false
	}
	return true
}
