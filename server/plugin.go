package main

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/mattermost/mattermost/server/public/plugin"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration
	configuration *Config

	// rateLimitCache stores rate limit entries
	rateLimitCache map[string]*RateLimitEntry
	rateLimitLock  sync.RWMutex
}

// OnActivate is called when the plugin is activated
func (p *Plugin) OnActivate() error {
	p.rateLimitCache = make(map[string]*RateLimitEntry)

	config := p.getConfiguration()
	if config == nil {
		return errors.New("failed to load configuration")
	}

	// Register slash command
	if err := p.API.RegisterCommand(&model.Command{
		Trigger:          "group",
		AutoComplete:     true,
		AutoCompleteDesc: "Manage custom mention groups",
		AutoCompleteHint: "[create|add|remove|list|show|delete] <args>",
		DisplayName:      "Group Mention",
		Description:      "Create and manage custom mention groups for @mentions",
	}); err != nil {
		return errors.Wrap(err, "failed to register /group command")
	}

	p.API.LogInfo("Group Mention Plugin activated")
	return nil
}

// getConfiguration retrieves the active configuration under lock
func (p *Plugin) getConfiguration() *Config {
	p.configurationLock.RLock()
	defer p.configurationLock.RUnlock()

	if p.configuration == nil {
		p.configuration = p.loadConfiguration()
	}

	return p.configuration
}

// loadConfiguration loads the plugin configuration from Mattermost
func (p *Plugin) loadConfiguration() *Config {
	config := &Config{
		MaxExpandUsers:                  50,
		MaxMentionsPerMessage:           3,
		PerUserPerMinute:                10,
		PerChannelPerMinute:             20,
		PerGroupPerMinute:               15,
		AllowUserManagedGroups:          false,
		DefaultGroupVisibility:          "public",
		ExpandMode:                      "notify-only",
		LargeChannelMemberThreshold:     1000,
		RequireChannelAdminInLargeChannels: true,
		EnableDebugLogging:              false,
	}

	// Load from plugin configuration
	pluginConfig := p.API.GetConfig()
	if pluginConfig != nil && pluginConfig.PluginSettings.Plugins != nil {
		if settings, ok := pluginConfig.PluginSettings.Plugins["com.mattermost.plugin-group-mention"]; ok {
			if v, ok := settings["MaxExpandUsers"].(float64); ok {
				config.MaxExpandUsers = int(v)
			}
			if v, ok := settings["MaxMentionsPerMessage"].(float64); ok {
				config.MaxMentionsPerMessage = int(v)
			}
			if v, ok := settings["PerUserPerMinute"].(float64); ok {
				config.PerUserPerMinute = int(v)
			}
			if v, ok := settings["PerChannelPerMinute"].(float64); ok {
				config.PerChannelPerMinute = int(v)
			}
			if v, ok := settings["PerGroupPerMinute"].(float64); ok {
				config.PerGroupPerMinute = int(v)
			}
			if v, ok := settings["AllowUserManagedGroups"].(bool); ok {
				config.AllowUserManagedGroups = v
			}
			if v, ok := settings["DefaultGroupVisibility"].(string); ok {
				config.DefaultGroupVisibility = v
			}
			if v, ok := settings["ExpandMode"].(string); ok {
				config.ExpandMode = v
			}
			if v, ok := settings["LargeChannelMemberThreshold"].(float64); ok {
				config.LargeChannelMemberThreshold = int(v)
			}
			if v, ok := settings["RequireChannelAdminInLargeChannels"].(bool); ok {
				config.RequireChannelAdminInLargeChannels = v
			}
			if v, ok := settings["EnableDebugLogging"].(bool); ok {
				config.EnableDebugLogging = v
			}
		}
	}

	return config
}

// OnConfigurationChange is called when the plugin configuration changes
func (p *Plugin) OnConfigurationChange() error {
	p.configurationLock.Lock()
	defer p.configurationLock.Unlock()

	p.configuration = p.loadConfiguration()
	return nil
}

// ServeHTTP handles HTTP requests to the plugin
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.handleHTTP(w, r)
}

// ExecuteCommand handles slash commands
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	return p.executeCommand(args)
}

// MessageHasBeenPosted is no longer used - group mentions are handled in MessageWillBePosted
// This hook is kept for potential future use (e.g., analytics, logging)
func (p *Plugin) MessageHasBeenPosted(c *plugin.Context, post *model.Post) {
	config := p.getConfiguration()
	if config.EnableDebugLogging {
		p.API.LogDebug("Message posted", "post_id", post.Id, "channel_id", post.ChannelId)
	}

	// Note: Group mention processing moved to MessageWillBePosted hook
	// This allows us to modify the message before it's saved, triggering native notifications
}

// logDebug logs debug messages if debug logging is enabled
func (p *Plugin) logDebug(msg string, keyValuePairs ...interface{}) {
	config := p.getConfiguration()
	if config != nil && config.EnableDebugLogging {
		p.API.LogDebug(msg, keyValuePairs...)
	}
}

// logInfo logs info messages
func (p *Plugin) logInfo(msg string, keyValuePairs ...interface{}) {
	p.API.LogInfo(msg, keyValuePairs...)
}

// logError logs error messages
func (p *Plugin) logError(msg string, keyValuePairs ...interface{}) {
	p.API.LogError(msg, keyValuePairs...)
}

// MarshalJSON is required for the plugin interface but not used
func (c *Config) MarshalJSON() ([]byte, error) {
	return json.Marshal(*c)
}
