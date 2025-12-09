// attachment_chips_bar.tsx
import { FiTool, FiX } from 'react-icons/fi';

import type { PlateEditor } from 'platejs/react';
import { useEditorRef } from 'platejs/react';

import type { AttachmentMode } from '@/spec/attachment';

import { AttachmentChip } from '@/chats/attachments/attachment_chip';
import { DirectoryChip } from '@/chats/attachments/attachment_directory_chip';
import {
	type DirectoryAttachmentGroup,
	type EditorAttachment,
	editorAttachmentKey,
} from '@/chats/attachments/attachment_editor_utils';
import { getToolNodesWithPath, removeToolByKey, toolIdentityKey } from '@/chats/tools/tool_editor_utils';

interface AttachmentChipsBarProps {
	attachments: EditorAttachment[];
	directoryGroups: DirectoryAttachmentGroup[];
	onRemoveAttachment: (att: EditorAttachment) => void;
	onChangeAttachmentMode: (att: EditorAttachment, mode: AttachmentMode) => void;
	onRemoveDirectoryGroup: (groupId: string) => void;
	onRemoveOverflowDir?: (groupId: string, dirPath: string) => void;
}

/**
 * Scrollable chips bar for attachments, directory groups, and tools.
 */
export function AttachmentChipsBar({
	attachments,
	directoryGroups,
	onRemoveAttachment,
	onChangeAttachmentMode,
	onRemoveDirectoryGroup,
	onRemoveOverflowDir,
}: AttachmentChipsBarProps) {
	const editor = useEditorRef() as PlateEditor;
	const toolEntries = getToolNodesWithPath(editor);

	const hasAnyChips = attachments.length > 0 || directoryGroups.length > 0 || toolEntries.length > 0;
	if (!hasAnyChips) return null;

	// Attachments that are "owned" by a directory group should not show as top-level chips.
	const ownedKeys = new Set<string>();
	for (const group of directoryGroups) {
		for (const k of group.ownedAttachmentKeys) {
			ownedKeys.add(k);
		}
	}

	const standaloneAttachments = attachments.filter(att => !ownedKeys.has(editorAttachmentKey(att)));

	return (
		<div className="flex min-h-8 max-w-full min-w-0 flex-1 items-center gap-1 overflow-x-auto px-1 py-0 text-xs">
			{/* Standalone attachments (not owned by a folder) */}
			{standaloneAttachments.map(att => (
				<AttachmentChip
					key={editorAttachmentKey(att)}
					attachment={att}
					onRemoveAttachment={onRemoveAttachment}
					onChangeAttachmentMode={onChangeAttachmentMode}
				/>
			))}

			{/* Folder groups */}
			{directoryGroups.map(group => (
				<DirectoryChip
					key={group.id}
					group={group}
					attachments={attachments}
					onRemoveAttachment={onRemoveAttachment}
					onChangeAttachmentMode={onChangeAttachmentMode}
					onRemoveDirectoryGroup={onRemoveDirectoryGroup}
					onRemoveOverflowDir={onRemoveOverflowDir}
				/>
			))}

			{/* Tool chips unchanged */}
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
						<FiTool size={14} />
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
							<FiX size={14} />
						</button>
					</div>
				);
			})}
		</div>
	);
}
