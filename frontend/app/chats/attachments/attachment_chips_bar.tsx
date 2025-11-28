import { FiChevronDown, FiChevronUp, FiFileText, FiImage, FiLink, FiTool, FiX } from 'react-icons/fi';

import { Menu, MenuButton, MenuItem, useMenuStore, useStoreState } from '@ariakit/react';
import type { PlateEditor } from 'platejs/react';
import { useEditorRef } from 'platejs/react';

import { AttachmentKind, AttachmentMode } from '@/spec/attachment';

import {
	ATTACHMENT_MODE_DESC,
	ATTACHMENT_MODE_LABELS,
	type EditorAttachment,
	editorAttachmentKey,
	getEditorAttachmentPath,
} from '@/chats/attachments/editor_attachment_utils';
import { getToolNodesWithPath, removeToolByKey, toolIdentityKey } from '@/chats/attachments/tool_editor_utils';

/**
 * Shared styles for the small "mode" menu on each attachment chip.
 * This keeps the same visual language as other menus, but shrink-wraps
 * to the max width of its items.
 */
const modeMenuClasses =
	'rounded-box bg-base-100 text-base-content z-50 ' +
	'max-h-72 w-max min-w-0 overflow-y-auto ' +
	'border border-base-300 p-1 shadow-xl';

const modeMenuItemClasses =
	'flex items-center gap-2 rounded-xl px-2 py-1 text-xs outline-none transition-colors ' +
	'hover:bg-base-200 data-[active-item]:bg-base-300 whitespace-nowrap';

interface AttachmentChipsBarProps {
	attachments: EditorAttachment[];
	onRemoveAttachment: (att: EditorAttachment) => void;
	onChangeAttachmentMode: (att: EditorAttachment, mode: AttachmentMode) => void;
}

interface AttachmentModeMenuProps {
	attachment: EditorAttachment;
	onChangeAttachmentMode: (att: EditorAttachment, mode: AttachmentMode) => void;
}

function getAttachmentModeLabel(mode: AttachmentMode): string {
	// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
	return ATTACHMENT_MODE_LABELS[mode] ?? mode;
}

function getAttachmentModeTooltip(mode: AttachmentMode): string {
	if (mode === AttachmentMode.notReadable) {
		return 'This attachment is not readable (unsupported or inaccessible content).';
	}
	const l = getAttachmentModeLabel(mode);
	// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
	const desc = ATTACHMENT_MODE_DESC[mode] ?? l;
	return `Mode: ${l} (${desc})`;
}

function getAttachmentModePillClasses(mode: AttachmentMode, interactive: boolean): string {
	const base =
		'inline-flex items-center gap-1 rounded-full border px-2 py-[1px] text-[10px] leading-tight ' + 'transition-colors';

	const interactiveClasses = interactive
		? 'cursor-pointer focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-base-300'
		: '';

	const isError = mode === AttachmentMode.notReadable;

	if (isError) {
		return [base, interactive ? 'hover:bg-error/20' : '', 'border-error/40 bg-error/10 text-error', interactiveClasses]
			.filter(Boolean)
			.join(' ');
	}

	return [
		base,
		interactive ? 'hover:bg-base-200' : '',
		'border-base-300 bg-base-100 text-base-content/80',
		interactiveClasses,
	]
		.filter(Boolean)
		.join(' ');
}

/**
 * Per-attachment mode selector using Ariakit Menu.
 * Replaces the native <select>, while keeping a small pill-like trigger
 * and a shrink-wrapped dropdown.
 */
function AttachmentModeMenu({ attachment, onChangeAttachmentMode }: AttachmentModeMenuProps) {
	const menu = useMenuStore({
		placement: 'bottom-start',
	});

	const open = useStoreState(menu, 'open');

	const currentLabel = getAttachmentModeLabel(attachment.mode);
	const tooltip = getAttachmentModeTooltip(attachment.mode);

	const ChevronIcon = open ? FiChevronDown : FiChevronUp;

	return (
		<>
			<MenuButton
				store={menu}
				className={getAttachmentModePillClasses(attachment.mode, true)}
				aria-label="Change attachment mode"
				title={tooltip}
				data-attachment-mode-button
			>
				<span>{currentLabel}</span>
				<ChevronIcon className="shrink-0" size={10} aria-hidden="true" />
			</MenuButton>

			<Menu store={menu} gutter={4} className={modeMenuClasses} data-attachment-mode-menu autoFocusOnShow>
				{attachment.availableModes.map(mode => {
					const label = getAttachmentModeLabel(mode);
					const modeTooltip = getAttachmentModeTooltip(mode);
					const isActive = mode === attachment.mode;
					const isError = mode === AttachmentMode.notReadable;

					return (
						<MenuItem
							key={mode}
							className={modeMenuItemClasses}
							onClick={() => {
								onChangeAttachmentMode(attachment, mode);
								menu.hide();
							}}
							aria-pressed={isActive}
							title={modeTooltip}
						>
							<span className={[isActive ? 'font-medium' : '', isError ? 'text-error' : ''].filter(Boolean).join(' ')}>
								{label}
							</span>
						</MenuItem>
					);
				})}
			</Menu>
		</>
	);
}

/**
 * Scrollable chips bar for attachments and tools, shown inside the Plate editor
 */
export function AttachmentChipsBar({
	attachments,
	onRemoveAttachment,
	onChangeAttachmentMode,
}: AttachmentChipsBarProps) {
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

				const icon =
					att.kind === AttachmentKind.image ? (
						<FiImage />
					) : att.kind === AttachmentKind.url ? (
						<FiLink />
					) : (
						<FiFileText />
					);

				const titleParts = [`${att.kind} attachment: ${att.label}`];
				if (path) titleParts.push(`(${path})`);
				if (ATTACHMENT_MODE_LABELS[att.mode]) {
					titleParts.push(`[${ATTACHMENT_MODE_LABELS[att.mode]}]`);
				}

				return (
					<div
						key={key}
						className="bg-base-200 hover:bg-base-300/80 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0"
						title={titleParts.join(' ')}
						data-attachment-chip="attachment"
					>
						{icon}
						<span className="max-w-64 truncate">{label}</span>

						{/* Mode pill / Ariakit menu */}
						{att.availableModes.length > 1 ? (
							<AttachmentModeMenu attachment={att} onChangeAttachmentMode={onChangeAttachmentMode} />
						) : att.availableModes.length === 1 ? (
							<span
								className={getAttachmentModePillClasses(att.mode, false)}
								title={getAttachmentModeTooltip(att.mode)}
								data-attachment-mode-pill
							>
								{getAttachmentModeLabel(att.mode)}
							</span>
						) : null}

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
