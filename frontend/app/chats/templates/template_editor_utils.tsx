import { ElementApi, NodeApi, type Value } from 'platejs';
import type { PlateEditor, usePlateEditor } from 'platejs/react';

import { expandTabsToSpaces } from '@/lib/text_utils';

import type { SelectedTemplateForRun, TemplateSelectionElementNode } from '@/chats/templates/template_processing';
import { makeSelectedTemplateForRun } from '@/chats/templates/template_processing';

export const KEY_TEMPLATE_SELECTION = 'templateSelection';
export const KEY_TEMPLATE_VARIABLE = 'templateVariable';
export const KEY_TEMPLATE_SLASH_COMMAND = 'templateSlash';
export const KEY_TEMPLATE_SLASH_INPUT = 'templateInput';

export const EMPTY_VALUE: Value = [{ type: 'p', children: [{ text: '' }] }];

export type TemplateVariableElementNode = {
	type: typeof KEY_TEMPLATE_VARIABLE;
	bundleID: string;
	templateSlug: string;
	templateVersion: string;
	selectionID: string;
	name: string;
	// for layout only (computed again at render)
	required?: boolean;
	children: [{ text: '' }];
};

// Utility to get selections for sending
export function getTemplateSelections(editor: PlateEditor): SelectedTemplateForRun[] {
	const elList = NodeApi.elements(editor);
	const selections: SelectedTemplateForRun[] = [];
	for (const [el, _path] of elList) {
		if (ElementApi.isElementType(el, KEY_TEMPLATE_SELECTION)) {
			const node = el as unknown as TemplateSelectionElementNode;
			selections.push(makeSelectedTemplateForRun(node));
		}
	}

	return selections;
}

// Utility to get the first template node with its path
export function getFirstTemplateNodeWithPath(editor: PlateEditor): [TemplateSelectionElementNode, any] | undefined {
	const elList = NodeApi.elements(editor);
	for (const [el, path] of elList) {
		if (ElementApi.isElementType(el, KEY_TEMPLATE_SELECTION)) {
			return [el as unknown as TemplateSelectionElementNode, path];
		}
	}
	return undefined;
}

// Utility to get all template selection nodes with their paths (document order)
export function getTemplateNodesWithPath(editor: PlateEditor): Array<[TemplateSelectionElementNode, any]> {
	const out: Array<[TemplateSelectionElementNode, any]> = [];
	const elList = NodeApi.elements(editor);
	for (const [el, path] of elList) {
		if (ElementApi.isElementType(el, KEY_TEMPLATE_SELECTION)) {
			out.push([el as unknown as TemplateSelectionElementNode, path]);
		}
	}
	return out;
}

// Flatten current editor content into plain text (single-block), replacing variable pills of the first template.
// Used when extracting text to submit without mutating content.
export function toPlainTextReplacingVariables(editor: PlateEditor): string {
	// Map selectionID -> variables for correct per-chip resolution
	const selections = getTemplateNodesWithPath(editor);
	const varsBySelection = new Map<string, Record<string, unknown>>();
	selections.forEach(([node]) => {
		if (node.selectionID) varsBySelection.set(node.selectionID, node.variables);
	});

	function toStringDeepWithVars(n: any): string {
		if (!n || typeof n !== 'object' || n === null) return '';

		if (isTemplateVarNode(n)) {
			const name = n.name;
			const sid = n.selectionID as string | undefined;
			const vars = ((sid && varsBySelection.get(sid)) ?? {}) as Record<PropertyKey, unknown>;
			let v = `{{${name}}}`;
			if (name in vars && vars[name] !== undefined && vars[name] !== null && vars[name] !== '') {
				// eslint-disable-next-line @typescript-eslint/no-base-to-string
				v = String(vars[name]);
			}
			return v;
		}

		const obj = n as Record<PropertyKey, unknown>;

		if ('text' in obj && typeof obj.text === 'string') {
			return obj.text;
		}

		if ('children' in obj && Array.isArray(obj.children)) {
			return obj.children.map(toStringDeepWithVars).join('');
		}

		return '';
	}

	const childnodes = (editor.children[0]?.children ?? []) as any[];
	return childnodes.map(toStringDeepWithVars).join('');
}

function isTemplateVarNode(n: unknown): n is TemplateVariableElementNode {
	if (!n || typeof n !== 'object') return false;
	const obj = n as Record<PropertyKey, unknown>;
	return 'type' in obj && obj.type === KEY_TEMPLATE_VARIABLE && 'name' in obj && typeof obj.name === 'string';
}

export function insertPlainTextAsSingleBlock(ed: ReturnType<typeof usePlateEditor>, text: string, tabSize = 2) {
	if (!ed) {
		return;
	}
	const editor = ed as PlateEditor;
	const normalized = text.replace(/\r\n?/g, '\n');
	const lines = normalized.split('\n').map(l => expandTabsToSpaces(l, tabSize));

	editor.tf.insertText(lines[0] ?? '');
	for (let i = 1; i < lines.length; i++) {
		editor.tf.insertSoftBreak();
		editor.tf.insertText(lines[i]);
	}
}

// Deeper (longer) paths first; for equal depth, lexicographic descending.
export function comparePathDeepestFirst(a: ReadonlyArray<number>, b: ReadonlyArray<number>): number {
	if (a.length !== b.length) return b.length - a.length;
	for (let i = 0; i < a.length; i++) {
		const da = a[i] ?? 0;
		const db = b[i] ?? 0;
		if (da !== db) return db - da;
	}
	return 0;
}

export function compareEntryByPathDeepestFirst<T extends [unknown, ReadonlyArray<number>]>(a: T, b: T): number {
	return comparePathDeepestFirst(a[1], b[1]);
}

export function hasNonEmptyUserText(ed: PlateEditor | null | undefined): boolean {
	if (!ed) return false;
	// If NodeApi.texts exists:
	for (const [t] of NodeApi.texts(ed)) {
		if (t.text.trim().length > 0) return true;
	}
	return false;
}
