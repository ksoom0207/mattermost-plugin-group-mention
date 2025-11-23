package main

import (
	"time"
)

// Group represents a custom mention group
type Group struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	TeamID     string    `json:"team_id"`
	Visibility string    `json:"visibility"` // "public" or "private"
	Owners     []string  `json:"owners"`     // User IDs who can manage this group
	Members    []string  `json:"members"`    // User IDs who are part of this group
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// GroupIndex stores all group names for a team (for autocomplete)
type GroupIndex struct {
	TeamID     string   `json:"team_id"`
	GroupNames []string `json:"group_names"`
}

// RateLimitEntry tracks rate limiting
type RateLimitEntry struct {
	Key       string
	Count     int
	ExpiresAt time.Time
}

// Config holds plugin configuration
type Config struct {
	MaxExpandUsers                  int
	MaxMentionsPerMessage           int
	PerUserPerMinute                int
	PerChannelPerMinute             int
	PerGroupPerMinute               int
	AllowUserManagedGroups          bool
	DefaultGroupVisibility          string
	ExpandMode                      string
	LargeChannelMemberThreshold     int
	RequireChannelAdminInLargeChannels bool
	EnableDebugLogging              bool
}

// MentionContext holds context for processing a mention
type MentionContext struct {
	Post      interface{}
	Channel   interface{}
	User      interface{}
	TeamID    string
	ChannelID string
	UserID    string
}
