import { FiChevronDown, FiPaperclip, FiTool, FiX } from 'react-icons/fi';

import { Menu, MenuButton, MenuItem, useMenuStore } from '@ariakit/react';
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

interface StandaloneAttachmentsChipProps {
	attachments: EditorAttachment[];
	onRemoveAttachment: (att: EditorAttachment) => void;
	onChangeAttachmentMode: (att: EditorAttachment, mode: AttachmentMode) => void;
}

/**
 * Aggregated "Attachments" chip for standalone files/links.
 * Opens a dropdown listing each attachment as a full-width chip.
 */
function StandaloneAttachmentsChip({
	attachments,
	onRemoveAttachment,
	onChangeAttachmentMode,
}: StandaloneAttachmentsChipProps) {
	const count = attachments.length;
	const menu = useMenuStore({ placement: 'bottom-start', focusLoop: true });

	if (count === 0) return null;

	const title = `Attachments\n${count} item${count === 1 ? '' : 's'} attached`;

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-1 rounded-2xl px-2 py-0"
			title={title}
			data-attachment-chip="attachments-group"
		>
			<FiPaperclip size={14} />
			<span className="max-w-24 truncate">Attachments</span>
			<span className="text-base-content/70 text-[11px] whitespace-nowrap">{count}</span>

			<MenuButton
				store={menu}
				className="btn btn-ghost btn-xs px-0 py-0 shadow-none"
				aria-label="Show attached files and links"
				title="Show attached files and links"
			>
				<FiChevronDown size={14} />
			</MenuButton>

			<Menu
				store={menu}
				gutter={6}
				className="rounded-box bg-base-100 text-base-content border-base-300 z-50 max-h-72 min-w-[260px] overflow-y-auto border p-2 shadow-xl focus-visible:outline-none"
				autoFocusOnShow
			>
				<div className="text-base-content/70 mb-1 text-[11px] font-semibold">Attachments</div>

				{attachments.map(att => (
					<MenuItem
						key={editorAttachmentKey(att)}
						store={menu}
						hideOnClick={false}
						className="data-active-item:bg-base-200 mb-1 rounded-xl last:mb-0"
					>
						<AttachmentChip
							attachment={att}
							onRemoveAttachment={onRemoveAttachment}
							onChangeAttachmentMode={onChangeAttachmentMode}
							fullWidth
						/>
					</MenuItem>
				))}
			</Menu>
		</div>
	);
}

/**
 * Chips bar for attachments, directory groups, and tool choices.
 * Tool-call and tool-output chips live in a separate row (see ToolChipsComposerRow).
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

	// Attachments that are "owned" by a directory group should not show as top-level attachments.
	const ownedKeys = new Set<string>();
	for (const group of directoryGroups) {
		for (const k of group.ownedAttachmentKeys) {
			ownedKeys.add(k);
		}
	}

	const standaloneAttachments = attachments.filter(att => !ownedKeys.has(editorAttachmentKey(att)));

	return (
		<div className="flex shrink-0 items-center gap-1">
			{/* Aggregated chip for standalone attachments */}
			<StandaloneAttachmentsChip
				attachments={standaloneAttachments}
				onRemoveAttachment={onRemoveAttachment}
				onChangeAttachmentMode={onChangeAttachmentMode}
			/>

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

			{/* Tool choices (selected tools for this conversation) */}
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
						data-attachment-chip="tool-choice"
						data-selection-id={n.selectionID}
					>
						<FiTool size={14} />
						<span className="truncate">{display.length > 32 ? `${display.slice(0, 32)}…` : display}</span>
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
