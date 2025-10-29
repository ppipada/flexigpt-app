import { ElementApi, KEYS, NodeApi } from 'platejs';
import type { PlateEditor, usePlateEditor } from 'platejs/react';

import type { PreProcessorCall, PromptTemplate, PromptVariable } from '@/spec/prompt';

import { expandTabsToSpaces } from '@/lib/text_utils';

import { dispatchTemplateVarsUpdated } from '@/chats/events/template_toolbar_vars_updated';
import {
	buildInitialToolStates,
	computeEffectiveTemplate,
	computeRequirements,
	effectiveVarValueLocal,
	makeSelectedTemplateForRun,
} from '@/chats/templates/template_processing';
import type { SelectedTemplateForRun, TemplateSelectionElementNode, ToolState } from '@/chats/templates/template_spec';
import {
	KEY_TEMPLATE_SELECTION,
	KEY_TEMPLATE_VARIABLE,
	type TemplateVariableElementNode,
	ToolStatus,
} from '@/chats/templates/template_spec';
import { runPreprocessor } from '@/chats/templates/template_tool_processing';

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
	const nnode: TemplateSelectionElementNode = {
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
	// Build per-selection effective context (defs + overrides + tools) so we can resolve values consistently
	const selections = getTemplateNodesWithPath(editor);
	const ctxBySelection = new Map<
		string,
		{
			defsByName: Map<string, PromptVariable>;
			userValues: Record<string, unknown>;
			preProcessors: PreProcessorCall[];
			toolStates?: Record<string, ToolState>;
		}
	>();

	for (const [node] of selections) {
		if (!node.selectionID) continue;
		const { variablesSchema, preProcessors } = computeEffectiveTemplate(node);
		ctxBySelection.set(node.selectionID, {
			defsByName: new Map(variablesSchema.map(v => [v.name, v] as const)),
			userValues: node.variables,
			preProcessors,
			toolStates: node.toolStates,
		});
	}
	function toStringDeepWithVars(n: any): string {
		if (!n || typeof n !== 'object' || n === null) return '';

		if (isTemplateVarNode(n)) {
			const name = n.name;
			const sid = n.selectionID as string | undefined;
			const placeholder = `{{${name}}}`;

			if (!sid) return placeholder;
			const ctx = ctxBySelection.get(sid);
			if (!ctx) return placeholder;

			const def = ctx.defsByName.get(name);
			if (!def) return placeholder; // unknown var (shouldn't happen)

			// Resolve the effective value like the inline pill does
			const val = effectiveVarValueLocal(def, ctx.userValues, ctx.toolStates, ctx.preProcessors);
			if (val !== undefined && val !== null) {
				// eslint-disable-next-line @typescript-eslint/no-base-to-string
				return String(val);
			}
			// If variable is optional and still unresolved, substitute empty string
			if (!def.required) {
				return '';
			}
			// For required or unknown vars, keep the placeholder to signal missing data
			return placeholder;
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

function patchToolState(editor: PlateEditor, tsPath: any, toolId: string, patch: Partial<ToolState>) {
	const nnode = NodeApi.get(editor, tsPath);
	if (!nnode) return;
	const el = nnode as unknown as TemplateSelectionElementNode;
	const prev = el.toolStates ?? {};
	const nextTool = { ...(prev[toolId] ?? {}), ...patch };
	const next = { ...prev, [toolId]: nextTool };
	editor.tf.setNodes({ toolStates: next }, { at: tsPath });
}

function getResolvedToolEntryForSelection(
	tsenode: TemplateSelectionElementNode,
	toolId: string
): {
	args: Record<string, unknown>;
	unresolved: string[];
	status: ToolState['status'];
} {
	const { variablesSchema, preProcessors } = computeEffectiveTemplate(tsenode);
	const req = computeRequirements(variablesSchema, tsenode.variables, preProcessors, tsenode.toolStates);
	const t = req.toolsToRun.find(tt => tt.id === toolId);
	return {
		args: t?.args as Record<string, unknown>,
		unresolved: t?.unresolved ?? [],
		status: t?.status as ToolState['status'],
	};
}

export async function runPreprocessorForSelection(
	editor: PlateEditor,
	tsenode: TemplateSelectionElementNode,
	tsPath: any,
	preproc: PreProcessorCall
): Promise<{ ok: boolean; error?: string }> {
	const { args, unresolved } = getResolvedToolEntryForSelection(tsenode, preproc.id);
	if (unresolved.length > 0) {
		return { ok: false, error: `Missing values: ${unresolved.join(', ')}` };
	}

	// Guard against duplicate clicks
	const st = tsenode.toolStates?.[preproc.id];
	if (st?.status === ToolStatus.RUNNING) return { ok: true };

	patchToolState(editor, tsPath, preproc.id, {
		status: ToolStatus.RUNNING,
		error: undefined,
		lastRunAt: new Date().toISOString(),
	});

	try {
		const result = await runPreprocessor(
			{
				toolBundleID: preproc.toolBundleID,
				toolSlug: preproc.toolSlug,
				toolVersion: preproc.toolVersion,
			},
			args
		);
		// Save result and mark done
		patchToolState(editor, tsPath, preproc.id, {
			status: ToolStatus.DONE,
			result,
			error: undefined,
		});
		if (tsenode.selectionID) {
			dispatchTemplateVarsUpdated(tsenode.selectionID);
		}
		return { ok: true };
	} catch (e) {
		const errMsg =
			e && typeof e === 'object' && 'message' in e ? (e.message as string | undefined) : 'Tool execution failed';
		patchToolState(editor, tsPath, preproc.id, {
			status: ToolStatus.ERROR,
			error: errMsg,
		});
		if (tsenode.selectionID) {
			dispatchTemplateVarsUpdated(tsenode.selectionID);
		}
		return { ok: false, error: errMsg };
	}
}

async function runReadyPreprocessorsForSelection(
	editor: PlateEditor,
	tsenode: TemplateSelectionElementNode,
	tsPath: any
): Promise<{ ok: boolean; errors: Array<{ preprocId: string; saveAs: string; error: string }> }> {
	const { preProcessors, variablesSchema } = computeEffectiveTemplate(tsenode);
	const req = computeRequirements(variablesSchema, tsenode.variables, preProcessors, tsenode.toolStates);

	const ready = req.toolsToRun.filter(t => t.status === ToolStatus.READY || t.status === ToolStatus.ERROR);
	const errors: Array<{ preprocId: string; saveAs: string; error: string }> = [];

	for (const t of ready) {
		const p = preProcessors.find(pp => pp.id === t.id);
		if (!p) continue;
		const r = await runPreprocessorForSelection(editor, tsenode, tsPath, p);
		if (!r.ok) {
			errors.push({ preprocId: p.id, saveAs: p.saveAs, error: r.error ?? 'Tool execution failed' });
		}
	}

	return { ok: errors.length === 0, errors };
}

export async function runAllReadyPreprocessors(editor: PlateEditor): Promise<{
	ok: boolean;
	errors: Array<{ selectionID?: string; preprocId: string; saveAs: string; error: string }>;
}> {
	const items = getTemplateNodesWithPath(editor);
	const allErrors: Array<{ selectionID?: string; preprocId: string; saveAs: string; error: string }> = [];

	for (const [tsenode, tsPath] of items) {
		const r = await runReadyPreprocessorsForSelection(editor, tsenode, tsPath);
		if (!r.ok) {
			for (const e of r.errors) {
				allErrors.push({ selectionID: tsenode.selectionID, ...e });
			}
		}
	}
	return { ok: allErrors.length === 0, errors: allErrors };
}
