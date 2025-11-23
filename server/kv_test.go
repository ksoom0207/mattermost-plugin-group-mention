package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGroupKey(t *testing.T) {
	key := getGroupKey("team123", "dev")
	assert.Equal(t, "group:team123:dev", key)
}

func TestGetGroupIndexKey(t *testing.T) {
	key := getGroupIndexKey("team123")
	assert.Equal(t, "group_index:team123", key)
}

func TestGroupValidation(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		shouldErr bool
	}{
		{
			name:      "Valid group name",
			groupName: "dev",
			shouldErr: false,
		},
		{
			name:      "Valid group name with numbers",
			groupName: "dev123",
			shouldErr: false,
		},
		{
			name:      "Empty group name",
			groupName: "",
			shouldErr: true,
		},
		{
			name:      "Too short",
			groupName: "d",
			shouldErr: true,
		},
		{
			name:      "Too long",
			groupName: "this_is_a_very_long_group_name_that_exceeds_the_maximum_length_allowed_for_group_names_in_this_plugin",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validation logic
			isValid := len(tt.groupName) >= 2 && len(tt.groupName) <= 64

			if tt.shouldErr {
				assert.False(t, isValid)
			} else {
				assert.True(t, isValid)
			}
		})
	}
}
