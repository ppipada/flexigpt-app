import { FiFileText, FiPaperclip, FiSend, FiTool, FiX } from 'react-icons/fi';

import { type PlateEditor, useEditorRef } from 'platejs/react';

import type { ConversationAttachment } from '@/spec/conversation';

import { getToolNodesWithPath, removeToolByKey, toolIdentityKey } from '@/chats/attachments/tool_editor_utils';

interface AttachmentBottomBarProps {
	attachments: ConversationAttachment[];
	onAttachFiles: () => void;
	onRemoveAttachment: (att: ConversationAttachment) => void;
	onOpenToolPicker: () => void;
	isBusy: boolean;
	isSendButtonEnabled: boolean;
}

/**
  Bottom bar for rendering attached items (tools/files/docs) and housing
  attach/tool/send controls. Always visible; chips area is horizontally scrollable.
*/
export function AttachmentBottomBar({
	attachments,
	onAttachFiles,
	onRemoveAttachment,
	onOpenToolPicker,
	isBusy,
	isSendButtonEnabled,
}: AttachmentBottomBarProps) {
	const editor = useEditorRef() as PlateEditor;

	// Recompute on each render; the parent EditorArea re-renders on doc changes.
	const toolEntries = getToolNodesWithPath(editor);

	const hasAttachments = attachments.length > 0 || toolEntries.length > 0;

	return (
		<div
			className="border-base-300 bg-base-100/95 supports-backdrop-filter:bg-base-100/60 w-full border-t backdrop-blur"
			data-attachments-bottom-bar
			aria-label="Attachments and tools"
		>
			<div className="flex items-center gap-2 px-2 py-1 text-xs">
				<div className="flex items-center gap-1">
					<button
						type="button"
						className="btn btn-ghost btn-xs px-2"
						onClick={onAttachFiles}
						title="Attach files"
						aria-label="Attach files"
					>
						<FiPaperclip size={14} />
					</button>
					<button
						type="button"
						className="btn btn-ghost btn-xs px-2"
						onClick={onOpenToolPicker}
						title="Add tool choice"
						aria-label="Add tool choice"
					>
						<FiTool size={14} />
					</button>
				</div>

				{/* Chips scroller */}
				<div className="no-scrollbar flex min-w-0 flex-1 items-center gap-2 overflow-x-auto py-0">
					{attachments.map(att => {
						const key = `${att.kind}:${att.ref}`;
						const label = att.label.length > 40 ? att.label.slice(0, 37) + '...' : att.label;
						return (
							<div
								key={key}
								className="bg-base-200 hover:bg-base-300/80 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0 text-xs"
								title={`${att.kind} attachment: ${att.label}`}
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
								className="bg-base-200 hover:bg-base-300/80 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0 text-xs"
								title={`Tool choice: ${display} (${slug})`}
								data-attachment-chip="tool"
								data-selection-id={n.selectionID}
							>
								<FiTool />
								<span className="truncate">{display.length > 36 ? display.slice(0, 36) + '...' : display}</span>
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

					{!hasAttachments ? (
						<span className="text-base-content/60 truncate text-[11px] whitespace-nowrap">
							Add tools or attach files for this turn
						</span>
					) : null}
				</div>

				<button
					type="submit"
					className={`btn btn-circle btn-ghost shrink-0 ${!isSendButtonEnabled || isBusy ? 'btn-disabled' : ''}`}
					disabled={isBusy || !isSendButtonEnabled}
					aria-label="Send message"
					title="Send message"
				>
					<FiSend size={16} />
				</button>
			</div>
		</div>
	);
}
