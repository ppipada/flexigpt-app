import type { PlateEditor } from 'platejs/react';
import { useEditorRef } from 'platejs/react';

import type { AttachmentMode } from '@/spec/attachment';
import type { ToolOutput } from '@/spec/tool';

import { DirectoryChip } from '@/chats/attachments/attachment_directory_chip';
import {
	type DirectoryAttachmentGroup,
	type EditorAttachment,
	editorAttachmentKey,
} from '@/chats/attachments/attachment_editor_utils';
import { StandaloneAttachmentsChip } from '@/chats/attachments/attachment_standalone_chips';
import { ToolChipsComposerRow } from '@/chats/tools/tool_chips_composer';
import { ToolChoicesChip } from '@/chats/tools/tool_choices_chip';
import type { EditorToolCall } from '@/chats/tools/tool_editor_utils';
import { getToolNodesWithPath } from '@/chats/tools/tool_editor_utils';

interface EditorChipsBarProps {
	attachments: EditorAttachment[];
	directoryGroups: DirectoryAttachmentGroup[];

	// Tool calls & outputs (tool runners / results)
	toolCalls?: EditorToolCall[];
	toolOutputs?: ToolOutput[];
	isBusy?: boolean;
	onRunToolCall?: (id: string) => void | Promise<void>;
	onDiscardToolCall?: (id: string) => void;
	onOpenOutput?: (output: ToolOutput) => void;
	onRemoveOutput?: (id: string) => void;

	onRemoveAttachment: (att: EditorAttachment) => void;
	onChangeAttachmentMode: (att: EditorAttachment, mode: AttachmentMode) => void;
	onRemoveDirectoryGroup: (groupId: string) => void;
	onRemoveOverflowDir?: (groupId: string, dirPath: string) => void;
}

/**
 * Unified chips bar for:
 *   - Standalone attachments
 *   - Directory groups
 *   - Tool choices (attached tools)
 *   - Tool call & tool output chips
 *
 * Order (left → right):
 *   attachments → directories → tools → tool runners / outputs
 */
export function EditorChipsBar({
	attachments,
	directoryGroups,
	toolCalls = [],
	toolOutputs = [],
	isBusy = false,
	onRunToolCall,
	onDiscardToolCall,
	onOpenOutput,
	onRemoveOutput,
	onRemoveAttachment,
	onChangeAttachmentMode,
	onRemoveDirectoryGroup,
	onRemoveOverflowDir,
}: EditorChipsBarProps) {
	const editor = useEditorRef() as PlateEditor;
	const toolEntries = getToolNodesWithPath(editor);

	const hasVisibleToolCalls = toolCalls.some(
		toolCall => toolCall.status !== 'discarded' && toolCall.status !== 'succeeded'
	);

	const hasAnyChips =
		attachments.length > 0 ||
		directoryGroups.length > 0 ||
		toolEntries.length > 0 ||
		hasVisibleToolCalls ||
		toolOutputs.length > 0;

	if (!hasAnyChips) return null;

	// Attachments that are "owned" by a directory group should not show as top-level attachments.
	const ownedKeys = new Set<string>();
	for (const group of directoryGroups) {
		for (const k of group.ownedAttachmentKeys) {
			ownedKeys.add(k);
		}
	}

	const standaloneAttachments = attachments.filter(att => !ownedKeys.has(editorAttachmentKey(att)));

	const runToolCall = onRunToolCall ?? (() => {});
	const discardToolCall = onDiscardToolCall ?? (() => {});
	const openOutput = onOpenOutput ?? (() => {});
	const removeOutput = onRemoveOutput ?? (() => {});

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

			{/* Tool choices (selected tools for this conversation), aggregated into a dropdown chip */}
			<ToolChoicesChip editor={editor} toolEntries={toolEntries} />

			{/* Tool-call chips (pending/running/failed) and tool output chips */}
			<ToolChipsComposerRow
				toolCalls={toolCalls}
				toolOutputs={toolOutputs}
				isBusy={isBusy}
				onRunToolCall={runToolCall}
				onDiscardToolCall={discardToolCall}
				onOpenOutput={openOutput}
				onRemoveOutput={removeOutput}
			/>
		</div>
	);
}
