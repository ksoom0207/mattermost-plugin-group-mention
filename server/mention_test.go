package main

import (
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

			var found []string
			for _, match := range matches {
				if len(match) > 1 {
					found = append(found, match[1])
				}
			}

			assert.Equal(t, tt.expected, found)
		})
	}
}
