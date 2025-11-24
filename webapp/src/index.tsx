import {PluginRegistry, PluginClass} from './types';
import Client from './api';

class Plugin implements PluginClass {
    private registry?: PluginRegistry;
    private store?: any;

    public initialize(registry: PluginRegistry, store: any): void {
        this.registry = registry;
        this.store = store;

        // Note: Group mention detection and notifications are handled server-side.
        // The server processes @group mentions in MessageHasBeenPosted hook and
        // sends notifications to all group members.

        // Future enhancement: Add autocomplete for @group mentions here
        // registry.registerAutocompleteProvider(...)

        console.log('Group Mention Plugin initialized');
    }

    public uninitialize(): void {
        console.log('Group Mention Plugin uninitialized');
    }

    // Utility methods for future enhancements (autocomplete, UI components, etc.)
    private async getTeamId(): Promise<string> {
        if (!this.store) {
            return '';
        }
        const state = this.store.getState();
        return state?.entities?.teams?.currentTeamId || '';
    }

    private async fetchGroupsForTeam(teamId: string) {
        try {
            return await Client.getGroups(teamId);
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
