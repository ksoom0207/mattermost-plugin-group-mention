package main

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMentionRegex(t *testing.T) {
	regex := regexp.MustCompile(`(?:^|\s)@([a-zA-Z0-9_\-.]{2,64})\b`)

	tests := []struct {
		name     string
		message  string
		expected []string
	}{
		{
			name:     "Single mention",
			message:  "Hello @dev team",
			expected: []string{"dev"},
		},
		{
			name:     "Multiple mentions",
			message:  "CC @dev @ops @qa",
			expected: []string{"dev", "ops", "qa"},
		},
		{
			name:     "Mention at start",
			message:  "@dev please review",
			expected: []string{"dev"},
		},
		{
			name:     "No mention",
			message:  "Hello team",
			expected: []string{},
		},
		{
			name:     "Mention with underscore",
			message:  "Hello @dev_team",
			expected: []string{"dev_team"},
		},
		{
			name:     "Mention with dash",
			message:  "Hello @dev-team",
			expected: []string{"dev-team"},
		},
		{
			name:     "Too short mention (should not match)",
			message:  "Hello @d",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := regex.FindAllStringSubmatch(tt.message, -1)

			found := make([]string, 0)
			for _, match := range matches {
				if len(match) > 1 {
					found = append(found, match[1])
				}
			}

			assert.Equal(t, tt.expected, found)
		})
	}
}

func TestResolveGroupMentionsSkipsUsersAndUnknownMentions(t *testing.T) {
	mentions := map[string]bool{
		"alice":   true,
		"dev":     true,
		"unknown": true,
	}
	groupsByName := map[string]*Group{
		"dev": &Group{Name: "dev"},
	}

	groups := resolveGroupMentions(
		mentions,
		func(name string) bool {
			return name == "alice"
		},
		func(name string) (*Group, error) {
			return groupsByName[name], nil
		},
		func(string, error) {
			t.Fatal("unexpected group lookup error")
		},
	)

	assert.Equal(t, map[string]*Group{"dev": groupsByName["dev"]}, groups)
}

func TestResolveGroupMentionsLogsLookupErrorsAndContinues(t *testing.T) {
	mentions := map[string]bool{
		"dev": true,
		"ops": true,
	}
	expectedErr := errors.New("kv lookup failed")
	loggedErrors := map[string]error{}

	groups := resolveGroupMentions(
		mentions,
		func(string) bool {
			return false
		},
		func(name string) (*Group, error) {
			if name == "ops" {
				return nil, expectedErr
			}
			return &Group{Name: name}, nil
		},
		func(name string, err error) {
			loggedErrors[name] = err
		},
	)

	assert.Contains(t, groups, "dev")
	assert.NotContains(t, groups, "ops")
	assert.Equal(t, map[string]error{"ops": expectedErr}, loggedErrors)
}

func TestActualGroupMentionLimitIgnoresUsersAndUnknownMentions(t *testing.T) {
	mentions := map[string]bool{
		"alice":   true,
		"bob":     true,
		"dev":     true,
		"unknown": true,
	}
	groupsByName := map[string]*Group{
		"dev": &Group{Name: "dev"},
	}

	groups := resolveGroupMentions(
		mentions,
		func(name string) bool {
			return name == "alice" || name == "bob"
		},
		func(name string) (*Group, error) {
			return groupsByName[name], nil
		},
		func(string, error) {
			t.Fatal("unexpected group lookup error")
		},
	)

	assert.Len(t, groups, 1, "only real groups should count toward MaxMentionsPerMessage")
}
