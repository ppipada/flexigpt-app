import { FiFileText, FiTool, FiX } from 'react-icons/fi';

import type { PlateEditor } from 'platejs/react';
import { useEditorRef } from 'platejs/react';

import {
	type EditorAttachment,
	editorAttachmentKey,
	getEditorAttachmentPath,
} from '@/chats/attachments/editor_attachment_utils';
import { getToolNodesWithPath, removeToolByKey, toolIdentityKey } from '@/chats/attachments/tool_editor_utils';

/**
 * Scrollable chips bar for attachments and tools, shown inside the Plate editor
 */
interface AttachmentChipsBarProps {
	attachments: EditorAttachment[];
	onRemoveAttachment: (att: EditorAttachment) => void;
}

export function AttachmentChipsBar({ attachments, onRemoveAttachment }: AttachmentChipsBarProps) {
	const editor = useEditorRef() as PlateEditor;
	const toolEntries = getToolNodesWithPath(editor);

	const hasAnyChips = attachments.length > 0 || toolEntries.length > 0;
	if (!hasAnyChips) return null;

	return (
		<div className="flex min-h-8 max-w-full min-w-0 flex-1 items-center gap-1 overflow-x-auto px-1 py-0 text-xs">
			{attachments.map(att => {
				const key = editorAttachmentKey(att);
				const label = att.label.length > 40 ? att.label.slice(0, 37) + '...' : att.label;
				const path = getEditorAttachmentPath(att);
				return (
					<div
						key={key}
						className="bg-base-200 hover:bg-base-300/80 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0"
						title={`${att.kind} attachment: ${att.label}${path ? ` (${path})` : ''}`}
						data-attachment-chip="attachment"
					>
						<FiFileText />
						<span className="truncate">{label}</span>
						<button
							type="button"
							className="btn btn-ghost btn-xs text-error shrink-0 px-1 py-0 shadow-none"
							onClick={() => {
								onRemoveAttachment(att);
							}}
							title="Remove attachment"
							aria-label="Remove attachment"
						>
							<FiX />
						</button>
					</div>
				);
			})}

			{toolEntries.map(([node]) => {
				const n = node;
				const display = n.toolSnapshot?.displayName ?? n.toolSlug;
				const slug = `${n.bundleSlug ?? n.bundleID}/${n.toolSlug}@${n.toolVersion}`;
				const identityKey = toolIdentityKey(n.bundleID, n.bundleSlug, n.toolSlug, n.toolVersion);

				return (
					<div
						key={n.selectionID}
						className="bg-base-200 hover:bg-base-300/80 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0"
						title={`Tool choice: ${display} (${slug})`}
						data-attachment-chip="tool"
						data-selection-id={n.selectionID}
					>
						<FiTool />
						<span className="truncate">{display.length > 32 ? display.slice(0, 32) + '...' : display}</span>
						<button
							type="button"
							className="btn btn-ghost btn-xs text-error shrink-0 px-1 py-0 shadow-none"
							onClick={() => {
								removeToolByKey(editor, identityKey);
							}}
							title="Remove tool choice"
							aria-label="Remove tool choice"
						>
							<FiX />
						</button>
					</div>
				);
			})}
		</div>
	);
}
