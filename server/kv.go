package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

const (
	groupKeyPrefix      = "group"
	groupIndexKeyPrefix = "group_index"
	schemaVersionKey    = "schema_version"
	currentSchemaVersion = 1
)

// getGroupKey returns the KV key for a group
func getGroupKey(teamID, groupName string) string {
	return fmt.Sprintf("%s:%s:%s", groupKeyPrefix, teamID, groupName)
}

// getGroupIndexKey returns the KV key for a team's group index
func getGroupIndexKey(teamID string) string {
	return fmt.Sprintf("%s:%s", groupIndexKeyPrefix, teamID)
}

// createGroup creates a new group in the KV store
func (p *Plugin) createGroup(teamID, name, visibility string, owners, members []string) (*Group, error) {
	// Validate group name
	if name == "" || len(name) < 2 || len(name) > 64 {
		return nil, errors.New("group name must be between 2 and 64 characters")
	}

	// Check if group already exists
	existing, _ := p.getGroup(teamID, name)
	if existing != nil {
		return nil, errors.New("group already exists")
	}

	group := &Group{
		ID:         uuid.New().String(),
		Name:       name,
		TeamID:     teamID,
		Visibility: visibility,
		Owners:     owners,
		Members:    members,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Store group
	if err := p.storeGroup(group); err != nil {
		return nil, err
	}

	// Update index
	if err := p.addGroupToIndex(teamID, name); err != nil {
		return nil, err
	}

	return group, nil
}

// getGroup retrieves a group from the KV store
func (p *Plugin) getGroup(teamID, name string) (*Group, error) {
	key := getGroupKey(teamID, name)
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to get group from KV store")
	}
	if data == nil {
		return nil, nil
	}

	var group Group
	if err := json.Unmarshal(data, &group); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal group")
	}

	return &group, nil
}

// storeGroup stores a group in the KV store
func (p *Plugin) storeGroup(group *Group) error {
	group.UpdatedAt = time.Now()

	data, err := json.Marshal(group)
	if err != nil {
		return errors.Wrap(err, "failed to marshal group")
	}

	key := getGroupKey(group.TeamID, group.Name)
	if appErr := p.API.KVSet(key, data); appErr != nil {
		return errors.Wrap(appErr, "failed to store group in KV store")
	}

	return nil
}

// deleteGroup deletes a group from the KV store
func (p *Plugin) deleteGroup(teamID, name string) error {
	key := getGroupKey(teamID, name)
	if appErr := p.API.KVDelete(key); appErr != nil {
		return errors.Wrap(appErr, "failed to delete group from KV store")
	}

	// Update index
	if err := p.removeGroupFromIndex(teamID, name); err != nil {
		return err
	}

	return nil
}

// listGroups lists all groups for a team
func (p *Plugin) listGroups(teamID string) ([]*Group, error) {
	index, err := p.getGroupIndex(teamID)
	if err != nil {
		return nil, err
	}

	if index == nil || len(index.GroupNames) == 0 {
		return []*Group{}, nil
	}

	groups := make([]*Group, 0, len(index.GroupNames))
	for _, name := range index.GroupNames {
		group, err := p.getGroup(teamID, name)
		if err != nil {
			p.logError("Failed to get group", "team_id", teamID, "name", name, "error", err.Error())
			continue
		}
		if group != nil {
			groups = append(groups, group)
		}
	}

	return groups, nil
}

// getGroupIndex retrieves the group index for a team
func (p *Plugin) getGroupIndex(teamID string) (*GroupIndex, error) {
	key := getGroupIndexKey(teamID)
	data, appErr := p.API.KVGet(key)
	if appErr != nil {
		return nil, errors.Wrap(appErr, "failed to get group index from KV store")
	}
	if data == nil {
		return &GroupIndex{
			TeamID:     teamID,
			GroupNames: []string{},
		}, nil
	}

	var index GroupIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal group index")
	}

	return &index, nil
}

// storeGroupIndex stores the group index for a team
func (p *Plugin) storeGroupIndex(index *GroupIndex) error {
	data, err := json.Marshal(index)
	if err != nil {
		return errors.Wrap(err, "failed to marshal group index")
	}

	key := getGroupIndexKey(index.TeamID)
	if appErr := p.API.KVSet(key, data); appErr != nil {
		return errors.Wrap(appErr, "failed to store group index in KV store")
	}

	return nil
}

// addGroupToIndex adds a group name to the team's index
func (p *Plugin) addGroupToIndex(teamID, groupName string) error {
	index, err := p.getGroupIndex(teamID)
	if err != nil {
		return err
	}

	// Check if already in index
	for _, name := range index.GroupNames {
		if name == groupName {
			return nil // Already in index
		}
	}

	index.GroupNames = append(index.GroupNames, groupName)
	return p.storeGroupIndex(index)
}

// removeGroupFromIndex removes a group name from the team's index
func (p *Plugin) removeGroupFromIndex(teamID, groupName string) error {
	index, err := p.getGroupIndex(teamID)
	if err != nil {
		return err
	}

	newNames := make([]string, 0, len(index.GroupNames))
	for _, name := range index.GroupNames {
		if name != groupName {
			newNames = append(newNames, name)
		}
	}

	index.GroupNames = newNames
	return p.storeGroupIndex(index)
}

// addMembersToGroup adds members to a group
func (p *Plugin) addMembersToGroup(teamID, groupName string, userIDs []string) error {
	group, err := p.getGroup(teamID, groupName)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	// Add members (avoid duplicates)
	memberSet := make(map[string]bool)
	for _, id := range group.Members {
		memberSet[id] = true
	}
	for _, id := range userIDs {
		memberSet[id] = true
	}

	group.Members = make([]string, 0, len(memberSet))
	for id := range memberSet {
		group.Members = append(group.Members, id)
	}

	return p.storeGroup(group)
}

// removeMembersFromGroup removes members from a group
func (p *Plugin) removeMembersFromGroup(teamID, groupName string, userIDs []string) error {
	group, err := p.getGroup(teamID, groupName)
	if err != nil {
		return err
	}
	if group == nil {
		return errors.New("group not found")
	}

	// Remove members
	removeSet := make(map[string]bool)
	for _, id := range userIDs {
		removeSet[id] = true
	}

	newMembers := make([]string, 0)
	for _, id := range group.Members {
		if !removeSet[id] {
			newMembers = append(newMembers, id)
		}
	}

	group.Members = newMembers
	return p.storeGroup(group)
}
