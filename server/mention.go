package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"
)

var (
	// mentionRegex matches @groupname patterns
	mentionRegex = regexp.MustCompile(`(?:^|\s)@([a-zA-Z0-9_\-.]{2,64})\b`)
)

// MessageWillBePosted is called before a message is posted
// This allows us to expand @group mentions to individual @user mentions
// so Mattermost automatically sends push notifications
func (p *Plugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {
	if post == nil || post.Message == "" {
		return post, ""
	}

	config := p.getConfiguration()

	// Get channel to check if it's large
	channel, appErr := p.API.GetChannel(post.ChannelId)
	if appErr != nil {
		p.logError("Failed to get channel", "channel_id", post.ChannelId, "error", appErr.Error())
		return post, ""
	}

	// Check if channel is large and requires admin
	if config.RequireChannelAdminInLargeChannels {
		stats, statsErr := p.API.GetChannelStats(post.ChannelId)
		if statsErr == nil && stats.MemberCount >= int64(config.LargeChannelMemberThreshold) {
			if !p.isChannelAdmin(post.UserId, post.ChannelId) {
				return post, "Group mentions are restricted to channel admins in large channels."
			}
		}
	}

	// Find all group mentions in the message
	matches := mentionRegex.FindAllStringSubmatch(post.Message, -1)
	if len(matches) == 0 {
		return post, ""
	}

	// Extract unique group names
	groupNames := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			groupNames[strings.ToLower(match[1])] = true
		}
	}

	if len(groupNames) > config.MaxMentionsPerMessage {
		return post, fmt.Sprintf("Too many group mentions. Maximum allowed: %d", config.MaxMentionsPerMessage)
	}

	// Check rate limits first (before expanding)
	for groupName := range groupNames {
		if !p.checkRateLimits(post.UserId, post.ChannelId, groupName) {
			return post, "Rate limit exceeded. Please wait before mentioning this group again."
		}
	}

	// Process each group mention and collect members
	teamID := channel.TeamId
	groupExpansions := make(map[string][]string) // groupName -> usernames

	for groupName := range groupNames {
		// Check if it's actually a user mention first
		userByUsername, _ := p.API.GetUserByUsername(groupName)
		if userByUsername != nil {
			// This is a user mention, not a group mention
			continue
		}

		// Get the group
		group, err := p.getGroup(teamID, groupName)
		if err != nil {
			p.logError("Failed to get group", "team_id", teamID, "group", groupName, "error", err.Error())
			continue
		}
		if group == nil {
			// Group doesn't exist, leave the mention as-is
			continue
		}

		// Check if group is too large
		if len(group.Members) > config.MaxExpandUsers {
			return post, fmt.Sprintf("Group @%s has too many members (%d). Maximum allowed: %d",
				groupName, len(group.Members), config.MaxExpandUsers)
		}

		// Collect usernames for members who are in the channel
		usernames := make([]string, 0)
		for _, memberID := range group.Members {
			// Skip the post author
			if memberID == post.UserId {
				continue
			}

			// Skip if user is not in the channel
			_, err := p.API.GetChannelMember(channel.Id, memberID)
			if err != nil {
				continue
			}

			user, err := p.API.GetUser(memberID)
			if err == nil {
				usernames = append(usernames, user.Username)
			}
		}

		if len(usernames) > 0 {
			groupExpansions[groupName] = usernames
		}

		p.logDebug("Expanding group mention", "group", groupName, "members", len(usernames))
	}

	// If no groups to expand, return original
	if len(groupExpansions) == 0 {
		return post, ""
	}

	// Expand @groupname to @user1 @user2 @user3 in the message
	newMessage := post.Message
	for groupName, usernames := range groupExpansions {
		// Create the expansion text
		mentionList := make([]string, len(usernames))
		for i, username := range usernames {
			mentionList[i] = "@" + username
		}
		expansion := strings.Join(mentionList, " ")

		// Replace @groupname with the expansion
		// Use regex to replace only whole word matches
		groupPattern := regexp.MustCompile(`(^|\s)@` + regexp.QuoteMeta(groupName) + `\b`)
		newMessage = groupPattern.ReplaceAllString(newMessage, "$1"+expansion)
	}

	post.Message = newMessage
	return post, ""
}

