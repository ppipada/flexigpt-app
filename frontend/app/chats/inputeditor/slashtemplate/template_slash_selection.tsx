import { ElementApi, NodeApi } from 'platejs';
import type { PlateEditor, PlateElementProps } from 'platejs/react';

import {
	computeEffectiveTemplate,
	computeRequirements,
	makeSelectedTemplateForRun,
	type SelectedTemplateForRun,
	type TemplateSelectionElementNode,
} from '@/chats/inputeditor/slashtemplate/template_processing';

export const KEY_TEMPLATE_SELECTION = 'templateSelection';

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

/**
 * Template selection element (data carrier).
 * We render it as a hidden inline element so it doesn't affect the text layout.
 * The toolbar is responsible for user-facing controls.
 */
export function TemplateSelectionElement(props: PlateElementProps<any>) {
	const { element, attributes, children } = props as any;
	const el = element as TemplateSelectionElementNode;

	// We still compute badges for accessibility/title, but hide it from visual flow.
	const { template, variablesSchema, preProcessors } = computeEffectiveTemplate(el);
	const req = computeRequirements(variablesSchema, el.variables, preProcessors, el.toolStates);

	return (
		<span
			{...attributes}
			contentEditable={false}
			className="pointer-events-none sr-only"
			data-template-chip
			title={`Template: ${el.overrides?.displayName ?? template?.displayName ?? el.templateSlug} • pending vars: ${req.requiredCount} • tools: ${
				preProcessors.length
			}`}
			aria-hidden="true"
		>
			{/* Invisible info holder */}
			<span className="sr-only">{el.overrides?.displayName ?? template?.displayName ?? el.templateSlug}</span>
			{children}
		</span>
	);
}
