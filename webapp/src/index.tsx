import React from 'react';

import {PluginRegistry, PluginClass} from './types';
import Client from './api';
import GroupPill from './components/GroupPill';

class Plugin implements PluginClass {
    private registry?: PluginRegistry;
    private store?: any;

    public initialize(registry: PluginRegistry, store: any): void {
        this.registry = registry;
        this.store = store;

        // Register message formatting hook to highlight @group mentions
        registry.registerMessageWillFormatHook(this.formatGroupMentions.bind(this));

        console.log('Group Mention Plugin initialized');
    }

    public uninitialize(): void {
        console.log('Group Mention Plugin uninitialized');
    }

    private formatGroupMentions(post: any, message: string): string {
        // This hook allows us to format @group mentions in messages
        // We'll highlight them similar to user mentions

        // Match @groupname patterns
        const groupMentionRegex = /@([a-zA-Z0-9_\-.]{2,64})\b/g;

        // Replace @group with styled spans
        const formattedMessage = message.replace(groupMentionRegex, (match, groupName) => {
            // Check if this is a known group (we'd need to fetch groups for the team)
            // For now, we'll style all @mentions that look like groups
            return `<span class="group-mention" style="background-color: #166de0; color: #ffffff; padding: 2px 6px; border-radius: 10px; font-weight: bold;">${match}</span>`;
        });

        return formattedMessage;
    }

    private async getTeamId(): Promise<string> {
        // Get current team ID from Redux store
        if (!this.store) {
            return '';
        }

        const state = this.store.getState();
        const currentTeamId = state?.entities?.teams?.currentTeamId || '';
        return currentTeamId;
    }

    private async fetchGroupsForTeam(teamId: string) {
        try {
            const groups = await Client.getGroups(teamId);
            return groups;
        } catch (error) {
            console.error('Failed to fetch groups:', error);
            return [];
        }
    }
}

// Register the plugin
if (typeof window !== 'undefined') {
    window.registerPlugin('com.mattermost.plugin-group-mention', new Plugin());
}

export default Plugin;
