package config

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

// RistrettoTokenBlacklist implements middleware.TokenBlacklist using dgraph-io/ristretto cache.
type RistrettoTokenBlacklist struct {
	cache *ristretto.Cache
}

// NewRistrettoTokenBlacklist creates a new TokenBlacklist instance.
func NewRistrettoTokenBlacklist(cache *ristretto.Cache) *RistrettoTokenBlacklist {
	return &RistrettoTokenBlacklist{cache: cache}
}

// IsBlacklisted checks if the given token is present in the cache.
func (b *RistrettoTokenBlacklist) IsBlacklisted(token string) bool {
	if b.cache == nil {
		return false
	}
	_, found := b.cache.Get(token)
	return found
}

// Blacklist adds a token to the cache with a specified expiration/TTL.
func (b *RistrettoTokenBlacklist) Blacklist(token string, ttl time.Duration) {
	if b.cache == nil {
		return
	}
	b.cache.SetWithTTL(token, true, 1, ttl)
}