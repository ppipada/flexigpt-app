import type { ToolCall, ToolChoice } from '@/spec/tool';

/**
 * Status for a tool call chip in the composer.
 */
export type ToolCallStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'discarded';

export interface ToolCallChip {
	id: string;
	callID: string;
	name: string;
	arguments: string;
	type: string;
	status: ToolCallStatus;
	errorMessage?: string;
	/** Optional original ToolChoice associated with this call, if the provider supplied it. */
	toolChoice?: ToolChoice;
}

/**
 * Build initial tool-call chip state from the provider's toolCalls array.
 */
export function buildToolCallChipsFromResponse(toolCalls: ToolCall[] | undefined | null): ToolCallChip[] {
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
export function summarizeToolCallArguments(args: string): string | undefined {
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
export function formatToolCallChipLabel(call: ToolCallChip): string {
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
