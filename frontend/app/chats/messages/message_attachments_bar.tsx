/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { FiChevronDown, FiFileText, FiImage, FiLink, FiPaperclip, FiTool } from 'react-icons/fi';

import { Menu, MenuButton, MenuItem, useMenuStore } from '@ariakit/react';

import type { Attachment } from '@/spec/attachment';
import { AttachmentKind, AttachmentMode } from '@/spec/attachment';
import type { ToolStoreChoice, UIToolCall, UIToolOutput } from '@/spec/tool';

import {
	getAttachmentModeLabel,
	getAttachmentModePillClasses,
	getAttachmentModeTooltip,
} from '@/chats/attachments/attachment_mode_menu';
import { formatToolCallLabel, getPrettyToolName } from '@/chats/tools/tool_editor_utils';

/**
 * Get a path/URL for tooltip display, similar to getUIAttachmentPath
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

interface MessageAttachmentInfoChipProps {
	attachment: Attachment;
	fullWidth?: boolean;
}

/**
 * Read‑only chip for a single attachment (file/image/url).
 * No remove, no mode menu — just info.
 */
function MessageAttachmentInfoChip({ attachment, fullWidth = false }: MessageAttachmentInfoChipProps) {
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

	const containerClasses = [
		'bg-base-200 text-base-content flex items-center gap-2 rounded-2xl px-2 py-0',
		fullWidth ? 'w-full' : 'shrink-0',
	]
		.filter(Boolean)
		.join(' ');

	const labelClasses = fullWidth ? 'min-w-0 flex-1 truncate' : 'max-w-64 truncate';

	return (
		<div className={containerClasses} title={title} data-message-chip="attachment">
			<span className="shrink-0">{icon}</span>
			<span className={labelClasses}>{truncated}</span>
			<span className={getAttachmentModePillClasses(mode, false)} title={modeTooltip} data-attachment-mode-pill>
				{modeLabel}
			</span>
		</div>
	);
}

interface MessageToolChoiceChipProps {
	tool: ToolStoreChoice;
	fullWidth?: boolean;
}

/**
 * Read‑only chip for a tool choice used for this message.
 */
function MessageToolChoiceChip({ tool, fullWidth = false }: MessageToolChoiceChipProps) {
	const name = tool.displayName || tool.toolSlug;
	const slug = `${tool.bundleID}/${tool.toolSlug}@${tool.toolVersion}`;
	const tooltipLines: string[] = [name, slug];
	if (tool.description) tooltipLines.push(tool.description);

	const containerClasses = [
		'bg-base-200 text-base-content flex items-center gap-2 rounded-2xl px-2 py-0',
		fullWidth ? 'w-full' : 'shrink-0',
	]
		.filter(Boolean)
		.join(' ');

	const labelClasses = fullWidth ? 'min-w-0 flex-1 truncate' : 'max-w-64 truncate';

	return (
		<div className={containerClasses} title={tooltipLines.join('\n')} data-message-chip="tool-choice">
			<FiTool size={14} />
			<span className={labelClasses}>{name}</span>
		</div>
	);
}

interface MessageToolCallChipProps {
	call: UIToolCall;
	fullWidth?: boolean;
}

/**
 * Read‑only chip for an assistant-suggested tool call under the assistant bubble.
 */
function MessageToolCallChip({ call, fullWidth = false }: MessageToolCallChipProps) {
	const tmpCall: UIToolCall = {
		id: call.id || call.callID,
		callID: call.callID,
		name: call.name,
		arguments: call.arguments,
		choiceID: call.choiceID,
		type: call.type,
		status: 'pending',
	};
	const label = formatToolCallLabel(tmpCall);

	const statusLabel = call.status ? ` (${call.status})` : '';
	const title = `Suggested tool call: ${label}${statusLabel}`;

	const containerClasses = [
		'bg-base-200 text-base-content flex items-center gap-2 rounded-2xl px-2 py-0',
		fullWidth ? 'w-full' : 'shrink-0',
	]
		.filter(Boolean)
		.join(' ');

	const labelClasses = fullWidth ? 'min-w-0 flex-1 truncate' : 'max-w-64 truncate';

	return (
		<div className={containerClasses} title={title} data-message-chip="tool-suggested">
			<FiTool size={14} />
			<span className="text-base-content/60 text-[10px] uppercase">Suggested</span>
			<span className={labelClasses}>{label}</span>
		</div>
	);
}

interface MessageToolOutputChipProps {
	output: UIToolOutput;
	fullWidth?: boolean;
}

/**
 * Read‑only chip for a tool output that was attached to this user turn.
 * History chips are not interactive; the full output was already used
 * when the turn was sent.
 */
