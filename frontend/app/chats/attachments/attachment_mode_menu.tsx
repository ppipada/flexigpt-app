import { FiChevronDown, FiChevronUp } from 'react-icons/fi';

import { Menu, MenuButton, MenuItem, useMenuStore, useStoreState } from '@ariakit/react';

import { ATTACHMENT_MODE_DESC, ATTACHMENT_MODE_LABELS, AttachmentMode } from '@/spec/attachment';

import { type EditorAttachment } from '@/chats/attachments/editor_attachment_utils';

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

interface AttachmentModeMenuProps {
	attachment: EditorAttachment;
	onChangeAttachmentMode: (att: EditorAttachment, mode: AttachmentMode) => void;
}

export function getAttachmentModeLabel(mode: AttachmentMode): string {
	// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
	return ATTACHMENT_MODE_LABELS[mode] ?? mode;
}

export function getAttachmentModeTooltip(mode: AttachmentMode): string {
	if (mode === AttachmentMode.notReadable) {
		return 'This attachment could not be read (unsupported type, too large, or inaccessible).';
	}

	const desc = ATTACHMENT_MODE_DESC[mode];
	if (desc) {
		// Tooltip focuses on *extra* explanation; the pill text already shows the label.
		return desc;
	}
	return getAttachmentModeLabel(mode);
}

export function getAttachmentModePillClasses(mode: AttachmentMode, interactive: boolean): string {
	const base =
		'inline-flex items-center gap-1 rounded-full border px-2 py-[1px] text-xs leading-tight ' + 'transition-colors';

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
export function AttachmentModeMenu({ attachment, onChangeAttachmentMode }: AttachmentModeMenuProps) {
	const menu = useMenuStore({
		placement: 'bottom-start',
		focusLoop: true,
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
