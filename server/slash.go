package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
)

// executeCommand handles /group slash commands
func (p *Plugin) executeCommand(args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	split := strings.Fields(args.Command)
	if len(split) < 2 {
		return p.helpResponse(), nil
	}

	command := split[0]
	if command != "/group" {
		return &model.CommandResponse{}, nil
	}

	action := split[1]
	actionArgs := split[2:]

	switch action {
	case "create":
		return p.handleGroupCreate(args, actionArgs)
	case "add":
		return p.handleGroupAdd(args, actionArgs)
	case "remove":
		return p.handleGroupRemove(args, actionArgs)
	case "list":
		return p.handleGroupList(args, actionArgs)
	case "show":
		return p.handleGroupShow(args, actionArgs)
	case "delete":
		return p.handleGroupDelete(args, actionArgs)
	case "help":
		return p.helpResponse(), nil
	default:
		return p.errorResponse(fmt.Sprintf("Unknown command: %s", action)), nil
	}
}

// handleGroupCreate handles /group create command
func (p *Plugin) handleGroupCreate(args *model.CommandArgs, cmdArgs []string) (*model.CommandResponse, *model.AppError) {
	if len(cmdArgs) < 1 {
		return p.errorResponse("Usage: /group create <name> [--public|--private] [--owners @user1 @user2] [--members @user1 @user2]"), nil
	}

	teamID := args.TeamId
	if !p.canManageGroups(args.UserId, teamID) {
		return p.errorResponse("You don't have permission to create groups."), nil
	}

	groupName := strings.ToLower(cmdArgs[0])
	visibility := p.getConfiguration().DefaultGroupVisibility
	var owners, members []string

	// Parse flags
	i := 1
	for i < len(cmdArgs) {
		flag := cmdArgs[i]
		switch flag {
		case "--public":
			visibility = "public"
			i++
		case "--private":
			visibility = "private"
			i++
		case "--owners":
			i++
			for i < len(cmdArgs) && !strings.HasPrefix(cmdArgs[i], "--") {
				username := strings.TrimPrefix(cmdArgs[i], "@")
				user, _ := p.API.GetUserByUsername(username)
				if user != nil {
					owners = append(owners, user.Id)
				}
				i++
			}
		case "--members":
			i++
			for i < len(cmdArgs) && !strings.HasPrefix(cmdArgs[i], "--") {
				username := strings.TrimPrefix(cmdArgs[i], "@")
				user, _ := p.API.GetUserByUsername(username)
				if user != nil {
					members = append(members, user.Id)
				}
				i++
			}
		default:
			i++
		}
	}

	// Add creator as owner if no owners specified
	if len(owners) == 0 {
		owners = append(owners, args.UserId)
	}

	// Create group
	group, err := p.createGroup(teamID, groupName, visibility, owners, members)
	if err != nil {
		return p.errorResponse(fmt.Sprintf("Failed to create group: %s", err.Error())), nil
	}

	message := fmt.Sprintf("✅ Group `@%s` created successfully!\n", group.Name)
	message += fmt.Sprintf("- Visibility: %s\n", group.Visibility)
	message += fmt.Sprintf("- Owners: %d\n", len(group.Owners))
	message += fmt.Sprintf("- Members: %d", len(group.Members))

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         message,
	}, nil
}

// handleGroupAdd handles /group add command
func (p *Plugin) handleGroupAdd(args *model.CommandArgs, cmdArgs []string) (*model.CommandResponse, *model.AppError) {
	if len(cmdArgs) < 2 {
		return p.errorResponse("Usage: /group add <name> @user1 @user2 ..."), nil
	}

	teamID := args.TeamId
	groupName := strings.ToLower(cmdArgs[0])

	group, err := p.getGroup(teamID, groupName)
	if err != nil {
		return p.errorResponse(fmt.Sprintf("Failed to get group: %s", err.Error())), nil
	}
	if group == nil {
		return p.errorResponse(fmt.Sprintf("Group `@%s` not found.", groupName)), nil
	}

	if !p.canManageGroup(args.UserId, group) {
		return p.errorResponse("You don't have permission to manage this group."), nil
	}

	// Parse user mentions
	var userIDs []string
	for _, username := range cmdArgs[1:] {
		username = strings.TrimPrefix(username, "@")
		user, _ := p.API.GetUserByUsername(username)
		if user != nil {
			userIDs = append(userIDs, user.Id)
		}
	}

	if len(userIDs) == 0 {
		return p.errorResponse("No valid users specified."), nil
	}

	// Add members
	if err := p.addMembersToGroup(teamID, groupName, userIDs); err != nil {
		return p.errorResponse(fmt.Sprintf("Failed to add members: %s", err.Error())), nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("✅ Added %d member(s) to group `@%s`", len(userIDs), groupName),
	}, nil
}

// handleGroupRemove handles /group remove command
func (p *Plugin) handleGroupRemove(args *model.CommandArgs, cmdArgs []string) (*model.CommandResponse, *model.AppError) {
	if len(cmdArgs) < 2 {
		return p.errorResponse("Usage: /group remove <name> @user1 @user2 ..."), nil
	}

	teamID := args.TeamId
	groupName := strings.ToLower(cmdArgs[0])

	group, err := p.getGroup(teamID, groupName)
	if err != nil {
		return p.errorResponse(fmt.Sprintf("Failed to get group: %s", err.Error())), nil
	}
	if group == nil {
		return p.errorResponse(fmt.Sprintf("Group `@%s` not found.", groupName)), nil
	}

	if !p.canManageGroup(args.UserId, group) {
		return p.errorResponse("You don't have permission to manage this group."), nil
	}

	// Parse user mentions
	var userIDs []string
	for _, username := range cmdArgs[1:] {
		username = strings.TrimPrefix(username, "@")
		user, _ := p.API.GetUserByUsername(username)
		if user != nil {
			userIDs = append(userIDs, user.Id)
		}
	}

	if len(userIDs) == 0 {
		return p.errorResponse("No valid users specified."), nil
	}

	// Remove members
	if err := p.removeMembersFromGroup(teamID, groupName, userIDs); err != nil {
		return p.errorResponse(fmt.Sprintf("Failed to remove members: %s", err.Error())), nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("✅ Removed %d member(s) from group `@%s`", len(userIDs), groupName),
	}, nil
}

