package main

// canManageGroups checks if a user can create/manage groups
func (p *Plugin) canManageGroups(userID, teamID string) bool {
	config := p.getConfiguration()

	// If user-managed groups are allowed, anyone can manage
	if config.AllowUserManagedGroups {
		return true
	}

	// Check if user is a team admin
	if p.isTeamAdmin(userID, teamID) {
		return true
	}

	// Check if user is a system admin
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return false
	}

	return user.IsSystemAdmin()
}

// canManageGroup checks if a user can manage a specific group
func (p *Plugin) canManageGroup(userID string, group *Group) bool {
	// Check if user is an owner of the group
	for _, ownerID := range group.Owners {
		if ownerID == userID {
			return true
		}
	}

	// Check if user is a team admin
	if p.isTeamAdmin(userID, group.TeamID) {
		return true
	}

	// Check if user is a system admin
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return false
	}

	return user.IsSystemAdmin()
}

// canViewGroup checks if a user can view a group
func (p *Plugin) canViewGroup(userID string, group *Group) bool {
	// Public groups can be viewed by anyone on the team
	if group.Visibility == "public" {
		return true
	}

	// Private groups can only be viewed by owners and admins
	return p.canManageGroup(userID, group)
}

// canViewGroupMembers checks if a user can view group members
func (p *Plugin) canViewGroupMembers(userID string, group *Group) bool {
	// Public groups: members list visible to all
	if group.Visibility == "public" {
		return true
	}

	// Private groups: only owners and admins can see members
	return p.canManageGroup(userID, group)
}

// isTeamAdmin checks if a user is a team admin
func (p *Plugin) isTeamAdmin(userID, teamID string) bool {
	member, appErr := p.API.GetTeamMember(teamID, userID)
	if appErr != nil {
		return false
	}

	return member.SchemeAdmin
}

// isChannelAdmin checks if a user is a channel admin
func (p *Plugin) isChannelAdmin(userID, channelID string) bool {
	// Get user
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return false
	}

	// System admins have all permissions
	if user.IsSystemAdmin() {
		return true
	}

	// Check channel membership
	member, appErr := p.API.GetChannelMember(channelID, userID)
	if appErr != nil {
		return false
	}

	// Check if user has channel admin role
	return member.SchemeAdmin
}

// isSystemAdmin checks if a user is a system admin
func (p *Plugin) isSystemAdmin(userID string) bool {
	user, appErr := p.API.GetUser(userID)
	if appErr != nil {
		return false
	}

	return user.IsSystemAdmin()
}

// getUserTeamID gets the team ID for a user based on the command context
func (p *Plugin) getUserTeamID(userID string, commandTeamID string) string {
	// If team ID is provided in command args, use it
	if commandTeamID != "" {
		return commandTeamID
	}

	// Otherwise, get user's teams and use the first one
	teams, appErr := p.API.GetTeamsForUser(userID)
	if appErr != nil || len(teams) == 0 {
		return ""
	}

	return teams[0].Id
}
