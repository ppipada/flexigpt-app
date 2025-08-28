import React from 'react';

import { ElementApi, NodeApi } from 'platejs';
import type { PlateEditor, PlateElementProps } from 'platejs/react';
import { FiEdit2, FiTool, FiX } from 'react-icons/fi';

import { TemplateEditModal } from '@/chats/inputeditor/slashtemplate/template_edit_modal';
import {
	computeEffectiveTemplate,
	computeRequirements,
	makeSelectedTemplateForRun,
	type SelectedTemplateForRun,
	type TemplateSelectionElementNode,
} from '@/chats/inputeditor/slashtemplate/template_processing';

export const KEY_TEMPLATE_SELECTION = 'templateSelection';

export function getTemplateSelections(editor: PlateEditor): SelectedTemplateForRun[] {
	const elList = NodeApi.elements(editor);
	const selections: SelectedTemplateForRun[] = [];

	elList.forEach(([el]) => {
		if (ElementApi.isElementType(el, KEY_TEMPLATE_SELECTION)) {
			const node = el as unknown as TemplateSelectionElementNode;
			selections.push(makeSelectedTemplateForRun(node));
		}
	});

	return selections;
}

export function TemplateSelectionElement(props: PlateElementProps<any>) {
	const { element, attributes, children, path, editor } = props as any;
	const plEditor = editor as PlateEditor;
	const el = element as TemplateSelectionElementNode;

	const { template, variablesSchema, preProcessors } = computeEffectiveTemplate(el);
	const req = computeRequirements(variablesSchema, el.variables, preProcessors, el.toolStates);

	const [open, setOpen] = React.useState(false);

	return (
		<span
			{...attributes}
			contentEditable={false}
			className="badge badge-neutral inline-flex items-center gap-1"
			data-template-chip
		>
			<span className="flex items-center gap-2">
				<span className="font-medium">{el.overrides?.displayName ?? template?.displayName ?? el.templateSlug}</span>

				{req.requiredCount > 0 ? (
					<span className="badge badge-warning badge-sm" title="Required variables">
						req {req.requiredCount}
					</span>
				) : (
					<span className="badge badge-success badge-sm" title="All variables provided">
						ready
					</span>
				)}

				{req.toolsToRun.some(t => t.status !== 'done') && (
					<span className="badge badge-info badge-sm inline-flex items-center gap-1" title="Pending tools">
						<FiTool className="inline" /> {req.toolsToRun.filter(t => t.status !== 'done').length}
					</span>
				)}
			</span>

			<button
				type="button"
				className="btn btn-ghost btn-xs p-0"
				onMouseDown={e => {
					e.preventDefault();
				}}
				onClick={() => {
					setOpen(true);
				}}
				title="Edit template values"
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
					plEditor.tf.removeNodes({ at: path });
				}}
				title="Remove"
			>
				<FiX size={12} />
			</button>

			<TemplateEditModal
				open={open}
				onClose={() => {
					setOpen(false);
				}}
				tsenode={el}
				editor={plEditor}
				path={path}
			/>

			{children}
		</span>
	);
}
