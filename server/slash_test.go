package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupNameUsernameConflictMessage(t *testing.T) {
	assert.Equal(
		t,
		"Group name `@alice` conflicts with an existing user. Choose a different group name, such as `team-alice` or `alice-group`.",
		groupNameUsernameConflictMessage("alice"),
	)
}

func TestNormalizeGroupNameForCommand(t *testing.T) {
	assert.Equal(t, "alice", normalizeGroupNameForCommand("@Alice"))
	assert.Equal(t, "team-alice", normalizeGroupNameForCommand("Team-Alice"))
}
