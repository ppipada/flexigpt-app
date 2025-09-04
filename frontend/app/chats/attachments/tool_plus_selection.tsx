import type { PlateElementProps } from 'platejs/react';

import { type ToolSelectionElementNode } from '@/chats/attachments/tool_editor_utils';

/**
 * Hidden inline element; acts as a data carrier for one selected tool.
 * Chips are rendered in the bottom attachments bar, not inline in content.
 */
export function ToolSelectionElement(props: PlateElementProps<any>) {
	const { element, attributes, children } = props as any;
	const el = element as ToolSelectionElementNode;

	const display = el.overrides?.displayName ?? el.toolSnapshot?.displayName ?? el.toolSlug;
	const slug = `${el.bundleSlug ?? el.bundleID}/${el.toolSlug}@${el.toolVersion}`;

	return (
		<span
			{...attributes}
			contentEditable={false}
			className="pointer-events-none sr-only"
			data-tool-chip
			title={`Tool: ${display} • ${slug}`}
			aria-hidden="true"
		>
			<span className="sr-only">{display}</span>
			{children}
		</span>
	);
}
