/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { FiFileText, FiImage, FiLink, FiTool } from 'react-icons/fi';

import type { Attachment } from '@/spec/attachment';
import { AttachmentKind, AttachmentMode } from '@/spec/attachment';
import type { ToolCall, ToolChoice, ToolOutput } from '@/spec/tool';

import {
	getAttachmentModeLabel,
	getAttachmentModePillClasses,
	getAttachmentModeTooltip,
} from '@/chats/attachments/attachment_mode_menu';
import type { ToolCallChip } from '@/chats/tools/tool_chips';
import { formatToolCallChipLabel, getPrettyToolName } from '@/chats/tools/tool_chips';

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
			<span className={getAttachmentModePillClasses(mode, false)} title={modeTooltip} data-attachment-mode-pill>
				{modeLabel}
			</span>
		</div>
	);
}

interface ToolChoiceChipProps {
	tool: ToolChoice;
}

/**
 * Read‑only chip for a tool choice used for this message.
 */
function ToolChoiceChip({ tool }: ToolChoiceChipProps) {
	const name = tool.displayName || tool.toolSlug;
	const slug = `${tool.bundleID}/${tool.toolSlug}@${tool.toolVersion}`;
	const tooltipLines: string[] = [name, slug];
	if (tool.description) tooltipLines.push(tool.description);

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0"
			title={tooltipLines.join('\n')}
			data-message-chip="tool-choice"
		>
			<FiTool size={14} />
			<span className="max-w-64 truncate">{name}</span>
		</div>
	);
}

interface ToolCallChipProps {
	call: ToolCall;
}

/**
 * Read‑only chip for an assistant-suggested tool call under the assistant bubble.
 */
function ToolCallChip({ call }: ToolCallChipProps) {
	const tmpChip: ToolCallChip = {
		id: call.id || call.callID,
		callID: call.callID,
		name: call.name,
		arguments: call.arguments,
		type: call.type,
		status: 'pending',
	};
	const label = formatToolCallChipLabel(tmpChip);

	const statusLabel = call.status ? ` (${call.status})` : '';
	const title = `Suggested tool call: ${label}${statusLabel}`;

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0"
			title={title}
			data-message-chip="tool-suggested"
		>
			<FiTool size={14} />
			<span className="text-base-content/60 text-[10px] uppercase">Suggested</span>
			<span className="max-w-64 truncate">{label}</span>
		</div>
	);
}

interface ToolOutputChipProps {
	output: ToolOutput;
}

/**
 * Read‑only chip for a tool output that was attached to this user turn.
 * History chips are not interactive; the full output was already used
 * when the turn was sent.
 */
function ToolOutputChip({ output }: ToolOutputChipProps) {
	const prettyName = getPrettyToolName(output.name);
	const label = output.summary || `Result: ${prettyName}`;
	const titleLines = [label, `Tool: ${output.name}`, `Call ID: ${output.callID}`];
	const title = titleLines.join('\n');

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0"
			title={title}
			data-message-chip="tool-output"
		>
			<FiTool size={14} />
			<span className="text-base-content/60 text-[10px] uppercase">Output</span>
			<span className="max-w-64 truncate">{label}</span>
		</div>
	);
}

interface MessageAttachmentsBarProps {
	attachments?: Attachment[];
	toolChoices?: ToolChoice[];
	toolCalls?: ToolCall[];
	toolOutputs?: ToolOutput[];
}

/**
 * Read‑only toolbar under a message bubble:
 * - For user messages: files, tool choices, and tool outputs.
 * - For assistant messages: files and suggested tool calls.
 */
export function MessageAttachmentsBar({
	attachments,
	toolChoices,
	toolCalls,
	toolOutputs,
}: MessageAttachmentsBarProps) {
	const hasAttachments = !!attachments && attachments.length > 0;
	const hasTools = !!toolChoices && toolChoices.length > 0;
	const hasToolCalls = !!toolCalls && toolCalls.length > 0;
	const hasToolOutputs = !!toolOutputs && toolOutputs.length > 0;

	if (!hasAttachments && !hasTools && !hasToolCalls && !hasToolOutputs) return null;

	return (
		<div className="border-base-200 flex min-h-8 max-w-full min-w-0 flex-1 items-center gap-1 overflow-x-auto border-t px-2 py-1 text-xs">
			{/* Attachments visible for both user & assistant */}
			{hasAttachments &&
				attachments?.map((att, index) => (
					<AttachmentInfoChip key={`${att.kind}:${att.label}:${index}`} attachment={att} />
				))}

			{hasTools &&
				toolChoices?.map(tool => (
					<ToolChoiceChip key={tool.toolID ?? `${tool.bundleID}-${tool.toolSlug}-${tool.toolVersion}`} tool={tool} />
				))}

			{/* Tool outputs only on user messages */}
			{hasToolOutputs && toolOutputs?.map(out => <ToolOutputChip key={out.id} output={out} />)}

			{/* Suggested tool calls only on assistant messages */}
			{hasToolCalls && toolCalls?.map(call => <ToolCallChip key={call.id || call.callID} call={call} />)}
		</div>
	);
}
