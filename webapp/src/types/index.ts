export interface Group {
    id: string;
    name: string;
    team_id: string;
    visibility: 'public' | 'private';
    owners: string[];
    members: string[];
    created_at: string;
    updated_at: string;
}

export interface AutocompleteResult {
    name: string;
    display_name: string;
    description: string;
    visibility: string;
    member_count: number;
}

export interface PluginRegistry {
    registerPostTypeComponent(typeName: string, component: React.ComponentType): void;
    registerMessageWillFormatHook(hook: (post: any, message: string) => string): void;
    registerSlashCommandWillBePostedHook(hook: (message: string, args: any) => Promise<any>): void;
    unregisterComponent(componentId: string): void;
    registerPopoverUserAttributesHook?: (hook: (user: any) => any) => string;
    registerPopoverUserActionsHook?: (hook: (user: any) => any) => string;
    registerRootComponent?: (component: React.ComponentType) => string;
}

export interface PluginClass {
    initialize(registry: PluginRegistry, store: any): void;
    uninitialize(): void;
}

declare global {
    interface Window {
        registerPlugin(id: string, plugin: PluginClass): void;
    }
}
