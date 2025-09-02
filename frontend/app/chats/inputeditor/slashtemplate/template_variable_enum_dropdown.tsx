// components/EnumDropdownInline.tsx
import React from 'react';

import { createPortal } from 'react-dom';

import { FiCheck, FiChevronDown, FiChevronUp, FiX } from 'react-icons/fi';

type EnumDropdownInlineProps = {
	options: string[];
	value?: string;
	onChange: (val: string | undefined) => void;

	placeholder?: string;
	clearLabel?: string;
	clearable?: boolean;
	disabled?: boolean;
	size?: 'xs' | 'sm' | 'md';
	triggerClassName?: string;
	withinSlate?: boolean;
	autoOpen?: boolean;
	onCancel?: () => void;

	placement?: 'top' | 'bottom' | 'auto';
	minWidthPx?: number;
	menuMaxHeightPx?: number;
};

export function EnumDropdownInline({
	options,
	value,
	onChange,
	placeholder = '-- select --',
	clearLabel = 'Clear',
	clearable = true,
	disabled = false,
	size = 'xs',
	triggerClassName,
	withinSlate = false,
	autoOpen = false,
	onCancel,
	placement = 'auto',
	minWidthPx = 176,
	menuMaxHeightPx = 240,
}: EnumDropdownInlineProps) {
	const [open, setOpen] = React.useState(false);
	const triggerRef = React.useRef<HTMLButtonElement | null>(null);
	const menuRef = React.useRef<HTMLDivElement | null>(null);

	// Positioning: use absolute + scroll offsets (more robust with modals/backdrops/filters)
	const [style, setStyle] = React.useState<React.CSSProperties>({
		position: 'absolute',
		top: -9999,
		left: -9999,
		zIndex: 9999,
		visibility: 'hidden',
	});

	const sizeCls = size === 'md' ? 'btn-md' : size === 'sm' ? 'btn-sm' : 'btn-xs';
	const defaultTrigger = `btn btn-ghost ${sizeCls} font-normal w-40 min-w-24 justify-between truncate bg-transparent`;

	const display = value === undefined || value === '' ? placeholder : value;

	const stopForSlate = (e: React.SyntheticEvent) => {
		if (withinSlate) {
			// Important in Slate/Plate: prevent default and stop propagation
			e.preventDefault();
			e.stopPropagation();
		}
	};

	const updatePosition = React.useCallback(() => {
		if (!open) return;
		const anchor = triggerRef.current;
		const menu = menuRef.current;
		if (!anchor || !menu) return;

		// Temporarily show the menu to measure
		const prevVis = menu.style.visibility;
		const prevTop = menu.style.top;
		const prevLeft = menu.style.left;
		const prevPos = menu.style.position;

		menu.style.visibility = 'hidden';
		menu.style.position = 'absolute';
		menu.style.top = '0px';
		menu.style.left = '0px';

		const GAP = 6;
		const rect = anchor.getBoundingClientRect();
		const menuRect = menu.getBoundingClientRect();

		// Decide placement
		const spaceAbove = rect.top;
		const spaceBelow = window.innerHeight - rect.bottom;
		const preferTop = placement === 'top' || (placement === 'auto' && spaceAbove > spaceBelow);

		let top = preferTop ? rect.top + window.scrollY - menuRect.height - GAP : rect.bottom + window.scrollY + GAP;

		// flip if overflowing
		if (preferTop && top < window.scrollY + 8 && spaceBelow >= menuRect.height + GAP) {
			top = rect.bottom + window.scrollY + GAP;
		} else if (
			!preferTop &&
			top + menuRect.height > window.scrollY + window.innerHeight - 8 &&
			spaceAbove >= menuRect.height + GAP
		) {
			top = rect.top + window.scrollY - menuRect.height - GAP;
		}

		let left = rect.left + window.scrollX;
		const maxLeft = window.scrollX + window.innerWidth - menuRect.width - 8;
		if (left > maxLeft) left = Math.max(window.scrollX + 8, maxLeft);

		setStyle({
			position: 'absolute',
			top,
			left,
			minWidth: Math.max(minWidthPx, rect.width),
			zIndex: 9999,
			visibility: 'visible',
		});

		// restore temp mutations (React owns styles via state)
		menu.style.visibility = prevVis;
		menu.style.position = prevPos;
		menu.style.top = prevTop;
		menu.style.left = prevLeft;
	}, [open, placement, minWidthPx]);

	const close = React.useCallback(
		(cancel = true) => {
			setOpen(false);
			if (cancel) onCancel?.();
		},
		[onCancel]
	);

	// Ensure autoOpen opens on mount/update (not only as initial state)
	React.useEffect(() => {
		if (autoOpen) setOpen(true);
	}, [autoOpen]);

	// Use layout effect for first position calc to avoid visible jump
	React.useLayoutEffect(() => {
		if (!open) return;
		updatePosition();
	}, [open, updatePosition]);

	// Reposition and outside click handling
	React.useEffect(() => {
		if (!open) return;

		// Reposition on scroll/resize
		const onReposition = () => {
			updatePosition();
		};

		// Close on ESC
		const onKey = (e: KeyboardEvent) => {
			if (e.key === 'Escape') {
				e.preventDefault();
				close(true);
			}
		};

		// Close on outside click
		const onOutside = (e: MouseEvent | PointerEvent) => {
			const path = e.composedPath();
			const a = triggerRef.current;
			const m = menuRef.current;
			if (!a || !m) return;
			if (!path.includes(a) && !path.includes(m)) {
				close(true);
			}
		};

		window.addEventListener('scroll', onReposition, true);
		window.addEventListener('resize', onReposition);
		document.addEventListener('keydown', onKey, true);
		document.addEventListener('pointerdown', onOutside, true);

		return () => {
			window.removeEventListener('scroll', onReposition, true);
			window.removeEventListener('resize', onReposition);
			document.removeEventListener('keydown', onKey, true);
			document.removeEventListener('pointerdown', onOutside, true);
		};
	}, [open, updatePosition, close]);

	// Focus trigger for keyboard users
	React.useEffect(() => {
		if (!autoOpen) return;
		requestAnimationFrame(() => triggerRef.current?.focus());
	}, [autoOpen]);

	return (
		<>
			<button
				ref={triggerRef}
				type="button"
				className={triggerClassName ?? defaultTrigger}
				aria-haspopup="listbox"
				aria-expanded={open}
				aria-label="Open selection menu"
				title="Open selection"
				disabled={disabled}
				onMouseDown={stopForSlate}
				onPointerDown={stopForSlate}
				onClick={e => {
					stopForSlate(e);
					if (disabled) return;
					setOpen(o => !o);
				}}
				onKeyDown={e => {
					stopForSlate(e);
					if (disabled) return;
					if (e.key === 'Enter' || e.key === ' ') {
						e.preventDefault();
						setOpen(true);
					} else if (e.key === 'Escape') {
						e.preventDefault();
						close(true);
					} else if (e.key === 'ArrowDown') {
						e.preventDefault();
						setOpen(true);
					}
				}}
			>
				<span className={`truncate font-normal ${size === 'xs' ? 'text-xs' : size === 'sm' ? 'text-sm' : 'text-base'}`}>
					{display}
				</span>
				{open ? <FiChevronDown size={12} /> : <FiChevronUp size={12} />}
			</button>

			{open &&
				createPortal(
					<div
						ref={menuRef}
						style={style}
						className="bg-base-100 rounded-xl border p-1 shadow"
						onMouseDown={stopForSlate}
						onPointerDown={stopForSlate}
						onClick={stopForSlate}
					>
						<ul
							role="listbox"
							className="w-full"
							style={{ maxHeight: menuMaxHeightPx, overflow: 'auto', minWidth: style.minWidth }}
						>
							{options.map(opt => {
								const isActive = (value ?? '') === opt;
								return (
									<li key={opt} className="w-full">
										<button
											type="button"
											role="option"
											aria-selected={isActive}
											className={`hover:bg-base-200 w-full justify-start rounded-lg px-2 py-1 text-left font-normal ${
												size === 'xs' ? 'text-xs' : size === 'sm' ? 'text-sm' : 'text-base'
											}`}
											onClick={e => {
												stopForSlate(e);
												onChange(opt);
												setOpen(false);
											}}
										>
											<span className="truncate">{opt}</span>
											{isActive && <FiCheck size={12} className="ml-2 inline-block" />}
										</button>
									</li>
								);
							})}
							{clearable && (
								<li className="w-full">
									<button
										type="button"
										className={`text-error hover:bg-base-200 flex w-full items-center justify-center gap-1 rounded-lg px-2 py-1 text-left ${
											size === 'xs' ? 'text-xs' : size === 'sm' ? 'text-sm' : 'text-base'
										}`}
										onClick={e => {
											stopForSlate(e);
											onChange(undefined);
											setOpen(false);
										}}
									>
										<FiX /> {clearLabel}
									</button>
								</li>
							)}
						</ul>
					</div>,
					document.body
				)}
		</>
	);
}
