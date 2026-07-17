/*
 * Contract ID: CTR-006
 * Service Name: SupermarketService
 * Description: Ristretto In-Memory Cache Initialization with parameters per LLD Section 9.
 */

package config

import (
	"fmt"

	"github.com/dgraph-io/ristretto"
	"go.uber.org/zap"
)

// NewCache constructs and initializes a new thread-safe Ristretto cache instance.
// It applies the exact sizing parameters defined in the low-level design document:
// - NumCounters: 10,000 (tracks access frequency for 10,000 keys)
// - MaxCost: 100MB (limit for memory allocation)
// - BufferItems: 64 (concurrency channel buffer sizing)
func NewCache(cfg *Config, zapLog *zap.Logger) (*ristretto.Cache, error) {
	zapLog.Info("Initializing Ristretto in-memory cache...",
		zap.Int64("num_counters", cfg.CACHE_NUM_COUNTERS),
		zap.Int64("max_cost_bytes", cfg.CACHE_MAX_COST),
		zap.Int64("buffer_items", cfg.CACHE_BUFFER_ITEMS),
	)

	// Build dgraph-io/ristretto config
	cacheConfig := &ristretto.Config{
		NumCounters: cfg.CACHE_NUM_COUNTERS,
		MaxCost:     cfg.CACHE_MAX_COST,
		BufferItems: cfg.CACHE_BUFFER_ITEMS,
	}

	// Instantiate the cache
	cache, err := ristretto.NewCache(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate Ristretto cache: %w", err)
	}

	zapLog.Info("Ristretto in-memory cache initialized successfully")
	return cache, nil
}
