import { useEffect, useMemo, useRef, useState } from 'react';

import { FiChevronDown, FiChevronUp, FiMoreHorizontal } from 'react-icons/fi';

import { buildShortcutDisplay, type ShortcutConfig } from '@/lib/keyboard_shortcuts';

type TipKey = 'lastWins' | 'remove' | 'collapse';

const tipsText: Record<TipKey, string> = {
	lastWins: 'Last system prompt wins: when multiple templates add system prompts, the last one added becomes active.',
	remove: 'Removing a template also removes the system prompt that came from it (if it was created by that template).',
	collapse:
		'Collapse to text decouples removes the system prompts, vars, tools. The current user block with placeholders for vars is inserted as plain text',
};

interface CommandTipsBarProps {
	shortcutConfig: ShortcutConfig;
}

export function CommandTipsBar({ shortcutConfig }: CommandTipsBarProps) {
	// Dropdown state/refs
	const [isShortcutsOpen, setIsShortcutsOpen] = useState(false);
	const [isTipsOpen, setIsTipsOpen] = useState(false);
	const shortcutsDetailsRef = useRef<HTMLDetailsElement>(null);
	const tipsDetailsRef = useRef<HTMLDetailsElement>(null);

	const tips = useMemo(
		() => [
			{ key: 'lastWins' as TipKey, title: 'Last system prompt wins', description: tipsText.lastWins },
			{ key: 'remove' as TipKey, title: 'Removing a template', description: tipsText.remove },
			{ key: 'collapse' as TipKey, title: 'Collapse to text decouples', description: tipsText.collapse },
		],
		[]
	);

	// All shortcuts (global + insert) from central config
	const shortcutItems = useMemo(() => buildShortcutDisplay(shortcutConfig), [shortcutConfig]);

	const chatShortcuts = shortcutItems.filter(i => i.group === 'Chat');
	const insertShortcuts = shortcutItems.filter(i => i.group === 'Insert');

	// Close dropdowns on outside click or Escape
	useEffect(() => {
		const handlePointerDown = (e: MouseEvent | TouchEvent) => {
			if (!isShortcutsOpen && !isTipsOpen) return;
			const target = e.target as Node;

			const shortcutsNode = shortcutsDetailsRef.current;
			const tipsNode = tipsDetailsRef.current;

			if (isShortcutsOpen && shortcutsNode && !shortcutsNode.contains(target)) {
				shortcutsNode.open = false;
				setIsShortcutsOpen(false);
			}
			if (isTipsOpen && tipsNode && !tipsNode.contains(target)) {
				tipsNode.open = false;
				setIsTipsOpen(false);
			}
		};

		const handleKeyDown = (e: KeyboardEvent) => {
			if (e.key !== 'Escape') return;

			const shortcutsNode = shortcutsDetailsRef.current;
			const tipsNode = tipsDetailsRef.current;

			if (isShortcutsOpen && shortcutsNode) {
				shortcutsNode.open = false;
				setIsShortcutsOpen(false);
			}
			if (isTipsOpen && tipsNode) {
				tipsNode.open = false;
				setIsTipsOpen(false);
			}
		};

		document.addEventListener('mousedown', handlePointerDown);
		document.addEventListener('touchstart', handlePointerDown, { passive: true });
		document.addEventListener('keydown', handleKeyDown);
		return () => {
			document.removeEventListener('mousedown', handlePointerDown);
			document.removeEventListener('touchstart', handlePointerDown);
			document.removeEventListener('keydown', handleKeyDown);
		};
	}, [isShortcutsOpen, isTipsOpen]);

	return (
		<div className="text-neutral-custom flex items-center justify-between px-2 py-0 text-xs">
			{/* Additional shortcuts dropdown (left) */}
			<details
				ref={shortcutsDetailsRef}
				className="dropdown dropdown-top dropdown-start"
				open={isShortcutsOpen}
				onToggle={e => {
					setIsShortcutsOpen((e.currentTarget as HTMLDetailsElement).open);
				}}
			>
				<summary
					className="btn btn-xs text-neutral-custom flex items-center gap-2 overflow-hidden border-none text-left"
					title="Additional keyboard shortcuts"
					aria-expanded={isShortcutsOpen}
				>
					<span className="text-xs font-normal whitespace-nowrap">Keyboard shortcuts</span>
					{isShortcutsOpen ? (
						<FiChevronDown size={14} className="shrink-0" />
					) : (
						<FiChevronUp size={14} className="shrink-0" />
					)}
				</summary>

				<ul className="dropdown-content max-w-[80vw] min-w-[24vw] list-none overflow-visible rounded-xl p-0 shadow-xl">
					{/* Chat group */}
					{chatShortcuts.length > 0 && (
						<>
							<li className="text-neutral-custom/70 px-3 pt-2 pb-1 text-[11px] tracking-wide uppercase">
								Chat shortcuts
							</li>
							{chatShortcuts.map(item => (
								<li key={item.action} className="m-0 p-0 text-xs">
									<div className="hover:bg-base-200 flex cursor-default items-center gap-2 rounded-lg px-3 py-1.5">
										<span className="text-left">{item.label}</span>
										<span className="text-neutral-custom ml-auto text-[11px] whitespace-nowrap">{item.keys}</span>
									</div>
								</li>
							))}
						</>
					)}

					{/* Insert group */}
					{insertShortcuts.length > 0 && (
						<>
							{chatShortcuts.length > 0 && <li className="border-base-200 mx-2 mt-1 border-t" />}
							<li className="text-neutral-custom/70 px-3 pt-2 pb-1 text-[11px] tracking-wide uppercase">
								Insert shortcuts
							</li>
							{insertShortcuts.map(item => (
								<li key={item.action} className="m-0 p-0 text-xs">
									<div className="hover:bg-base-200 flex cursor-default items-center gap-2 rounded-lg px-3 py-1.5">
										<span className="text-left">{item.label}</span>
										<span className="text-neutral-custom ml-auto text-[11px] whitespace-nowrap">{item.keys}</span>
									</div>
								</li>
							))}
						</>
					)}
				</ul>
			</details>

			{/* Additional input tips dropdown (right) */}
			<details
				ref={tipsDetailsRef}
				className="dropdown dropdown-top dropdown-end"
				open={isTipsOpen}
				onToggle={e => {
					setIsTipsOpen((e.currentTarget as HTMLDetailsElement).open);
				}}
			>
				<summary
					className="btn btn-xs text-neutral-custom flex items-center gap-2 overflow-hidden border-none text-left shadow-none"
					title="Additional input tips"
					aria-expanded={isTipsOpen}
				>
					<span className="text-xs font-normal whitespace-nowrap">Additional input tips</span>
					{isTipsOpen ? (
						<FiChevronDown size={14} className="shrink-0" />
					) : (
						<FiChevronUp size={14} className="shrink-0" />
					)}
				</summary>

				<ul className="dropdown-content max-w-[80vw] min-w-[20vw] list-none overflow-visible rounded-xl p-0 shadow-xl">
					{tips.map(tip => (
						<li key={tip.key} className="m-0 p-0 text-xs">
							<div className="tooltip tooltip-top w-full" data-tip={tip.description}>
								<div className="hover:bg-base-200 flex cursor-default items-center gap-2 rounded-lg p-2">
									<span className="text-left font-medium">{tip.title}</span>
									<FiMoreHorizontal size={14} className="ml-auto opacity-60" aria-hidden="true" />
								</div>
							</div>
						</li>
					))}
				</ul>
			</details>
		</div>
	);
}
