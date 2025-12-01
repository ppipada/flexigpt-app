import { FiFileText, FiImage, FiLink, FiTool } from 'react-icons/fi';

import type { Attachment } from '@/spec/attachment';
import { AttachmentKind, AttachmentMode } from '@/spec/attachment';
import type { ConversationToolChoice } from '@/spec/conversation';

import {
	getAttachmentModeLabel,
	getAttachmentModePillClasses,
	getAttachmentModeTooltip,
} from '@/chats/attachments/attachment_mode_menu';

interface MessageAttachmentsBarProps {
	attachments?: Attachment[];
	toolChoices?: ConversationToolChoice[];
}

/**
 * Get a path/URL for tooltip display, similar to getEditorAttachmentPath
 * but for persisted Conversation attachments.
 */
function getAttachmentPath(att: Attachment): string {
	if (att.kind === AttachmentKind.file && att.fileRef) {
		return att.fileRef.path;
	}
	if (att.kind === AttachmentKind.image && att.imageRef) {
		return att.imageRef.path;
	}
	if (att.kind === AttachmentKind.url && att.urlRef) {
		return att.urlRef.url;
	}
	return '';
}

interface AttachmentInfoChipProps {
	attachment: Attachment;
}

/**
 * Read‑only chip for a single attachment (file/image/url).
 * No remove, no mode menu — just info.
 */
function AttachmentInfoChip({ attachment }: AttachmentInfoChipProps) {
	const { kind, label } = attachment;

	const icon =
		kind === AttachmentKind.image ? (
			<FiImage size={14} />
		) : kind === AttachmentKind.url ? (
			<FiLink size={14} />
		) : (
			<FiFileText size={14} />
		);

	const isLabelTruncated = label.length > 40;
	const truncated = isLabelTruncated ? label.slice(0, 37) + '…' : label;

	const path = getAttachmentPath(attachment);

	const tooltipLines: string[] = [];
	if (isLabelTruncated) tooltipLines.push(label);
	if (path && path !== label) tooltipLines.push(path);

	const title = tooltipLines.length > 0 ? tooltipLines.join('\n') : undefined;

	const mode = attachment.mode ?? AttachmentMode.notReadable;
	const modeLabel = getAttachmentModeLabel(mode);
	const modeTooltip = getAttachmentModeTooltip(mode);

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0"
			title={title}
			data-message-chip="attachment"
		>
			<span className="shrink-0">{icon}</span>
			<span className="max-w-64 truncate">{truncated}</span>
			<span
				className={getAttachmentModePillClasses(mode, false /* non‑interactive */)}
				title={modeTooltip}
				data-attachment-mode-pill
			>
				{modeLabel}
			</span>
		</div>
	);
}

interface ToolInfoChipProps {
	tool: ConversationToolChoice;
}

/**
 * Read‑only chip for a tool choice used for this message.
 */
function ToolInfoChip({ tool }: ToolInfoChipProps) {
	const name = tool.displayName || tool.toolSlug;
	const slug = `${tool.bundleID}/${tool.toolSlug}@${tool.toolVersion}`;
	const tooltipLines: string[] = [name, slug];
	if (tool.description) tooltipLines.push(tool.description);

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0"
			title={tooltipLines.join('\n')}
			data-message-chip="tool"
		>
			<FiTool size={14} />
			<span className="max-w-64 truncate">{name}</span>
		</div>
	);
}

/**
 * Read‑only, flat “toolbar” at bottom of a message bubble.
 * Shows attachments + tool choices for that message.
 * No directory groups, no remove buttons, no menus.
 */
export function MessageAttachmentsBar({ attachments, toolChoices }: MessageAttachmentsBarProps) {
	const hasAttachments = !!attachments && attachments.length > 0;
	const hasTools = !!toolChoices && toolChoices.length > 0;

	if (!hasAttachments && !hasTools) return null;

	return (
		<div className="border-base-200 flex min-h-8 max-w-full min-w-0 flex-1 items-center gap-1 overflow-x-auto border-t px-2 py-1 text-xs">
			{hasAttachments &&
				attachments.map((att, index) => (
					<AttachmentInfoChip
						// Attachments are not guaranteed unique; key by index + core identity.
						key={`${att.kind}:${att.label}:${index}`}
						attachment={att}
					/>
				))}
			{hasTools &&
				toolChoices.map(tool => (
					<ToolInfoChip key={tool.id ?? `${tool.bundleID}-${tool.toolSlug}-${tool.toolVersion}`} tool={tool} />
				))}
		</div>
	);
}
