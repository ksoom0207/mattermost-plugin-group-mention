package main

import (
	"fmt"
	"time"
)

// checkRateLimits checks if rate limits are exceeded
func (p *Plugin) checkRateLimits(userID, channelID, groupName string) bool {
	config := p.getConfiguration()

	// Check user rate limit
	userKey := fmt.Sprintf("user:%s", userID)
	if !p.checkAndIncrementRateLimit(userKey, config.PerUserPerMinute) {
		return false
	}

	// Check channel rate limit
	channelKey := fmt.Sprintf("channel:%s", channelID)
	if !p.checkAndIncrementRateLimit(channelKey, config.PerChannelPerMinute) {
		return false
	}

	// Check group rate limit
	groupKey := fmt.Sprintf("group:%s", groupName)
	if !p.checkAndIncrementRateLimit(groupKey, config.PerGroupPerMinute) {
		return false
	}

	return true
}

// checkAndIncrementRateLimit checks and increments a rate limit counter
func (p *Plugin) checkAndIncrementRateLimit(key string, limit int) bool {
	p.rateLimitLock.Lock()
	defer p.rateLimitLock.Unlock()

	now := time.Now()

	// Clean up expired entries
	if entry, exists := p.rateLimitCache[key]; exists {
		if now.After(entry.ExpiresAt) {
			delete(p.rateLimitCache, key)
		}
	}

	// Get or create entry
	entry, exists := p.rateLimitCache[key]
	if !exists {
		entry = &RateLimitEntry{
			Key:       key,
			Count:     0,
			ExpiresAt: now.Add(time.Minute),
		}
		p.rateLimitCache[key] = entry
	}

	// Check limit
	if entry.Count >= limit {
		return false
	}

	// Increment counter
	entry.Count++
	return true
}

// cleanupRateLimitCache periodically cleans up expired rate limit entries
func (p *Plugin) cleanupRateLimitCache() {
	p.rateLimitLock.Lock()
	defer p.rateLimitLock.Unlock()

	now := time.Now()
	for key, entry := range p.rateLimitCache {
		if now.After(entry.ExpiresAt) {
			delete(p.rateLimitCache, key)
		}
	}
}

// resetRateLimitForUser resets rate limits for a specific user (for testing/admin)
func (p *Plugin) resetRateLimitForUser(userID string) {
	p.rateLimitLock.Lock()
	defer p.rateLimitLock.Unlock()

	userKey := fmt.Sprintf("user:%s", userID)
	delete(p.rateLimitCache, userKey)
}
