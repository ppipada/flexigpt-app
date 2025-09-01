/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import * as React from 'react';

import { NodeApi } from 'platejs';
import type { PlateEditor, PlateElementProps } from 'platejs/react';
import { FiEdit2 } from 'react-icons/fi';

import type { PromptVariable } from '@/spec/prompt';
import { VarSource, VarType } from '@/spec/prompt';

import {
	computeEffectiveTemplate,
	computeRequirements,
	type TemplateSelectionElementNode,
	type ToolState,
} from '@/chats/inputeditor/slashtemplate/template_processing';
import {
	getFirstTemplateNodeWithPath,
	KEY_TEMPLATE_SELECTION,
} from '@/chats/inputeditor/slashtemplate/template_slash_selection';

export const KEY_TEMPLATE_VARIABLE = 'templateVariable';

type TemplateVariableElementNode = {
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

function useEvent(eventName: string, handler: (e: any) => void) {
	React.useEffect(() => {
		const h = (e: Event) => {
			handler(e);
		};
		window.addEventListener(eventName, h);
		return () => {
			window.removeEventListener(eventName, h);
		};
	}, [eventName, handler]);
}

function findTemplateNode(
	editor: PlateEditor,
	bundleID: string,
	templateSlug: string,
	templateVersion: string,
	selectionID: string
): [TemplateSelectionElementNode, any] | undefined {
	const els = NodeApi.elements(editor);
	for (const [el, path] of els) {
		if (el.type === KEY_TEMPLATE_SELECTION) {
			const node = el as unknown as TemplateSelectionElementNode;
			const nodeSelectionID = node.selectionID as string | undefined;
			const matchesById = selectionID && nodeSelectionID === selectionID;
			const matchesByTriple =
				node.bundleID === bundleID && node.templateSlug === templateSlug && node.templateVersion === templateVersion;
			if (matchesById || matchesByTriple) {
				return [node, path];
			}
		}
	}
	return undefined;
}

function effectiveVarValueLocal(
	varDef: PromptVariable,
	userValues: Record<string, unknown>,
	toolStates?: Record<string, ToolState>
): unknown {
	if (userValues[varDef.name] !== undefined && userValues[varDef.name] !== null) {
		return userValues[varDef.name];
	}
	if (varDef.source === VarSource.Static && varDef.staticVal !== undefined) {
		return varDef.staticVal;
	}
	if (varDef.default !== undefined && varDef.default !== '') {
		return varDef.default;
	}
	if (varDef.source === VarSource.Tool && toolStates) {
		const hit = Object.values(toolStates).find(st => st.result !== undefined);
		if (hit?.result !== undefined) return hit.result;
	}
	return undefined;
}

// Inline variable pill renderer with inline editing
export function TemplateVariableElement(props: PlateElementProps<any>) {
	const { element, attributes, children, editor } = props as any;
	const el = element as TemplateVariableElementNode;
	const plEditor = editor as PlateEditor;

	const tpl = findTemplateNode(plEditor, el.bundleID, el.templateSlug, el.templateVersion, el.selectionID);
	const [tsenode, tsPath] = tpl ?? [];
	const eff = tsenode ? computeEffectiveTemplate(tsenode) : undefined;
	const variablesSchema = eff?.variablesSchema ?? [];

	const varDef = variablesSchema.find(v => v.name === el.name);
	const isRequired = Boolean(varDef?.required);

	// Compute requirements for status color using effective preProcessors (overrides included)
	const req = tsenode
		? computeRequirements(variablesSchema, tsenode.variables, eff?.preProcessors ?? [], tsenode.toolStates)
		: { variableValues: {}, requiredVariables: [] as string[], requiredCount: 0 };

	const isMissing = isRequired && req.requiredVariables.includes(el.name);

	// Local inline edit state
	const [isEditing, setIsEditing] = React.useState(false);
	const [refreshTick, setRefreshTick] = React.useState(0);

	// Ensure badges re-render when variables updated from modal or elsewhere
	useEvent('tpl-vars:updated', (e: CustomEvent<{ selectionID?: string }>) => {
		if (!el.selectionID) return;
		if (e?.detail?.selectionID === el.selectionID) {
			setRefreshTick(t => t + 1);
		}
	});

	// Current effective value for display and starting edit
	const currentValue = React.useMemo(() => {
		if (!tsenode || !varDef) return undefined;
		const v = effectiveVarValueLocal(varDef, tsenode.variables ?? {}, tsenode.toolStates);
		return v;
	}, [tsenode?.variables, tsenode?.toolStates, varDef?.name, refreshTick]);

	// Commit helper
	function commitValue(next: unknown) {
		if (!tsenode || !tsPath) return;

		const nextVars = { ...(tsenode.variables ?? {}) };

		// Normalize empty to undefined (remove)
		const shouldUnset = next === '' || next === undefined || next === null;
		if (shouldUnset) {
			// eslint-disable-next-line @typescript-eslint/no-dynamic-delete
			delete nextVars[el.name];
		} else {
			nextVars[el.name] = next;
		}

		plEditor.tf.setNodes(
			{
				variables: nextVars,
			},
			{ at: tsPath }
		);

		// Signal badges/toolbar for this selection only
		if (tsenode.selectionID) {
			window.dispatchEvent(new CustomEvent('tpl-vars:updated', { detail: { selectionID: tsenode.selectionID } }));
		}

		setIsEditing(false);
		// focus back into editor for good UX
		plEditor.tf.focus();
	}

	// Cancel helper
	function cancelEdit() {
		setIsEditing(false);
		plEditor.tf.focus();
	}

	// Inline input renderer based on var type
	function InlineEditor() {
		const type = varDef?.type ?? VarType.String;

		const commonProps = {
			onKeyDown: (e: React.KeyboardEvent) => {
				e.stopPropagation();
				if (e.key === 'Escape') {
					e.preventDefault();
					cancelEdit();
				}
				if (e.key === 'Enter') {
					// For inputs where Enter is meaningful, we commit. Select/checkbox commit on change anyway.
					e.preventDefault();
					const target = e.target as HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement;
					if (type === VarType.Number) {
						const num = target.value.trim() === '' ? undefined : Number(target.value);
						commitValue(Number.isFinite(num as number) ? num : undefined);
					} else if (type === VarType.Boolean) {
						// no-op; toggles commit on change
					} else if (type === VarType.Enum) {
						commitValue(target.value || undefined);
					} else if (type === VarType.Date) {
						commitValue(target.value || undefined);
					} else {
						commitValue(target.value ?? '');
					}
				}
			},
			onMouseDown: (e: React.MouseEvent) => {
				// keep focus inside input
				e.stopPropagation();
			},
		};

		if (type === VarType.Boolean) {
			return (
				<label className="flex items-center gap-1">
					<input
						type="checkbox"
						className="toggle toggle-accent toggle-xs rounded-full"
						checked={Boolean(currentValue)}
						onChange={e => {
							commitValue(e.target.checked);
						}}
						onBlur={() => {
							commitValue(Boolean(currentValue));
						}}
						{...commonProps}
					/>
				</label>
			);
		}

		if (type === VarType.Number) {
			return (
				<input
					autoFocus
					type="number"
					className="input input-ghost input-xs w-28 min-w-20 rounded-md bg-transparent"
					defaultValue={currentValue === undefined || currentValue === null ? '' : String(currentValue as number)}
					placeholder={varDef?.default !== undefined ? varDef.default : ''}
					onBlur={e => {
						const val = e.currentTarget.value.trim();
						commitValue(val === '' ? undefined : Number(val));
					}}
					{...commonProps}
				/>
			);
		}

		if (type === VarType.Enum) {
			return (
				<select
					autoFocus
					className="select select-ghost select-xs w-32 min-w-24 bg-transparent"
					defaultValue={currentValue === undefined || currentValue === null ? '' : (currentValue as string)}
					onChange={e => {
						commitValue(e.target.value || undefined);
					}}
					onBlur={() => {
						cancelEdit();
					}}
					{...commonProps}
				>
					<option value="">-- select --</option>
					{(varDef?.enumValues ?? []).map(opt => (
						<option value={opt} key={opt}>
							{opt}
						</option>
					))}
				</select>
			);
		}

		if (type === VarType.Date) {
			return (
				<input
					autoFocus
					type="date"
					className="input input-ghost input-xs w-36 min-w-28 bg-transparent"
					defaultValue={currentValue ? (currentValue as string) : ''}
					onBlur={e => {
						commitValue(e.currentTarget.value || undefined);
					}}
					{...commonProps}
				/>
			);
		}

		// String
		return (
			<input
				autoFocus
				className="input input-ghost input-xs w-40 min-w-24 rounded-md bg-transparent"
				// eslint-disable-next-line @typescript-eslint/no-base-to-string
				defaultValue={currentValue === undefined || currentValue === null ? '' : String(currentValue)}
				maxLength={32}
				placeholder={varDef?.default !== undefined ? varDef.default : ''}
				onBlur={e => {
					commitValue(e.currentTarget.value);
				}}
				{...commonProps}
			/>
		);
	}

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
			title={isEditing ? `Editing: ${el.name}` : isMissing ? `Required: ${el.name}` : `Variable: ${el.name}`}
			onKeyDown={e => {
				// Enter toggles inline edit; Esc handled inside edit
				if (e.key === 'Enter' && !isEditing) {
					e.preventDefault();
					e.stopPropagation();
					setIsEditing(true);
				}
			}}
			onMouseDown={e => {
				// allow focusing pill without bubbling into editor selection changes
				e.preventDefault();
			}}
			onClick={() => {
				setIsEditing(true);
			}}
		>
			{isEditing ? (
				<div className="flex items-center gap-1">
					<InlineEditor />
				</div>
			) : (
				<span className="flex items-center gap-1">
					<span className="font-medium">{el.name}</span>
					<FiEdit2 className="opacity-70" size={12} />
				</span>
			)}
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
	const selectionID: string | undefined = tsenode.selectionID;

	const result: any[] = [];
	const re = /\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g;
	let idx = 0;
	let m: RegExpExecArray | null;
	while ((m = re.exec(text)) !== null) {
		const pre = text.slice(idx, m.index);

		if (pre) result.push({ text: pre, ownerSelectionID: selectionID });

		const varName = m[1];
		if (known.has(varName)) {
			const node: TemplateVariableElementNode = {
				type: KEY_TEMPLATE_VARIABLE,
				bundleID: tsenode.bundleID,
				templateSlug: tsenode.templateSlug,
				templateVersion: tsenode.templateVersion,
				selectionID: selectionID,
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
	if (tail) result.push({ text: tail, ownerSelectionID: selectionID });

	if (result.length === 0) result.push({ text: '' });
	return result;
}

// Flatten current editor content into plain text (single-block), replacing variable pills of the first template.
// Used when extracting text to submit without mutating content.
export function toPlainTextReplacingVariables(editor: PlateEditor): string {
	// Find the first selection node for context (variables and values)
	const tpl = getFirstTemplateNodeWithPath(editor);
	const [tsenode] = tpl ?? [];
	const vars = tsenode ? tsenode.variables : {};

	// Walk the top-level single paragraph
	const childnodes = (editor.children?.[0]?.children ?? []) as any[];
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
			const s = toStringDeep(n);
			if (s) parts.push(s);
		}
	});

	return parts.join('');
}

function toStringDeep(n: any): string {
	if (!n || typeof n !== 'object' || n === null) return '';

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
	if (!n || typeof n !== 'object') return false;
	const obj = n as Record<PropertyKey, unknown>;
	return 'type' in obj && obj.type === KEY_TEMPLATE_VARIABLE && 'name' in obj && typeof obj.name === 'string';
}
