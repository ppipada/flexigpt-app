import type { PlateElementProps } from 'platejs/react';

import { computeEffectiveTemplate, computeRequirements } from '@/chats/templates/template_processing';
import { type TemplateSelectionElementNode } from '@/chats/templates/template_spec';

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
