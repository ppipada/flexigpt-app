import { FiAlertTriangle, FiPlay, FiTool, FiX } from 'react-icons/fi';

import type { ToolOutput } from '@/spec/tool';

import type { EditorToolCall } from '@/chats/tools/tool_editor_utils';
import { getPrettyToolName } from '@/chats/tools/tool_editor_utils';

interface ToolChipsComposerRowProps {
	toolCalls: EditorToolCall[];
	toolOutputs: ToolOutput[];
	isBusy: boolean;
	onRunToolCall: (id: string) => void | Promise<void>;
	onDiscardToolCall: (id: string) => void;
	onOpenOutput: (output: ToolOutput) => void;
	onRemoveOutput: (id: string) => void;
}

/**
 * Row of interactive tool-call and tool-output chips used in the composer.
 *
 * Order (left → right):
 *   - Pending / running / failed tool calls
 *   - Tool output chips
 */
export function ToolChipsComposerRow({
	toolCalls,
	toolOutputs,
	isBusy,
	onRunToolCall,
	onDiscardToolCall,
	onOpenOutput,
	onRemoveOutput,
}: ToolChipsComposerRowProps) {
	const visibleCalls = toolCalls.filter(toolCall => toolCall.status !== 'discarded' && toolCall.status !== 'succeeded');
	const hasAny = visibleCalls.length > 0 || toolOutputs.length > 0;
	if (!hasAny) return null;

	return (
		<div className="flex shrink-0 items-center gap-1">
			{visibleCalls.map(toolCall => (
				<ToolCallComposerChipView
					key={toolCall.id}
					toolCall={toolCall}
					isBusy={isBusy}
					onRun={() => {
						void onRunToolCall(toolCall.id);
					}}
					onDiscard={() => {
						onDiscardToolCall(toolCall.id);
					}}
				/>
			))}

			{toolOutputs.map(output => (
				<ToolOutputComposerChipView
					key={output.id}
					output={output}
					onOpen={() => {
						onOpenOutput(output);
					}}
					onRemove={() => {
						onRemoveOutput(output.id);
					}}
				/>
			))}
		</div>
	);
}

interface ToolCallComposerChipViewProps {
	toolCall: EditorToolCall;
	isBusy: boolean;
	onRun: () => void;
	onDiscard: () => void;
}

/**
 * Interactive chip for a single pending / running / failed tool call.
 * - "Run" button executes the tool once.
 * - "×" discards the suggestion from the composer only.
 */
function ToolCallComposerChipView({ toolCall, isBusy, onRun, onDiscard }: ToolCallComposerChipViewProps) {
	const label = getPrettyToolName(toolCall.name);
	const truncatedLabel = label.length > 64 ? `${label.slice(0, 61)}…` : label;

	const isRunning = toolCall.status === 'running';
	const isPending = toolCall.status === 'pending';
	const isFailed = toolCall.status === 'failed';

	const canRun = (isPending || isFailed) && !isBusy;

	const baseClasses =
		'bg-base-200 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0 border ' +
		'border-transparent hover:bg-base-300/80';

	const errorClasses = isFailed ? 'border-error/70 bg-error/5 text-error' : '';

	const titleLines: string[] = [`Suggested: ${label}`];
	if (toolCall.errorMessage && isFailed) {
		titleLines.push(`Error: ${toolCall.errorMessage}`);
	}
	const title = titleLines.join('\n');

	return (
		<div className={`${baseClasses} ${errorClasses}`} title={title} data-attachment-chip="tool-call">
			<FiTool size={14} className={isFailed ? 'text-error' : ''} />
			<span className="text-base-content/60 text-[10px] uppercase">Suggested</span>
			<span className="max-w-64 truncate">{truncatedLabel}</span>

			<div className="ml-auto flex items-center gap-2 p-0">
				{isRunning ? (
					<span className="loading loading-spinner loading-xs" aria-label="Running tool call" />
				) : (
					<button
						type="button"
						className={`btn btn-ghost btn-xs gap-0 p-0 shadow-none ${!canRun ? 'btn-disabled' : ''}`}
						onClick={onRun}
						disabled={!canRun}
						title={isFailed ? 'Retry this tool call' : 'Run this tool call'}
						aria-label={isFailed ? 'Retry tool call' : 'Run tool call'}
					>
						<FiPlay size={12} />
						<span className="ml-1 text-[11px]">Run</span>
					</button>
				)}

				{isFailed && (
					<FiAlertTriangle
						size={12}
						className="text-error"
						title={toolCall.errorMessage}
						aria-label="Tool call failed"
					/>
				)}

				<button
					type="button"
					className="btn btn-ghost btn-xs text-error p-0 shadow-none"
					onClick={onDiscard}
					title="Discard this tool call"
					aria-label="Discard tool call"
				>
					<FiX size={12} />
				</button>
			</div>
		</div>
	);
}

interface ToolOutputComposerChipViewProps {
	output: ToolOutput;
	onOpen: () => void;
	onRemove: () => void;
}

/**
 * Interactive chip for a single tool output in the composer.
 * - Click opens the full JSON/text in a modal.
 * - "×" discards the output from the next turn.
 */
function ToolOutputComposerChipView({ output, onOpen, onRemove }: ToolOutputComposerChipViewProps) {
	const label = getPrettyToolName(output.name);
	const truncatedLabel = label.length > 64 ? `${label.slice(0, 61)}…` : label;

	const titleLines = [label, `Tool: ${output.name}`, `Call ID: ${output.callID}`];
	const title = titleLines.join('\n');

	return (
		<div
			className="bg-base-200 text-base-content hover:bg-base-300/80 flex shrink-0 cursor-pointer items-center gap-2 rounded-2xl px-2 py-0"
			title={title}
			role="button"
			tabIndex={0}
			onClick={onOpen}
			onKeyDown={e => {
				if (e.key === 'Enter' || e.key === ' ') {
					e.preventDefault();
					onOpen();
				}
			}}
			data-attachment-chip="tool-output"
		>
			<FiTool size={14} />
			<span className="text-base-content/60 text-[10px] uppercase">Result</span>
			<span className="max-w-64 truncate">{truncatedLabel}</span>

			<button
				type="button"
				className="btn btn-ghost btn-xs text-error ml-1 px-1 py-0 shadow-none"
				onClick={e => {
					e.stopPropagation();
					onRemove();
				}}
				title="Discard this tool output"
				aria-label="Discard tool output"
			>
				<FiX size={12} />
			</button>
		</div>
	);
}
