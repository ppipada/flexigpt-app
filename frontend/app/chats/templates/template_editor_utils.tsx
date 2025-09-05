import { ElementApi, KEYS, NodeApi } from 'platejs';
import type { PlateEditor, usePlateEditor } from 'platejs/react';

import type { PromptTemplate } from '@/spec/prompt';

import { expandTabsToSpaces } from '@/lib/text_utils';

import { buildInitialToolStates, makeSelectedTemplateForRun } from '@/chats/templates/template_processing';
import type { SelectedTemplateForRun, TemplateSelectionElementNode } from '@/chats/templates/template_spec';
import {
	KEY_TEMPLATE_SELECTION,
	KEY_TEMPLATE_VARIABLE,
	type TemplateVariableElementNode,
} from '@/chats/templates/template_spec';

export function insertTemplateSelectionNode(
	editor: PlateEditor,
	bundleID: string,
	templateSlug: string,
	templateVersion: string,
	template?: PromptTemplate
) {
	const selectionID = `tpl:${bundleID}/${templateSlug}@${templateVersion}:${Date.now().toString(36)}${Math.random()
		.toString(36)
		.slice(2, 8)}`;
	const nnode = {
		type: KEY_TEMPLATE_SELECTION,
		bundleID,
		templateSlug,
		templateVersion,
		selectionID,
		variables: {} as Record<string, unknown>,
		// Snapshot full template for downstream sync "get" to have the full context.
		templateSnapshot: template,
		// Local overrides
		overrides: {} as {
			displayName?: string;
			description?: string;
			tags?: string[];
			blocks?: PromptTemplate['blocks'];
			variables?: PromptTemplate['variables'];
			preProcessors?: PromptTemplate['preProcessors'];
		},
		// Each preprocessor call state
		toolStates: buildInitialToolStates(template),
		// void elements still need one empty text child in Slate
		children: [{ text: '' }],
	};

	editor.tf.withoutNormalizing(() => {
		// Insert the chip (inline+void)
		editor.tf.insertNodes([nnode, { type: KEYS.p, text: '\n' }], { select: true });
		// Move caret after the chip and add a trailing space so the user can keep typing
		editor.tf.collapse({ edge: 'end' });
		editor.tf.select(undefined, { edge: 'end' }); // Select end of block above
	});
	editor.tf.focus();
}

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

export function hasNonEmptyUserText(ed: PlateEditor | null | undefined): boolean {
	if (!ed) return false;
	// If NodeApi.texts exists:
	for (const [t] of NodeApi.texts(ed)) {
		if (t.text.trim().length > 0) return true;
	}
	return false;
}