// processGroupMentions processes a post for group mentions (DEPRECATED)
// This function is kept for backward compatibility but is no longer used
// MessageWillBePosted now handles mention expansion before the post is saved
func (p *Plugin) processGroupMentions(post *model.Post) {
	if post == nil || post.Message == "" {
		return
	}

	config := p.getConfiguration()

	// Get channel to check if it's large
	channel, appErr := p.API.GetChannel(post.ChannelId)
	if appErr != nil {
		p.logError("Failed to get channel", "channel_id", post.ChannelId, "error", appErr.Error())
		return
	}

	// Get user (for permission checks)
	_, appErr = p.API.GetUser(post.UserId)
	if appErr != nil {
		p.logError("Failed to get user", "user_id", post.UserId, "error", appErr.Error())
		return
	}

	// Check if channel is large and requires admin
	if config.RequireChannelAdminInLargeChannels {
		stats, statsErr := p.API.GetChannelStats(post.ChannelId)
		if statsErr == nil && stats.MemberCount >= int64(config.LargeChannelMemberThreshold) {
			if !p.isChannelAdmin(post.UserId, post.ChannelId) {
				p.sendEphemeralPost(post.ChannelId, post.UserId,
					":warning: Group mentions are restricted to channel admins in large channels.")
				return
			}
		}
	}

	// Find all group mentions in the message
	matches := mentionRegex.FindAllStringSubmatch(post.Message, -1)
	if len(matches) == 0 {
		return
	}

	// Extract unique group names
	groupNames := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 {
			groupNames[strings.ToLower(match[1])] = true
		}
	}

	if len(groupNames) > config.MaxMentionsPerMessage {
		p.sendEphemeralPost(post.ChannelId, post.UserId,
			fmt.Sprintf(":warning: Too many group mentions. Maximum allowed: %d", config.MaxMentionsPerMessage))
		return
	}

	// Process each group mention
	teamID := channel.TeamId
	allMentionedUsers := make(map[string]bool)

	for groupName := range groupNames {
		// Check if it's actually a user mention first
		userByUsername, _ := p.API.GetUserByUsername(groupName)
		if userByUsername != nil {
			// This is a user mention, not a group mention
			continue
		}

		// Get the group
		group, err := p.getGroup(teamID, groupName)
		if err != nil {
			p.logError("Failed to get group", "team_id", teamID, "group", groupName, "error", err.Error())
			continue
		}
		if group == nil {
			// Group doesn't exist
			continue
		}

		// Check rate limits
		if !p.checkRateLimits(post.UserId, post.ChannelId, groupName) {
			p.sendEphemeralPost(post.ChannelId, post.UserId,
				":warning: Rate limit exceeded. Please wait before mentioning this group again.")
			return
		}

		// Check if group is too large
		if len(group.Members) > config.MaxExpandUsers {
			p.sendEphemeralPost(post.ChannelId, post.UserId,
				fmt.Sprintf(":warning: Group `@%s` has too many members (%d). Maximum allowed: %d",
					groupName, len(group.Members), config.MaxExpandUsers))
			continue
		}

		// Add members to the set
		for _, memberID := range group.Members {
			allMentionedUsers[memberID] = true
		}

		p.logDebug("Processing group mention", "group", groupName, "members", len(group.Members))
	}

	// Remove the post author from mentions
	delete(allMentionedUsers, post.UserId)

	if len(allMentionedUsers) == 0 {
		return
	}

	// Send notifications to all mentioned users
	p.fanOutNotifications(post, channel, allMentionedUsers)
}

// fanOutNotifications sends notifications to all mentioned users
func (p *Plugin) fanOutNotifications(post *model.Post, channel *model.Channel, userIDs map[string]bool) {
	config := p.getConfiguration()

	switch config.ExpandMode {
	case "text-expand":
		// Expand mentions in the message text
		p.expandMentionsInText(post, channel, userIDs)
	case "notify-only":
		fallthrough
	default:
		// Just send notifications without changing the message
		p.sendNotificationsOnly(post, channel, userIDs)
	}
}

