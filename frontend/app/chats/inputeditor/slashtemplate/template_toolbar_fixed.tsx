import { FiFileText, FiTool, FiX } from 'react-icons/fi';

import type { SelectedTemplateForRun } from '@/chats/inputeditor/slashtemplate/template_processing';

export function TemplateFixedToolbar(props: {
	selection: SelectedTemplateForRun;
	flashing?: boolean;
	onOpenModal: () => void;
	onRemove: () => void;
	onFlatten: () => void;
}) {
	const { selection, flashing, onOpenModal, onRemove, onFlatten } = props;

	const templateName =
		selection.template.displayName || selection.templateSlug || selection.template.slug || 'Template';

	const varBadge =
		selection.requiredCount > 0 ? (
			<span
				className="badge badge-warning"
				title="Required variables pending"
			>{`pending ${selection.requiredCount} vars`}</span>
		) : (
			<span className="badge badge-success" title="All variables provided">
				vars ok
			</span>
		);

	const totalTools = selection.toolsToRun.length;
	const pendingTools = selection.toolsToRun.filter(t => t.status === 'pending').length;
	const toolsBadge =
		totalTools > 0 ? (
			<span
				className={`badge ${pendingTools > 0 ? 'badge-info' : 'badge-success'}`}
				title={pendingTools > 0 ? `${pendingTools} tools pending` : 'All tools ready/done'}
			>
				<FiTool className="mr-1" /> {totalTools} tool{totalTools === 1 ? '' : 's'}
			</span>
		) : (
			<span className="badge" title="No tools configured">
				<FiTool className="mr-1" /> 0 tool
			</span>
		);

	return (
		<div
			className={`flex w-full items-center gap-3 border-b px-3 py-2 ${
				flashing ? 'ring-error rounded-t-2xl ring ring-offset-0' : ''
			}`}
			data-template-toolbar
		>
			{/* Clickable area opens modal (except buttons on the right) */}
			<button
				type="button"
				className="btn btn-ghost btn-sm no-animation flex items-center gap-3"
				onClick={onOpenModal}
				title="Open template editor"
			>
				<span className="font-medium">{templateName}</span>
				<span className="opacity-60">·</span>
				{varBadge}
				<span className="opacity-60">·</span>
				{toolsBadge}
			</button>

			<div className="grow" />

			{/* Expand to plain text */}
			<button
				type="button"
				className="btn btn-ghost btn-sm"
				onClick={onFlatten}
				title="Convert chips into plain text"
				aria-label="Convert template chips into plain text"
			>
				<FiFileText />
			</button>

			{/* Remove template */}
			<button
				type="button"
				className="btn btn-ghost btn-sm text-error"
				onClick={onRemove}
				title="Remove template"
				aria-label="Remove template"
			>
				<FiX />
			</button>
		</div>
	);
}