function MessageToolOutputChip({ output, fullWidth = false }: MessageToolOutputChipProps) {
	const prettyName = getPrettyToolName(output.name);
	const label = output.summary || `Result: ${prettyName}`;
	const titleLines = [label, `Tool: ${output.name}`, `Call ID: ${output.callID}`];
	const title = titleLines.join('\n');

	const containerClasses = [
		'bg-base-200 text-base-content flex items-center gap-2 rounded-2xl px-2 py-0',
		fullWidth ? 'w-full' : 'shrink-0',
	]
		.filter(Boolean)
		.join(' ');

	const labelClasses = fullWidth ? 'min-w-0 flex-1 truncate' : 'max-w-64 truncate';

	return (
		<div className={containerClasses} title={title} data-message-chip="tool-output">
			<FiTool size={14} />
			<span className="text-base-content/60 text-[10px] uppercase">Output</span>
			<span className={labelClasses}>{label}</span>
		</div>
	);
}

interface AttachmentsGroupChipProps {
	attachments: Attachment[];
}

function AttachmentsGroupChip({ attachments }: AttachmentsGroupChipProps) {
	const count = attachments.length;
	if (count === 0) return null;

	const menu = useMenuStore({ placement: 'bottom-start', focusLoop: true });

	const titleLines = ['Attachments', `${count} item${count === 1 ? '' : 's'} attached`];
	const title = titleLines.join('\n');

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-1 rounded-2xl px-2 py-0"
			title={title}
			data-message-chip="attachments-group"
		>
			<FiPaperclip size={14} />
			<span className="max-w-24 truncate">Attachments</span>
			<span className="text-base-content/70 text-[11px] whitespace-nowrap">{count}</span>

			<MenuButton
				store={menu}
				className="btn btn-ghost btn-xs px-0 py-0 shadow-none"
				aria-label="Show attachments for this message"
				title="Show attachments for this message"
			>
				<FiChevronDown size={14} />
			</MenuButton>

			<Menu
				store={menu}
				gutter={6}
				className="rounded-box bg-base-100 text-base-content border-base-300 z-50 max-h-72 min-w-65 overflow-y-auto border p-2 shadow-xl focus-visible:outline-none"
				autoFocusOnShow
			>
				<div className="text-base-content/70 mb-1 text-[11px] font-semibold">Attachments</div>

				{attachments.map((att, index) => (
					<MenuItem
						key={`${att.kind}:${att.label}:${index}`}
						store={menu}
						hideOnClick={false}
						className="data-active-item:bg-base-200 mb-1 rounded-xl last:mb-0"
					>
						<MessageAttachmentInfoChip attachment={att} fullWidth />
					</MenuItem>
				))}
			</Menu>
		</div>
	);
}

interface ToolChoicesGroupChipProps {
	tools: ToolStoreChoice[];
}

function ToolChoicesGroupChip({ tools }: ToolChoicesGroupChipProps) {
	const count = tools.length;
	if (count === 0) return null;

	const menu = useMenuStore({ placement: 'bottom-start', focusLoop: true });

	const titleLines = ['Tools', `${count} tool${count === 1 ? '' : 's'} used for this turn`];
	const title = titleLines.join('\n');

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-1 rounded-2xl px-2 py-0"
			title={title}
			data-message-chip="tools-group"
		>
			<FiTool size={14} />
			<span className="max-w-24 truncate">Tools</span>
			<span className="text-base-content/60 whitespace-nowrap"> {count}</span>

			<MenuButton
				store={menu}
				className="btn btn-ghost btn-xs px-0 py-0 shadow-none"
				aria-label="Show tools for this message"
				title="Show tools for this message"
			>
				<FiChevronDown size={14} />
			</MenuButton>

			<Menu
				store={menu}
				gutter={6}
				className="rounded-box bg-base-100 text-base-content border-base-300 z-50 max-h-72 min-w-65 overflow-y-auto border p-2 shadow-xl focus-visible:outline-none"
				autoFocusOnShow
			>
				<div className="text-base-content/70 mb-1 text-[11px] font-semibold">Tools</div>

				{tools.map(tool => (
					<MenuItem
						key={tool.toolID ?? `${tool.bundleID}-${tool.toolSlug}-${tool.toolVersion}`}
						store={menu}
						hideOnClick={false}
						className="data-active-item:bg-base-200 mb-1 rounded-xl last:mb-0"
					>
						<MessageToolChoiceChip tool={tool} fullWidth />
					</MenuItem>
				))}
			</Menu>
		</div>
	);
}

