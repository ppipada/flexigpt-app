import { DOCUMENT_COLLECTION_INVOKE_CHAR, PROMPT_TEMPLATE_INVOKE_CHAR, TOOL_INVOKE_CHAR } from '@/spec/command';

interface ShortcutConfig {
	key: string;
	ctrlOrMeta?: boolean;
	alt?: boolean;
	shift?: boolean;
}

type ShortcutKeyEvent = Pick<KeyboardEvent, 'key' | 'ctrlKey' | 'metaKey' | 'altKey' | 'shiftKey'>;

export const EDITOR_SHORTCUTS = {
	templates: { key: PROMPT_TEMPLATE_INVOKE_CHAR, ctrlOrMeta: true } satisfies ShortcutConfig,
	tools: { key: TOOL_INVOKE_CHAR, ctrlOrMeta: true } satisfies ShortcutConfig,
	attachments: {
		key: DOCUMENT_COLLECTION_INVOKE_CHAR,
		ctrlOrMeta: true,
	} satisfies ShortcutConfig,
};

export function matchShortcut(event: ShortcutKeyEvent, shortcut: ShortcutConfig): boolean {
	if (shortcut.ctrlOrMeta && !(event.metaKey || event.ctrlKey)) return false;
	if (!shortcut.ctrlOrMeta && (event.metaKey || event.ctrlKey)) return false;
	if (shortcut.alt !== undefined && shortcut.alt !== event.altKey) return false;
	if (shortcut.shift !== undefined && shortcut.shift !== event.shiftKey) return false;
	return event.key === shortcut.key;
}

export function formatShortcut(shortcut: ShortcutConfig): string {
	const parts: string[] = [];
	if (shortcut.ctrlOrMeta) parts.push('⌘/Ctrl');
	if (shortcut.alt) parts.push('Alt');
	if (shortcut.shift) parts.push('Shift');
	parts.push("'" + shortcut.key + "'");
	return parts.join(' + ');
}
