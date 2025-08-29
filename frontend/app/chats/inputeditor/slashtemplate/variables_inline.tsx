import React from 'react';

import { NodeApi } from 'platejs';
import type { PlateEditor, PlateElementProps } from 'platejs/react';
import { FiEdit2 } from 'react-icons/fi';

import { TemplateEditModal } from '@/chats/inputeditor/slashtemplate/template_edit_modal';
import {
	computeEffectiveTemplate,
	computeRequirements,
	type TemplateSelectionElementNode,
} from '@/chats/inputeditor/slashtemplate/template_processing';
import {
	getFirstTemplateNodeWithPath,
	KEY_TEMPLATE_SELECTION,
} from '@/chats/inputeditor/slashtemplate/template_selection_element';

export const KEY_TEMPLATE_VARIABLE = 'templateVariable';

type TemplateVariableElementNode = {
	type: typeof KEY_TEMPLATE_VARIABLE;
	bundleID: string;
	templateSlug: string;
	name: string;
	// for layout only (computed again at render)
	required?: boolean;
	children: [{ text: '' }];
};

// Inline variable pill renderer
export function TemplateVariableElement(props: PlateElementProps<any>) {
	const { element, attributes, children, editor } = props as any;
	const el = element as TemplateVariableElementNode;
	const plEditor = editor as PlateEditor;

	const tpl = findTemplateNode(plEditor, el.bundleID, el.templateSlug);
	const [tsenode, tsPath] = tpl ?? [];
	const { variablesSchema } = tsenode ? computeEffectiveTemplate(tsenode) : { variablesSchema: [] };
	const varDef = variablesSchema.find(v => v.name === el.name);
	const isRequired = Boolean(varDef?.required);

	// Compute requirements for status color
	const req = tsenode
		? computeRequirements(
				variablesSchema,
				tsenode.variables,
				tsenode.overrides?.preProcessors ?? [],
				tsenode.toolStates
			)
		: { variableValues: {}, requiredVariables: [] as string[], requiredCount: 0 };

	const isMissing = isRequired && req.requiredVariables.includes(el.name);

	const [open, setOpen] = React.useState(false);

	return (
		<span
			{...attributes}
			contentEditable={false}
			tabIndex={0}
			data-template-variable
			data-var-name={el.name}
			data-state={isMissing ? 'required' : 'ready'}
			className={`badge inline-flex items-center gap-1 whitespace-nowrap select-none ${
				isMissing ? 'badge-warning' : 'badge-success'
			}`}
			title={isMissing ? `Required: ${el.name}` : `Variable: ${el.name}`}
			onKeyDown={e => {
				if (e.key === 'Enter') {
					e.preventDefault();
					e.stopPropagation();
					setOpen(true);
				}
			}}
			onMouseDown={e => {
				// allow focusing pill without bubbling into editor selection changes
				e.preventDefault();
			}}
			onClick={() => {
				setOpen(true);
			}}
		>
			<span className="flex items-center gap-1">
				<span className="font-medium">{el.name}</span>
				<FiEdit2 className="opacity-70" size={12} />
			</span>

			{/* Modal for power edit, scoped to this template */}
			{tsenode ? (
				<TemplateEditModal
					open={open}
					onClose={() => {
						setOpen(false);
					}}
					tsenode={tsenode}
					editor={plEditor}
					path={tsPath}
				/>
			) : null}

			{children}
		</span>
	);
}

/**
 * Build Slate inline children from a plain text that may include {{varName}} tokens.
 * Unknown variables are left as plain text.
 */
export function buildUserInlineChildrenFromText(tsenode: TemplateSelectionElementNode, text: string): any[] {
	const { variablesSchema } = computeEffectiveTemplate(tsenode);
	const known = new Set(variablesSchema.map(v => v.name));

	const result: any[] = [];
	const re = /\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g;
	let idx = 0;
	let m: RegExpExecArray | null;
	while ((m = re.exec(text)) !== null) {
		const pre = text.slice(idx, m.index);
		if (pre) result.push({ text: pre });

		const varName = m[1];
		if (known.has(varName)) {
			const node: TemplateVariableElementNode = {
				type: KEY_TEMPLATE_VARIABLE,
				bundleID: tsenode.bundleID,
				templateSlug: tsenode.templateSlug,
				name: varName,
				required: variablesSchema.find(v => v.name === varName)?.required ?? false,
				children: [{ text: '' }],
			};
			result.push(node as any);
		} else {
			// unknown variable -> keep as literal
			result.push({ text: m[0] });
		}

		idx = m.index + m[0].length;
	}
	const tail = text.slice(idx);
	if (tail) result.push({ text: tail });

	if (result.length === 0) result.push({ text: '' });
	return result;
}

/**
 * Flatten current editor content into plain text, replacing variable pills with value (or {{name}} if not set).
 */
export function toPlainTextReplacingVariables(editor: PlateEditor): string {
	// Find the first selection node for context (variables and values)
	const tpl = getFirstTemplateNodeWithPath(editor);
	const [tsenode] = tpl ?? [];
	const vars = tsenode ? tsenode.variables : {};

	// Walk the top-level single paragraph
	const childnodes = (editor.children[0]?.children ?? []) as any[];
	const parts: string[] = [];

	childnodes.forEach(n => {
		if (isTemplateVarNode(n)) {
			const name = n.name;
			const v = vars[name];
			if (v !== undefined && v !== null && v !== '') {
				// eslint-disable-next-line @typescript-eslint/no-base-to-string
				parts.push(String(v));
			} else {
				parts.push(`{{${name}}}`);
			}
		} else {
			// Fallback to string content of nested children
			const s = toStringDeep(n);
			if (s) parts.push(s);
		}
	});

	return parts.join('');
}

function toStringDeep(n: any): string {
	// 1️⃣ be sure we really have a non-null object first
	if (!n || typeof n !== 'object' || n === null) return '';

	// 2️⃣ from now on we can treat it as a generic dictionary
	const obj = n as Record<PropertyKey, unknown>;

	if ('text' in obj && typeof obj.text === 'string') {
		return obj.text;
	}

	if ('children' in obj && Array.isArray(obj.children)) {
		return obj.children.map(toStringDeep).join('');
	}
	return '';
}

function isTemplateVarNode(n: unknown): n is TemplateVariableElementNode {
	// 1️⃣ be sure we really have a non-null object first
	if (!n || typeof n !== 'object') return false;

	// 2️⃣ from now on we can treat it as a generic dictionary
	const obj = n as Record<PropertyKey, unknown>;

	// 3️⃣ check that the expected keys exist and have the right runtime shape
	return 'type' in obj && obj.type === KEY_TEMPLATE_VARIABLE && 'name' in obj && typeof obj.name === 'string';
}

function findTemplateNode(
	editor: PlateEditor,
	bundleID: string,
	templateSlug: string
): [TemplateSelectionElementNode, any] | undefined {
	const els = NodeApi.elements(editor);
	for (const [el, path] of els) {
		if (el.type === KEY_TEMPLATE_SELECTION) {
			const node = el as unknown as TemplateSelectionElementNode;
			if (node.bundleID === bundleID && node.templateSlug === templateSlug) {
				return [node, path];
			}
		}
	}
	return undefined;
}
