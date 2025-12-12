import { ElementApi, KEYS, NodeApi, type Path } from 'platejs';
import type { PlateEditor } from 'platejs/react';

import type { Tool, ToolCall, ToolChoice } from '@/spec/tool';

// Keys for the tools combobox and inline elements
export const KEY_TOOL_SELECTION = 'toolSelection';
// export const KEY_TOOL_PLUS_COMMAND = 'toolPlus';
// const KEY_TOOL_PLUS_INPUT = 'toolPlusInput';

export type ToolSelectionElementNode = {
	type: typeof KEY_TOOL_SELECTION;
	bundleID: string;
	bundleSlug?: string;
	toolSlug: string;
	toolVersion: string;
	selectionID: string;

	toolSnapshot?: Tool;
	overrides?: {
		displayName?: string;
		description?: string;
		tags?: string[];
	};

	// inline+void node needs a text child
	children: [{ text: '' }];
};

/**
 * Status for a tool call chip in the composer.
 */
type EditorToolCallStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'discarded';

export interface EditorToolCall {
	id: string;
	callID: string;
	name: string;
	arguments: string;
	type: string;
	status: EditorToolCallStatus;
	errorMessage?: string;
	/** Optional original ToolChoice associated with this call, if the provider supplied it. */
	toolChoice?: ToolChoice;
}

/**
 * Build initial tool-call chip state from the provider's toolCalls array.
 */
export function buildToolCallFromResponse(toolCalls: ToolCall[] | undefined | null): EditorToolCall[] {
	if (!toolCalls || toolCalls.length === 0) return [];
	return toolCalls.map(tc => ({
		id: tc.id || tc.callID,
		callID: tc.callID,
		name: tc.name,
		arguments: tc.arguments,
		type: tc.type,
		status: 'pending',
		toolChoice: tc.toolChoice,
	}));
}

/**
 * Human-friendly tool name for display.
 * Accepts forms like:
 *   "bundleSlug/toolSlug@version"
 *   "bundleID/toolSlug@version"
 *   "toolSlug"
 */
export function getPrettyToolName(name: string): string {
	if (!name) return 'Tool';
	let base = name;
	if (base.includes('/')) {
		const parts = base.split('/');
		base = parts[parts.length - 1] || base;
	}
	if (base.includes('@')) {
		base = base.split('@')[0] || base;
	}
	return base.replace(/[-_]/g, ' ');
}

/**
 * Best-effort short summary of tool-call arguments for chip labels.
 */
function summarizeToolCallArguments(args: string): string | undefined {
	if (!args) return undefined;
	try {
		const parsed = JSON.parse(args);
		if (parsed == null || typeof parsed !== 'object') {
			return typeof parsed === 'string' ? parsed : undefined;
		}
		const obj = parsed as Record<string, unknown>;
		const primaryKeys = ['file', 'path', 'url', 'query', 'id', 'name'];
		const parts: string[] = [];

		for (const key of primaryKeys) {
			if (obj[key] != null) {
				// eslint-disable-next-line @typescript-eslint/no-base-to-string
				parts.push(String(obj[key]));
			}
		}

		if (parts.length === 0) {
			const keys = Object.keys(obj);
			for (const key of keys.slice(0, 2)) {
				parts.push(`${key}=${String(obj[key])}`);
			}
		}

		return parts.length ? parts.join(', ') : undefined;
	} catch {
		return undefined;
	}
}

/**
 * Label used for tool-call chips in composer / history.
 */
export function formatToolCallLabel(call: EditorToolCall): string {
	const pretty = getPrettyToolName(call.name);
	const argSummary = summarizeToolCallArguments(call.arguments);
	return argSummary ? `${pretty}: ${argSummary}` : pretty;
}

/**
 * Default summary label for a tool-output chip.
 */
export function formatToolOutputSummary(name: string): string {
	const pretty = getPrettyToolName(name);
	return `Result: ${pretty}`;
}

export type EditorAttachedToolChoice = {
	bundleID: string;
	bundleSlug?: string;

	toolID?: string;
	toolSlug: string;
	toolVersion: string;

	displayName?: string;
	description?: string;

	// Non ToolChoice fields
	selectionID: string;
};

// Build a stable identity key for a tool selection (bundle + slug + version).
// Prefer bundleID when present, otherwise fall back to bundleSlug.
export function toolIdentityKey(
	bundleID: string | undefined,
	bundleSlug: string | undefined,
	toolSlug: string,
	toolVersion: string
): string {
	const bundlePart = bundleID ? `id:${bundleID}` : `slug:${bundleSlug ?? ''}`;
	return `${bundlePart}/${toolSlug}@${toolVersion}`;
}

