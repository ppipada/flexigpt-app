import { ElementApi, NodeApi } from 'platejs';
import type { PlateEditor, PlateElementProps } from 'platejs/react';
import { FiEdit2, FiX } from 'react-icons/fi';

import { usePromptTemplate } from '@/hooks/use_template';

export const KEY_TEMPLATE_SELECTION = 'templateSelection';

interface TemplateSelectionElementBaseProps {
	type: typeof KEY_TEMPLATE_SELECTION;
	bundleID: string;
	templateSlug: string;
	templateVersion: string;
	variables: Record<string, unknown>;
}

export function getTemplateSelections(editor: PlateEditor) {
	const elList = NodeApi.elements(editor);
	const selections: TemplateSelectionElementBaseProps[] = [];
	elList.forEach(([el]) => {
		if (ElementApi.isElementType(el, KEY_TEMPLATE_SELECTION)) {
			selections.push({
				type: KEY_TEMPLATE_SELECTION,
				bundleID: el.bundleID as string,
				templateSlug: el.templateSlug as string,
				templateVersion: el.templateVersion as string,
				variables: el.variables as Record<string, unknown>,
			});
		}
	});

	// console.log(JSON.stringify(selections, null, 2));
	return selections;
}

/* eslint-disable @typescript-eslint/no-unsafe-member-access */
export function TemplateSelectionElement(props: PlateElementProps<any>) {
	const { element, attributes, children, path, editor } = props as any;
	const { bundleID, templateSlug, templateVersion } = element;
	const tmpl = usePromptTemplate(bundleID, templateSlug, templateVersion);

	return (
		<span {...attributes} contentEditable={false} className="badge badge-neutral inline-flex items-center gap-1">
			{tmpl?.displayName ?? templateSlug}
			<button
				type="button"
				className="btn btn-ghost btn-xs p-0"
				onMouseDown={e => {
					e.preventDefault();
				}}
				onClick={() => {
					alert('variable editor TBD');
				}}
			>
				<FiEdit2 size={12} />
			</button>
			<button
				type="button"
				className="btn btn-ghost btn-xs p-0"
				onMouseDown={e => {
					e.preventDefault();
				}}
				onClick={() => {
					editor.tf.removeNodes({ at: path });
				}}
			>
				<FiX size={12} />
			</button>
			{children}
		</span>
	);
}