interface ToolOutputsGroupChipProps {
	outputs: UIToolOutput[];
}

function ToolOutputsGroupChip({ outputs }: ToolOutputsGroupChipProps) {
	const count = outputs.length;
	if (count === 0) return null;

	const menu = useMenuStore({ placement: 'bottom-start', focusLoop: true });

	const titleLines = ['Tool outputs', `${count} result${count === 1 ? '' : 's'} used for this turn`];
	const title = titleLines.join('\n');

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-1 rounded-2xl px-2 py-0"
			title={title}
			data-message-chip="tool-outputs-group"
		>
			<FiTool size={14} />
			<span className="max-w-24 truncate">Tool results</span>
			<span className="text-base-content/60 whitespace-nowrap">{count}</span>

			<MenuButton
				store={menu}
				className="btn btn-ghost btn-xs px-0 py-0 shadow-none"
				aria-label="Show tool results for this message"
				title="Show tool results for this message"
			>
				<FiChevronDown size={14} />
			</MenuButton>

			<Menu
				store={menu}
				gutter={6}
				className="rounded-box bg-base-100 text-base-content border-base-300 z-50 max-h-72 min-w-65 overflow-y-auto border p-2 shadow-xl focus-visible:outline-none"
				autoFocusOnShow
			>
				<div className="text-base-content/70 mb-1 text-[11px] font-semibold">Tool results</div>

				{outputs.map(out => (
					<MenuItem
						key={out.id}
						store={menu}
						hideOnClick={false}
						className="data-active-item:bg-base-200 mb-1 rounded-xl last:mb-0"
					>
						<MessageToolOutputChip output={out} fullWidth />
					</MenuItem>
				))}
			</Menu>
		</div>
	);
}

interface ToolCallsGroupChipProps {
	calls: UIToolCall[];
}

function ToolCallsGroupChip({ calls }: ToolCallsGroupChipProps) {
	const count = calls.length;
	if (count === 0) return null;

	const menu = useMenuStore({ placement: 'bottom-start', focusLoop: true });

	const titleLines = ['Suggested tool calls', `${count} suggestion${count === 1 ? '' : 's'} from assistant`];
	const title = titleLines.join('\n');

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-1 rounded-2xl px-2 py-0"
			title={title}
			data-message-chip="tool-calls-group"
		>
			<FiTool size={14} />
			<span className="max-w-24 truncate">Tool calls</span>
			<span className="text-base-content/60 whitespace-nowrap">{count}</span>

			<MenuButton
				store={menu}
				className="btn btn-ghost btn-xs px-0 py-0 shadow-none"
				aria-label="Show suggested tool calls for this message"
				title="Show suggested tool calls for this message"
			>
				<FiChevronDown size={14} />
			</MenuButton>

			<Menu
				store={menu}
				gutter={6}
				className="rounded-box bg-base-100 text-base-content border-base-300 z-50 max-h-72 min-w-65 overflow-y-auto border p-2 shadow-xl focus-visible:outline-none"
				autoFocusOnShow
			>
				<div className="text-base-content/70 mb-1 text-[11px] font-semibold">Suggested tool calls</div>

				{calls.map(call => (
					<MenuItem
						key={call.id || call.callID}
						store={menu}
						hideOnClick={false}
						className="data-active-item:bg-base-200 mb-1 rounded-xl last:mb-0"
					>
						<MessageToolCallChip call={call} fullWidth />
					</MenuItem>
				))}
			</Menu>
		</div>
	);
}

interface MessageAttachmentsBarProps {
	attachments?: Attachment[];
	toolChoices?: ToolStoreChoice[];
	toolCalls?: UIToolCall[];
	toolOutputs?: UIToolOutput[];
}

/**
 * Read‑only toolbar under a message bubble:
 * - For user messages: files, tool choices, and tool outputs.
 * - For assistant messages: files and suggested tool calls.
 *
 * Uses compact dropdown chips similar to the composer, but without
 * any remove / edit actions.
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

	if (!hasAttachments && !hasTools && !hasToolCalls && !hasToolOutputs) {
		return null;
	}

	return (
		<div className="border-base-200 flex min-h-8 max-w-full min-w-0 flex-1 items-center gap-1 overflow-x-auto border-t px-2 py-1 text-xs">
			{hasAttachments && <AttachmentsGroupChip attachments={attachments ?? []} />}

			{hasTools && <ToolChoicesGroupChip tools={toolChoices ?? []} />}

			{hasToolOutputs && <ToolOutputsGroupChip outputs={toolOutputs ?? []} />}

			{hasToolCalls && <ToolCallsGroupChip calls={toolCalls ?? []} />}
		</div>
	);
}