// handleGroupList handles /group list command
func (p *Plugin) handleGroupList(args *model.CommandArgs, cmdArgs []string) (*model.CommandResponse, *model.AppError) {
	teamID := args.TeamId

	groups, err := p.listGroups(teamID)
	if err != nil {
		return p.errorResponse(fmt.Sprintf("Failed to list groups: %s", err.Error())), nil
	}

	if len(groups) == 0 {
		return &model.CommandResponse{
			ResponseType: model.CommandResponseTypeEphemeral,
			Text:         "No groups found for this team.",
		}, nil
	}

	message := fmt.Sprintf("## Groups for this team (%d)\n\n", len(groups))
	for _, group := range groups {
		if p.canViewGroup(args.UserId, group) {
			visibility := "🔓 Public"
			if group.Visibility == "private" {
				visibility = "🔒 Private"
			}
			message += fmt.Sprintf("- **@%s** %s - %d members\n", group.Name, visibility, len(group.Members))
		}
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         message,
	}, nil
}

// handleGroupShow handles /group show command
func (p *Plugin) handleGroupShow(args *model.CommandArgs, cmdArgs []string) (*model.CommandResponse, *model.AppError) {
	if len(cmdArgs) < 1 {
		return p.errorResponse("Usage: /group show <name>"), nil
	}

	teamID := args.TeamId
	groupName := strings.ToLower(cmdArgs[0])

	group, err := p.getGroup(teamID, groupName)
	if err != nil {
		return p.errorResponse(fmt.Sprintf("Failed to get group: %s", err.Error())), nil
	}
	if group == nil {
		return p.errorResponse(fmt.Sprintf("Group `@%s` not found.", groupName)), nil
	}

	if !p.canViewGroup(args.UserId, group) {
		return p.errorResponse("You don't have permission to view this group."), nil
	}

	message := fmt.Sprintf("## Group: @%s\n\n", group.Name)
	message += fmt.Sprintf("**Visibility:** %s\n", group.Visibility)
	message += fmt.Sprintf("**Members:** %d\n", len(group.Members))
	message += fmt.Sprintf("**Owners:** %d\n", len(group.Owners))
	message += fmt.Sprintf("**Created:** %s\n", group.CreatedAt.Format("2006-01-02 15:04:05"))

	if p.canViewGroupMembers(args.UserId, group) {
		message += "\n**Member list:**\n"
		for _, memberID := range group.Members {
			user, appErr := p.API.GetUser(memberID)
			if appErr == nil {
				message += fmt.Sprintf("- @%s\n", user.Username)
			}
		}
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         message,
	}, nil
}

// handleGroupDelete handles /group delete command
func (p *Plugin) handleGroupDelete(args *model.CommandArgs, cmdArgs []string) (*model.CommandResponse, *model.AppError) {
	if len(cmdArgs) < 1 {
		return p.errorResponse("Usage: /group delete <name>"), nil
	}

	teamID := args.TeamId
	groupName := strings.ToLower(cmdArgs[0])

	group, err := p.getGroup(teamID, groupName)
	if err != nil {
		return p.errorResponse(fmt.Sprintf("Failed to get group: %s", err.Error())), nil
	}
	if group == nil {
		return p.errorResponse(fmt.Sprintf("Group `@%s` not found.", groupName)), nil
	}

	if !p.canManageGroup(args.UserId, group) {
		return p.errorResponse("You don't have permission to delete this group."), nil
	}

	// Delete group
	if err := p.deleteGroup(teamID, groupName); err != nil {
		return p.errorResponse(fmt.Sprintf("Failed to delete group: %s", err.Error())), nil
	}

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         fmt.Sprintf("✅ Group `@%s` deleted successfully.", groupName),
	}, nil
}

// helpResponse returns the help message
func (p *Plugin) helpResponse() *model.CommandResponse {
	message := `## Group Mention Plugin - Help

Create and manage custom mention groups for your team.

### Commands

* **/group create \<name\> [--public|--private] [--owners \@user1 \@user2] [--members \@user1 \@user2]**
  Create a new group

* **/group add \<name\> \@user1 \@user2 ...**
  Add members to a group

* **/group remove \<name\> \@user1 \@user2 ...**
  Remove members from a group

* **/group list**
  List all groups for the current team

* **/group show \<name\>**
  Show details about a specific group

* **/group delete \<name\>**
  Delete a group

* **/group help**
  Show this help message

### Usage

Once created, mention a group using ` + "`@groupname`" + ` in any message.
All group members will receive a notification.

### Examples

` + "```" + `
/group create dev --public --members @alice @bob @charlie
/group add dev @david
/group show dev
` + "```" + `
`

	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         message,
	}
}

// errorResponse returns an error response
func (p *Plugin) errorResponse(message string) *model.CommandResponse {
	return &model.CommandResponse{
		ResponseType: model.CommandResponseTypeEphemeral,
		Text:         "❌ " + message,
	}
}