function toolIdentityKeyFromNode(
	n: Pick<ToolSelectionElementNode, 'bundleID' | 'bundleSlug' | 'toolSlug' | 'toolVersion'>
): string {
	return toolIdentityKey(n.bundleID, n.bundleSlug, n.toolSlug, n.toolVersion);
}

function getAttachedToolKeySet(editor: PlateEditor): Set<string> {
	const keys = new Set<string>();
	for (const [el] of NodeApi.elements(editor)) {
		if (ElementApi.isElementType(el, KEY_TOOL_SELECTION)) {
			keys.add(toolIdentityKeyFromNode(el as unknown as ToolSelectionElementNode));
		}
	}
	return keys;
}

// Insert a hidden tool selection chip (inline+void) to drive bottom bar UI.
export function insertToolSelectionNode(
	editor: PlateEditor,
	item: {
		bundleID: string;
		bundleSlug?: string;
		toolSlug: string;
		toolVersion: string;
	},
	toolSnapshot?: Tool
) {
	const identity = toolIdentityKey(item.bundleID, item.bundleSlug, item.toolSlug, item.toolVersion);
	if (getAttachedToolKeySet(editor).has(identity)) {
		editor.tf.focus();
		return;
	}

	const selectionID = `tool:${item.bundleID}/${item.toolSlug}@${item.toolVersion}:${Date.now().toString(36)}${Math.random()
		.toString(36)
		.slice(2, 8)}`;

	const node: ToolSelectionElementNode = {
		type: KEY_TOOL_SELECTION,
		bundleID: item.bundleID,
		bundleSlug: item.bundleSlug,
		toolSlug: item.toolSlug,
		toolVersion: item.toolVersion,
		selectionID,
		toolSnapshot,
		overrides: {},
		children: [{ text: '' }],
	};

	editor.tf.withoutNormalizing(() => {
		// Insert the tool chip (invisible inline) followed by a paragraph separator for smoother typing.
		editor.tf.insertNodes([node, { type: KEYS.p, text: '\n' }], { select: true });
		editor.tf.collapse({ edge: 'end' });
		editor.tf.select(undefined, { edge: 'end' });
	});

	editor.tf.focus();
}

// List tool nodes with path in document order.
// By default, returns only the first occurrence for each unique tool identity.
export function getToolNodesWithPath(
	editor: PlateEditor,
	uniqueByIdentity?: boolean
): Array<[ToolSelectionElementNode, Path]> {
	const out: Array<[ToolSelectionElementNode, Path]> = [];
	const unique = uniqueByIdentity ?? true;
	const seen = unique ? new Set<string>() : undefined;
	for (const [el, path] of NodeApi.elements(editor)) {
		if (ElementApi.isElementType(el, KEY_TOOL_SELECTION)) {
			const n = el as unknown as ToolSelectionElementNode;
			if (unique) {
				const key = toolIdentityKeyFromNode(n);
				if ((seen as Set<string>).has(key)) continue;
				(seen as Set<string>).add(key);
			}
			out.push([n, path]);
		}
	}
	return out;
}

// Remove all instances of a tool by identity key (bundle+slug+version).
export function removeToolByKey(editor: PlateEditor, identityKey: string) {
	const paths: Path[] = [];
	for (const [el, path] of NodeApi.elements(editor)) {
		if (ElementApi.isElementType(el, KEY_TOOL_SELECTION)) {
			const n = el as unknown as ToolSelectionElementNode;
			if (toolIdentityKeyFromNode(n) === identityKey) {
				paths.push(path);
			}
		}
	}
	// Remove from last to first to avoid path shift issues.
	for (const p of paths.reverse()) {
		try {
			editor.tf.removeNodes({ at: p });
		} catch {
			// swallow
		}
	}
	editor.tf.focus();
}

// Build a serializable list of attached tools for submission
export function getAttachedTools(editor: PlateEditor): EditorAttachedToolChoice[] {
	const items: EditorAttachedToolChoice[] = [];
	const seen = new Set<string>();

	for (const [el] of NodeApi.elements(editor)) {
		if (ElementApi.isElementType(el, KEY_TOOL_SELECTION)) {
			const n = el as unknown as ToolSelectionElementNode;
			const key = toolIdentityKeyFromNode(n);
			if (seen.has(key)) continue;
			seen.add(key);
			items.push({
				selectionID: n.selectionID,
				bundleID: n.bundleID,
				toolSlug: n.toolSlug,
				toolVersion: n.toolVersion,
				displayName: n.overrides?.displayName
					? n.overrides.displayName
					: n.toolSnapshot?.displayName
						? n.toolSnapshot.displayName
						: n.toolSlug,
				description: n.overrides?.description
					? n.overrides.description
					: n.toolSnapshot?.description
						? n.toolSnapshot.description
						: n.toolSlug,
				toolID: n.toolSnapshot?.id,
			});
		}
	}
	return items;
}
