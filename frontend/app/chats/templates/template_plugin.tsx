import { createSlatePlugin } from 'platejs';
import type { PlateElementProps } from 'platejs/react';

import { computeEffectiveTemplate, computeRequirements } from '@/chats/templates/template_processing';
import {
	KEY_TEMPLATE_SELECTION,
	KEY_TEMPLATE_VARIABLE,
	type TemplateSelectionElementNode,
} from '@/chats/templates/template_spec';
import { TemplateVariableElement } from '@/chats/templates/template_variables_inline';

/**
 * Template selection element (data carrier).
 * We render it as a hidden inline element so it doesn't affect the text layout.
 * The toolbar is responsible for user-facing controls.
 */
function TemplateSelectionElement(props: PlateElementProps<any>) {
	const { element, attributes, children } = props as any;
	const el = element as TemplateSelectionElementNode;

	// We still compute badges for accessibility/title, but hide it from visual flow.
	const { template, variablesSchema } = computeEffectiveTemplate(el);
	const req = computeRequirements(variablesSchema, el.variables);

	return (
		<span
			{...attributes}
			contentEditable={false}
			className="pointer-events-none sr-only"
			data-template-chip
			title={`Template: ${el.overrides?.displayName ?? template?.displayName ?? el.templateSlug} • pending vars: ${req.requiredCount}`}
			aria-hidden="true"
		>
			{/* Invisible info holder */}
			<span className="sr-only">{el.overrides?.displayName ?? template?.displayName ?? el.templateSlug}</span>
			{children}
		</span>
	);
}

const TemplateSelectionPlugin = createSlatePlugin({
	key: KEY_TEMPLATE_SELECTION,
	// this is the chip “schema”
	node: { isElement: true, isInline: true, isVoid: true },
});

// Plugin
const TemplateVariablePlugin = createSlatePlugin({
	key: KEY_TEMPLATE_VARIABLE,
	node: { isElement: true, isInline: true, isVoid: true },
});

export const TemplateSlashKit = [
	// Data-carrier for selections:
	TemplateSelectionPlugin.withComponent(TemplateSelectionElement),
	// Data-carrier for variables:
	TemplateVariablePlugin.withComponent(TemplateVariableElement),
];
