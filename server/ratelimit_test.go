package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRateLimitEntry(t *testing.T) {
	entry := &RateLimitEntry{
		Key:       "test",
		Count:     5,
		ExpiresAt: time.Now().Add(time.Minute),
	}

	assert.Equal(t, "test", entry.Key)
	assert.Equal(t, 5, entry.Count)
	assert.True(t, entry.ExpiresAt.After(time.Now()))
}

func TestRateLimitExpiry(t *testing.T) {
	entry := &RateLimitEntry{
		Key:       "test",
		Count:     5,
		ExpiresAt: time.Now().Add(-time.Second), // Already expired
	}

	assert.True(t, time.Now().After(entry.ExpiresAt))
}
