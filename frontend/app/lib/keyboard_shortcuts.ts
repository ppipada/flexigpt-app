import { useEffect, useRef } from 'react';

import { MOD_LABEL } from '@/lib/features';

interface ShortcutChord {
	key: string; // e.g. 'k'
	ctrlOrMeta?: boolean; // true = requires Ctrl or Cmd
	alt?: boolean;
	shift?: boolean;
}

enum ShortcutAction {
	newChat = 'newChat',
	focusSearch = 'focusSearch',
	focusInput = 'focusInput',
	insertTemplate = 'insertTemplate',
	insertTool = 'insertTool',
	insertAttachment = 'insertAttachment',
}

interface ShortcutKeyEvent {
	key: string;
	ctrlKey: boolean;
	metaKey: boolean;
	altKey: boolean;
	shiftKey: boolean;
}

export type ShortcutConfig = Partial<Record<ShortcutAction, ShortcutChord>>;

interface ShortcutDisplayItem {
	action: ShortcutAction;
	/** Short, human-friendly action name */
	label: string;
	/** Formatted key combo, e.g. "⌘ + Shift + 'P'" */
	keys: string;
	/** Optional grouping if you ever want to separate them in the UI */
	group: 'Chat' | 'Insert';
	/** For stable ordering */
	order: number;
}

// Default mappings; you can later load these from settings
export const defaultShortcutConfig: ShortcutConfig = {
	// Global-chat actions
	newChat: { key: 'n', ctrlOrMeta: true, shift: true }, // Mod+Shift+N
	focusSearch: { key: 'f', ctrlOrMeta: true, shift: true }, // Mod+Shift+F
	focusInput: { key: 'i', ctrlOrMeta: true, shift: true }, // Mod+Shift+I

	// Editor/insert actions
	insertTemplate: { key: 'p', ctrlOrMeta: true, shift: true }, // Mod+Shift+P
	insertTool: { key: 't', ctrlOrMeta: true, shift: true }, // Mod+Shift+T
	insertAttachment: { key: 'a', ctrlOrMeta: true, shift: true }, // Mod+Shift+A
};

const ACTION_META: Record<ShortcutAction, Omit<ShortcutDisplayItem, 'action' | 'keys'>> = {
	[ShortcutAction.newChat]: {
		label: 'New chat',
		group: 'Chat',
		order: 10,
	},
	[ShortcutAction.focusSearch]: {
		label: 'Focus search',
		group: 'Chat',
		order: 20,
	},
	[ShortcutAction.focusInput]: {
		label: 'Focus input',
		group: 'Chat',
		order: 30,
	},
	[ShortcutAction.insertTemplate]: {
		label: 'Insert template',
		group: 'Insert',
		order: 40,
	},
	[ShortcutAction.insertTool]: {
		label: 'Add tool',
		group: 'Insert',
		order: 50,
	},
	[ShortcutAction.insertAttachment]: {
		label: 'Attach context',
		group: 'Insert',
		order: 60,
	},
};

interface UseShortcutsOptions {
	config: ShortcutConfig;
	isBusy: boolean;
	handlers: Partial<Record<ShortcutAction, () => void>>;
}

function matchShortcut(event: ShortcutKeyEvent, shortcut?: ShortcutChord): boolean {
	if (!shortcut) return false;
	if (shortcut.ctrlOrMeta && !(event.ctrlKey || event.metaKey)) return false;
	if (!shortcut.ctrlOrMeta && (event.ctrlKey || event.metaKey)) return false;
	if (shortcut.alt !== undefined && shortcut.alt !== event.altKey) return false;
	if (shortcut.shift !== undefined && shortcut.shift !== event.shiftKey) return false;
	return event.key.toLowerCase() === shortcut.key.toLowerCase();
}

export function formatShortcut(shortcut?: ShortcutChord): string {
	if (!shortcut) return '';
	const parts: string[] = [];
	if (shortcut.ctrlOrMeta) parts.push(MOD_LABEL);
	if (shortcut.alt) parts.push('Alt');
	if (shortcut.shift) parts.push('Shift');
	parts.push(`'${shortcut.key}'`);
	return parts.join(' + ');
}

/**
 * Single global shortcuts hook. Attach once at the chat page level.
 * It uses the config map so you can swap mappings in the future.
 */
export function useChatShortcuts({ config, isBusy, handlers }: UseShortcutsOptions) {
	const ref = useRef({ config, isBusy, handlers });

	useEffect(() => {
		ref.current = { config, isBusy, handlers };
	}, [config, isBusy, handlers]);

	useEffect(() => {
		const onKeyDown = (event: KeyboardEvent) => {
			if (event.defaultPrevented) return;

			const { config, isBusy, handlers } = ref.current;

			const entries = Object.entries(config) as [ShortcutAction, ShortcutChord][];

			for (const [action, chord] of entries) {
				const handler = handlers[action];
				if (!handler) continue;

				if (!matchShortcut(event, chord)) continue;

				// Example of busy-sensitive behavior: no new chat while busy
				if (isBusy && action === ShortcutAction.newChat) return;

				event.preventDefault();
				event.stopPropagation();

				handler();
				return;
			}
		};

		window.addEventListener('keydown', onKeyDown, true);
		return () => {
			window.removeEventListener('keydown', onKeyDown, true);
		};
	}, []);
}

/**
 * Build a flat list of display items for all shortcuts that are configured.
 * This is what UI components should consume.
 */
export function buildShortcutDisplay(config: ShortcutConfig): ShortcutDisplayItem[] {
	const items: ShortcutDisplayItem[] = [];

	(Object.entries(ACTION_META) as [ShortcutAction, (typeof ACTION_META)[ShortcutAction]][]).forEach(
		([action, meta]) => {
			const chord = config[action];
			if (!chord) return;
			const keys = formatShortcut(chord);
			if (!keys) return;

			items.push({
				action,
				label: meta.label,
				group: meta.group,
				order: meta.order,
				keys,
			});
		}
	);

	// Stable order
	items.sort((a, b) => a.order - b.order);
	return items;
}
