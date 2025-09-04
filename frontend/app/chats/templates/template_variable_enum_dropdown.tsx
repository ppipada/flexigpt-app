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
	const optionRefs = React.useRef<(HTMLButtonElement | null)[]>([]);
	const [activeIndex, setActiveIndex] = React.useState<number>(-1);
	const listboxId = React.useId();

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
			// Return focus to trigger for accessibility
			requestAnimationFrame(() => {
				try {
					triggerRef.current?.focus();
				} catch {
					/*swallow */
				}
			});
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

	// Set initial active item and focus when menu opens
	React.useEffect(() => {
		if (!open) return;
		const selectedIdx = value ? options.findIndex(opt => opt === value) : -1;
		const idx = selectedIdx >= 0 ? selectedIdx : 0;
		setActiveIndex(idx);
		requestAnimationFrame(() => {
			const btn = optionRefs.current[idx];
			if (btn) {
				btn.focus();
			} else {
				// fallback
				try {
					menuRef.current?.focus();
				} catch {
					/*swallow*/
				}
			}
		});
	}, [open, options, value]);

	// Keep active option in view as user navigates with keyboard
	React.useEffect(() => {
		if (!open) return;
		const btn = optionRefs.current[activeIndex];
		try {
			btn?.scrollIntoView({ block: 'nearest' });
		} catch {
			// noop
		}
	}, [activeIndex, open]);

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

	const onMenuKeyDown = (e: React.KeyboardEvent) => {
		if (withinSlate) {
			e.stopPropagation();
		}
		const maxIdx = options.length - 1;
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			const next = Math.min(maxIdx, activeIndex < 0 ? 0 : activeIndex + 1);
			setActiveIndex(next);
			requestAnimationFrame(() => optionRefs.current[next]?.focus());
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			const next = Math.max(0, activeIndex < 0 ? 0 : activeIndex - 1);
			setActiveIndex(next);
			requestAnimationFrame(() => optionRefs.current[next]?.focus());
		} else if (e.key === 'Home') {
			e.preventDefault();
			setActiveIndex(0);
			requestAnimationFrame(() => optionRefs.current[0]?.focus());
		} else if (e.key === 'End') {
			e.preventDefault();
			setActiveIndex(maxIdx);
			requestAnimationFrame(() => optionRefs.current[maxIdx]?.focus());
		} else if (e.key === 'Enter') {
			e.preventDefault();
			if (activeIndex >= 0 && activeIndex <= maxIdx) {
				const opt = options[activeIndex];
				onChange(opt);
				close(false);
			}
		} else if (e.key === 'Escape') {
			e.preventDefault();
			close(true);
		}
	};

	return (
		<>
			<button
				ref={triggerRef}
				type="button"
				className={triggerClassName ?? defaultTrigger}
				aria-haspopup="listbox"
				aria-expanded={open}
				aria-label="Open selection menu"
				aria-controls={open ? listboxId : undefined}
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
					} else if (e.key === 'ArrowUp') {
						// Open and let initial focus go to selected or first item
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
						tabIndex={-1}
						onMouseDown={stopForSlate}
						onPointerDown={stopForSlate}
						onClick={stopForSlate}
						onKeyDown={onMenuKeyDown}
					>
						<ul
							role="listbox"
							id={listboxId}
							className="w-full p-1"
							style={{ maxHeight: menuMaxHeightPx, overflow: 'auto', minWidth: style.minWidth }}
						>
							{options.map((opt, idx) => {
								const isActive = (value ?? '') === opt;
								const isFocused = idx === activeIndex;

								return (
									<li key={opt} className="w-full">
										<button
											type="button"
											role="option"
											aria-selected={isActive}
											ref={el => {
												optionRefs.current[idx] = el;
											}}
											className={`w-full justify-start rounded-lg px-2 py-1 text-left font-normal ${
												size === 'xs' ? 'text-xs' : size === 'sm' ? 'text-sm' : 'text-base'
											} ${isFocused ? 'bg-base-200' : 'hover:bg-base-200'}`}
											onClick={e => {
												stopForSlate(e);
												onChange(opt);
												close(false);
											}}
										>
											<div className="flex justify-between">
												<span className="truncate">{opt}</span>
												{isActive && <FiCheck size={12} className="ml-2 inline-block" />}
											</div>
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
											close(false);
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
