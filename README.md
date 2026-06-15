# Mattermost Group Mention Plugin

Enable custom `@group` mentions for Mattermost Community Edition. Create and manage groups of users that can be mentioned all at once, similar to Mattermost Enterprise Edition's group mention feature.

![Version](https://img.shields.io/badge/version-1.0.0-blue)
![Mattermost](https://img.shields.io/badge/Mattermost-v9.0%2B-blue)
![License](https://img.shields.io/badge/license-Apache%202.0-green)

## Features

- ✅ **Custom Group Mentions**: Mention multiple users at once using `@groupname`
- 🔒 **Public & Private Groups**: Control who can see and use groups
- 🛡️ **Permission Control**: Restrict group management to admins or allow all users
- ⚡ **Rate Limiting**: Prevent spam with configurable rate limits
- 🎯 **Smart Notifications**: Notify all group members without cluttering messages
- 📊 **Large Channel Protection**: Prevent @here-style abuse in large channels
- 🎨 **Visual Highlighting**: Group mentions are highlighted in messages
- 💬 **Slash Commands**: Easy group management via `/group` commands

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
  - [Creating Groups](#creating-groups)
  - [Managing Groups](#managing-groups)
  - [Mentioning Groups](#mentioning-groups)
- [Slash Commands](#slash-commands)
- [Configuration](#configuration)
- [Permissions](#permissions)
- [FAQ](#faq)
- [Development](#development)
- [License](#license)

## Installation

### From Release

1. Download the latest release from the [Releases page](../../releases)
2. Upload the plugin file in **System Console > Plugin Management**
3. Enable the plugin
4. Configure plugin settings as needed

### From Source

**Requirements:**
- Go 1.22+
- Node.js 20+
- npm or yarn

**Build Steps:**

```bash
# Clone the repository
git clone https://github.com/mattermost/mattermost-plugin-group-mention.git
cd mattermost-plugin-group-mention

# Build the plugin
make dist

# The plugin bundle will be created at: dist/com.mattermost.plugin-group-mention-1.0.0.tar.gz
```

**Upload to Mattermost:**

1. Go to **System Console > Plugin Management**
2. Click **Upload Plugin**
3. Select the `.tar.gz` file
4. Click **Enable** under the uploaded plugin

## Usage

### Creating Groups

Use the `/group create` command to create a new group:

```bash
/group create dev --public --members @alice @bob @charlie
```

Group names cannot match existing Mattermost usernames. If a user `@alice` exists, create a group like `@team-alice` or `@alice-group` instead.

**Options:**
- `--public`: Make the group visible to all team members (default)
- `--private`: Make the group visible only to owners and admins
- `--owners @user1 @user2`: Specify group owners
- `--members @user1 @user2`: Add initial members

**Examples:**

```bash
# Create a public dev team group
/group create dev --public --members @alice @bob @charlie

# Create a private leadership group
/group create leadership --private --owners @ceo --members @cto @cfo @coo

# Create a group with multiple owners
/group create ops --owners @admin1 @admin2 --members @ops1 @ops2 @ops3
```

### Managing Groups

**Add members:**
```bash
/group add dev @david @eve
```

**Remove members:**
```bash
/group remove dev @eve
```

**View group details:**
```bash
/group show dev
```

**List all groups:**
```bash
/group list
```

**Delete a group:**
```bash
/group delete dev
```

### Mentioning Groups

Once a group is created, mention it in any message using `@groupname`:

```
Hey @dev, please review this PR!
```

**What happens:**
- All members of the group receive a mention notification
- Members get a direct message with a link to the original message
- The original message displays `@dev` (no text expansion by default)
- Rate limits prevent spam and abuse

## Slash Commands

| Command | Description | Example |
|---------|-------------|---------|
| `/group create <name> [options]` | Create a new group | `/group create dev --public --members @alice @bob` |
| `/group add <name> @user1 ...` | Add members to a group | `/group add dev @charlie` |
| `/group remove <name> @user1 ...` | Remove members from a group | `/group remove dev @charlie` |
| `/group list` | List all groups in the current team | `/group list` |
| `/group show <name>` | Show details about a group | `/group show dev` |
| `/group delete <name>` | Delete a group | `/group delete dev` |
| `/group help` | Show help message | `/group help` |

## Configuration

Configure the plugin in **System Console > Plugins > Group Mention**.

### Settings

| Setting | Default | Description |
|---------|---------|-------------|
| **Maximum Users Per Group Mention** | 50 | Maximum number of users that can be notified from a single group mention |
| **Maximum Group Mentions Per Message** | 3 | Maximum number of group mentions allowed in a single message |
| **Rate Limit: Per User Per Minute** | 10 | Maximum group mentions a user can send per minute |
| **Rate Limit: Per Channel Per Minute** | 20 | Maximum group mentions allowed in a channel per minute |
| **Rate Limit: Per Group Per Minute** | 15 | Maximum times a specific group can be mentioned per minute |
| **Allow Regular Users to Create Groups** | false | If enabled, any user can create groups. Otherwise, only admins can. |
| **Default Group Visibility** | public | Default visibility for new groups (public or private) |
| **Mention Expansion Mode** | notify-only | How mentions are handled:<br>- `notify-only`: Keep @groupname as-is, send notifications<br>- `text-expand`: Replace with individual @mentions |
| **Large Channel Threshold** | 1000 | Number of members above which a channel is considered "large" |
| **Require Channel Admin in Large Channels** | true | Only channel admins can use group mentions in large channels |
| **Enable Debug Logging** | false | Enable detailed debug logs for troubleshooting |

### Recommended Settings

**For small teams (< 100 users):**
```
MaxExpandUsers: 50
MaxMentionsPerMessage: 5
PerUserPerMinute: 20
AllowUserManagedGroups: true
RequireChannelAdminInLargeChannels: false
```

**For large organizations (> 1000 users):**
```
MaxExpandUsers: 30
MaxMentionsPerMessage: 2
PerUserPerMinute: 5
AllowUserManagedGroups: false
RequireChannelAdminInLargeChannels: true
LargeChannelMemberThreshold: 500
```

## Permissions

### Permission Matrix

| Action | Regular User | Group Owner | Team Admin | System Admin |
|--------|--------------|-------------|------------|--------------|
| Create Group | ⚙️ Configurable | ✅ | ✅ | ✅ |
| View Public Group | ✅ | ✅ | ✅ | ✅ |
| View Private Group | ❌ | ✅ | ✅ | ✅ |
| Manage Own Group | ❌ | ✅ | ✅ | ✅ |
| Manage Any Group | ❌ | ❌ | ✅ | ✅ |
| Delete Group | ❌ | ✅ | ✅ | ✅ |
| Mention Group | ✅* | ✅ | ✅ | ✅ |

*Subject to rate limits and large channel restrictions

### Permission Levels

1. **System Administrator**: Full access to all groups and settings
2. **Team Administrator**: Can manage all groups within their team
3. **Group Owner**: Can manage groups they own (add/remove members, delete)
4. **Regular User**: Can use groups (if visible) and create groups (if enabled in settings)

## FAQ

### How is this different from Mattermost EE's group mentions?

This plugin provides similar functionality for **Community Edition** users:
- Uses local KV store instead of LDAP/AD sync
- Manual group management via slash commands
- No external directory integration
- Suitable for teams that don't need LDAP/AD integration

### What happens when a group name conflicts with a username?

New groups and renamed groups cannot use an existing Mattermost username. This prevents ambiguous mentions.

For existing conflicting names, the plugin follows this priority:
1. Real username mentions are processed first by Mattermost
2. Only non-matching patterns are checked against groups
3. To avoid conflicts, use distinctive group names (e.g., `team-dev` instead of `dev`)

### Can I use group mentions in direct messages?

Group mentions work in:
- ✅ Public channels
- ✅ Private channels
- ✅ Group messages
- ❌ Direct messages (1-on-1)

### What are the performance considerations?

**Optimizations:**
- Notifications are sent asynchronously
- Rate limiting prevents abuse
- Large group warnings protect server resources
- KV store operations are batched

**Limits:**
- Maximum 50 users per group mention (configurable)
- Maximum 3 group mentions per message (configurable)
- Rate limits per user, channel, and group

### How do I prevent abuse in large channels?

Enable these settings:
1. Set `RequireChannelAdminInLargeChannels` to `true`
2. Set `LargeChannelMemberThreshold` to appropriate value (e.g., 500)
3. Lower `MaxExpandUsers` to prevent very large mentions
4. Restrict group creation to admins (`AllowUserManagedGroups: false`)

### Can groups span multiple teams?

No, groups are scoped to a single team. To mention users across teams, create separate groups in each team.

### How do I migrate groups between servers?

Groups are stored in the KV store. To migrate:
1. Use the Mattermost mmctl tool to export KV data
2. Look for keys starting with `group:` and `group_index:`
3. Import to the new server using mmctl

### What happens if a group member leaves the team?

The member remains in the group until manually removed. To clean up:
1. Use `/group show <name>` to view members
2. Use `/group remove <name> @user` to remove inactive members

## Development

### Prerequisites

- Go 1.22+
- Node.js 20+
- Mattermost Server 9.0+

### Building

```bash
# Install dependencies
cd server && go mod download
cd ../webapp && npm install

# Build plugin
make dist

# Run tests
make test

# Check code style
make check-style
```

### Testing

**Run Go tests:**
```bash
cd server
go test -v -race ./...
```

**Run webapp tests:**
```bash
cd webapp
npm test
```

### Project Structure

```
mattermost-plugin-group-mention/
├── server/              # Go backend
│   ├── plugin.go       # Main plugin implementation
│   ├── models.go       # Data models
│   ├── kv.go          # KV store operations
│   ├── mention.go     # Mention parsing & notifications
│   ├── slash.go       # Slash command handlers
│   ├── api.go         # REST API endpoints
│   ├── permissions.go # Permission checks
│   ├── ratelimit.go   # Rate limiting
│   └── *_test.go      # Unit tests
├── webapp/             # TypeScript/React frontend
│   ├── src/
│   │   ├── index.tsx         # Plugin entry point
│   │   ├── api.ts            # API client
│   │   ├── components/       # React components
│   │   └── types/           # TypeScript types
│   └── package.json
├── plugin.json         # Plugin manifest
├── Makefile           # Build scripts
└── README.md          # Documentation
```

## Production QA Checklist

Before deploying to production, verify:

- [ ] Plugin builds without errors (`make dist`)
- [ ] All unit tests pass (`make test`)
- [ ] Code style checks pass (`make check-style`)
- [ ] Plugin activates successfully on test server
- [ ] Slash commands work (`/group create`, `/group list`, etc.)
- [ ] Group mentions trigger notifications
- [ ] Rate limits are enforced
- [ ] Permission checks work correctly
- [ ] Large channel restrictions work
- [ ] Public/private group visibility works
- [ ] API endpoints return correct data
- [ ] No errors in Mattermost server logs
- [ ] Plugin configuration UI loads correctly
- [ ] Settings changes take effect immediately
- [ ] Groups persist after server restart
- [ ] No performance degradation under load

## Troubleshooting

### Plugin doesn't activate

1. Check Mattermost server logs: `System Console > Logs`
2. Verify minimum server version (9.0+)
3. Check plugin file permissions
4. Try re-uploading the plugin

### Group mentions don't trigger notifications

1. Check if group exists: `/group show <name>`
2. Verify you're using correct syntax: `@groupname` (not `@ groupname`)
3. Check rate limits in plugin settings
4. Verify channel permissions (large channel restrictions)
5. Check server logs for errors

### Slash commands not working

1. Verify plugin is enabled
2. Check user permissions (can they create groups?)
3. Try `/group help` to verify command registration
4. Check for typos in group names

### Performance issues

1. Check group sizes (reduce `MaxExpandUsers` if needed)
2. Review rate limit settings
3. Enable large channel restrictions
4. Check Mattermost server resources

## Security Considerations

- All API endpoints require authentication
- Permission checks on all group operations
- Rate limiting prevents DoS attacks
- Input validation on group names and commands
- No PII logging
- CSRF protection on HTTP endpoints

## Support

- **Issues**: [GitHub Issues](../../issues)
- **Documentation**: This README
- **Mattermost Community**: [community.mattermost.com](https://community.mattermost.com)

## License

Apache License 2.0. See [LICENSE](LICENSE) file.

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Credits

Developed for the Mattermost Community Edition.

## Changelog

### v1.0.0 (2025-01-23)

- Initial release
- Core group mention functionality
- Slash commands for group management
- Rate limiting and permissions
- Public and private groups
- REST API for webapp
- Webapp UI components
- Comprehensive documentation
