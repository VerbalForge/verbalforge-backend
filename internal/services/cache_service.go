package services

import (
	"fmt"
	"sync"
	"time"

	"verbalforge-backend/internal/models"
)

// CacheEntry represents a cached item with expiration
type CacheEntry struct {
	Items     []models.PracticeItem
	ExpiresAt time.Time
}

// CacheService provides in-memory caching for practice items
type CacheService struct {
	cache sync.Map
	ttl   time.Duration
	mu    sync.RWMutex
}

// NewCacheService creates a new cache service with 24-hour TTL
func NewCacheService() *CacheService {
	cs := &CacheService{
		ttl: 24 * time.Hour,
	}

	// Start background cleanup goroutine
	go cs.cleanupExpired()

	return cs
}

// GenerateCacheKey creates a cache key from filters
func (cs *CacheService) GenerateCacheKey(difficulty, itemType string) string {
	return fmt.Sprintf("practice:%s:%s", difficulty, itemType)
}

// Get retrieves items from cache if they exist and haven't expired
func (cs *CacheService) Get(key string) ([]models.PracticeItem, bool) {
	value, ok := cs.cache.Load(key)
	if !ok {
		return nil, false
	}

	entry, ok := value.(CacheEntry)
	if !ok {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		cs.cache.Delete(key)
		return nil, false
	}

	return entry.Items, true
}

// Set stores items in cache with TTL
func (cs *CacheService) Set(key string, items []models.PracticeItem) {
	entry := CacheEntry{
		Items:     items,
		ExpiresAt: time.Now().Add(cs.ttl),
	}
	cs.cache.Store(key, entry)
}

// Invalidate removes a specific cache entry
func (cs *CacheService) Invalidate(key string) {
	cs.cache.Delete(key)
}

// InvalidateAll clears the entire cache
func (cs *CacheService) InvalidateAll() {
	cs.cache.Range(func(key, value interface{}) bool {
		cs.cache.Delete(key)
		return true
	})
}

// InvalidateByPattern removes all cache entries matching a pattern
// For example, invalidate all "practice:medium:*" entries
func (cs *CacheService) InvalidateByPattern(pattern string) {
	cs.cache.Range(func(key, value interface{}) bool {
		keyStr, ok := key.(string)
		if !ok {
			return true
		}

		// Simple pattern matching (you can enhance this with regex if needed)
		if len(pattern) > 0 && len(keyStr) >= len(pattern) {
			if keyStr[:len(pattern)] == pattern {
				cs.cache.Delete(key)
			}
		}

		return true
	})
}

// cleanupExpired runs in the background and removes expired entries every hour
func (cs *CacheService) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		cs.cache.Range(func(key, value interface{}) bool {
			entry, ok := value.(CacheEntry)
			if !ok {
				return true
			}

			if now.After(entry.ExpiresAt) {
				cs.cache.Delete(key)
			}

			return true
		})
	}
}

// GetCacheStats returns cache statistics for monitoring
func (cs *CacheService) GetCacheStats() map[string]interface{} {
	var total, expired int
	now := time.Now()

	cs.cache.Range(func(key, value interface{}) bool {
		total++
		entry, ok := value.(CacheEntry)
		if ok && now.After(entry.ExpiresAt) {
			expired++
		}
		return true
	})

	return map[string]interface{}{
		"total_entries":   total,
		"expired_entries": expired,
		"ttl_hours":       cs.ttl.Hours(),
	}
}
