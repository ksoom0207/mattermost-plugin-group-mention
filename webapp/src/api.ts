import {Group, AutocompleteResult} from './types';

const PLUGIN_ID = 'com.mattermost.plugin-group-mention';

export class Client {
    private baseUrl: string;

    constructor() {
        this.baseUrl = `/plugins/${PLUGIN_ID}/api`;
    }

    async getGroups(teamId: string): Promise<Group[]> {
        const url = `${this.baseUrl}/groups?team_id=${teamId}`;
        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
        });

        if (!response.ok) {
            throw new Error(`Failed to fetch groups: ${response.statusText}`);
        }

        return response.json();
    }

    async autocompleteGroups(teamId: string, query: string): Promise<AutocompleteResult[]> {
        const url = `${this.baseUrl}/groups/autocomplete?team_id=${teamId}&q=${encodeURIComponent(query)}`;
        const response = await fetch(url, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            },
            credentials: 'include',
        });

        if (!response.ok) {
            throw new Error(`Failed to autocomplete groups: ${response.statusText}`);
        }

        return response.json();
    }
}

export default new Client();