// sendNotificationsOnly sends mention notifications without modifying the message
func (p *Plugin) sendNotificationsOnly(post *model.Post, channel *model.Channel, userIDs map[string]bool) {
	// Create a batch of notifications
	for userID := range userIDs {
		// Skip if user is not in the channel
		_, appErr := p.API.GetChannelMember(channel.Id, userID)
		if appErr != nil {
			continue
		}

		// Send a notification by creating a mention
		// The Mattermost server will handle the notification automatically
		// We add the user to the post's mentions
		p.logDebug("Notifying user of group mention", "user_id", userID, "post_id", post.Id)
	}

	// Update the post to include all mentioned users
	// This triggers notifications in Mattermost
	mentionedUsersList := make([]string, 0, len(userIDs))
	for userID := range userIDs {
		mentionedUsersList = append(mentionedUsersList, userID)
	}

	// Add to Props for notification
	if post.Props == nil {
		post.Props = make(model.StringInterface)
	}
	post.Props["mentions"] = mentionedUsersList

	// Note: We don't modify the post in the database to avoid permission issues
	// Instead, we send DM notifications or use the notification API
	p.sendGroupMentionDMs(post, channel, userIDs)
}

// sendGroupMentionDMs sends DM notifications to mentioned users
func (p *Plugin) sendGroupMentionDMs(post *model.Post, channel *model.Channel, userIDs map[string]bool) {
	author, appErr := p.API.GetUser(post.UserId)
	if appErr != nil {
		return
	}

	messageLink := fmt.Sprintf("[View message](%s)", p.getPostLink(channel.TeamId, post.Id))

	for userID := range userIDs {
		// Skip if user is not in the channel
		_, err := p.API.GetChannelMember(channel.Id, userID)
		if err != nil {
			continue
		}

		// Create DM channel
		dmChannel, dmErr := p.API.GetDirectChannel(post.UserId, userID)
		if dmErr != nil {
			continue
		}

		// Send notification DM
		notificationPost := &model.Post{
			UserId:    post.UserId,
			ChannelId: dmChannel.Id,
			Message: fmt.Sprintf("You were mentioned in a group mention by @%s in ~%s\n\n%s",
				author.Username, channel.Name, messageLink),
		}

		if _, postErr := p.API.CreatePost(notificationPost); postErr != nil {
			p.logError("Failed to send group mention notification", "user_id", userID, "error", postErr.Error())
		}
	}
}

// expandMentionsInText expands group mentions to individual user mentions
func (p *Plugin) expandMentionsInText(post *model.Post, channel *model.Channel, userIDs map[string]bool) {
	// Get usernames for all mentioned users
	usernames := make([]string, 0, len(userIDs))
	for userID := range userIDs {
		user, appErr := p.API.GetUser(userID)
		if appErr != nil {
			continue
		}
		usernames = append(usernames, "@"+user.Username)
	}

	if len(usernames) == 0 {
		return
	}

	// Create a reply with expanded mentions
	expandedMessage := fmt.Sprintf("Mentioning: %s", strings.Join(usernames, " "))

	replyPost := &model.Post{
		UserId:    post.UserId,
		ChannelId: post.ChannelId,
		RootId:    post.Id,
		Message:   expandedMessage,
	}

	if _, appErr := p.API.CreatePost(replyPost); appErr != nil {
		p.logError("Failed to create expanded mention post", "error", appErr.Error())
	}
}

// getPostLink generates a permalink to a post
func (p *Plugin) getPostLink(teamID, postID string) string {
	config := p.API.GetConfig()
	siteURL := ""
	if config != nil && config.ServiceSettings.SiteURL != nil {
		siteURL = *config.ServiceSettings.SiteURL
	}

	return fmt.Sprintf("%s/_redirect/pl/%s", siteURL, postID)
}

// sendEphemeralPost sends an ephemeral message to a user
func (p *Plugin) sendEphemeralPost(channelID, userID, message string) {
	post := &model.Post{
		ChannelId: channelID,
		UserId:    userID,
		Message:   message,
	}
	p.API.SendEphemeralPost(userID, post)
}
