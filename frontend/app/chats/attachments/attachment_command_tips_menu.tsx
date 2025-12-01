import { useMemo } from 'react';

import { FiChevronDown, FiChevronUp, FiMoreHorizontal } from 'react-icons/fi';

import { Menu, MenuButton, MenuItem, useMenuStore, useStoreState } from '@ariakit/react';

import { buildShortcutDisplay, type ShortcutConfig } from '@/lib/keyboard_shortcuts';

type TipKey = 'lastWins' | 'remove' | 'collapse';

const tipsText: Record<TipKey, string> = {
	lastWins: 'Last system prompt wins: when multiple templates add system prompts, the last one added becomes active.',
	remove: 'Removing a template also removes the system prompt that came from it (if it was created by that template).',
	collapse:
		'Collapse to text decouples removes the system prompts, vars, tools. The current user block with placeholders for vars is inserted as plain text',
};

interface CommandTipsMenuProps {
	shortcutConfig: ShortcutConfig;
}

const menuClasses =
	'rounded-box bg-base-100 text-base-content z-50 min-w-[240px] overflow-visible border border-base-300 p-1 shadow-xl';

const menuItemClasses =
	'flex items-center gap-2 rounded-xl px-2 py-1 text-xs outline-none transition-colors ' +
	'hover:bg-base-200 data-[active-item]:bg-base-300';

export function CommandTipsMenu({ shortcutConfig }: CommandTipsMenuProps) {
	const shortcutsMenu = useMenuStore({ placement: 'top-end' });
	const tipsMenu = useMenuStore({ placement: 'top-end' });
	const shortcutsOpen = useStoreState(shortcutsMenu, 'open');
	const tipsOpen = useStoreState(tipsMenu, 'open');
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

	return (
		<div className="flex items-center gap-1">
			{/* Keyboard shortcuts menu */}
			<MenuButton
				store={shortcutsMenu}
				className="btn btn-xs text-neutral-custom flex items-center gap-2 overflow-hidden border-none bg-transparent px-2 py-1 text-left shadow-none"
				title="Keyboard shortcuts"
				aria-label="Keyboard shortcuts"
			>
				<span className="text-xs font-normal whitespace-nowrap">Keyboard shortcuts</span>
				{shortcutsOpen ? (
					<FiChevronDown size={14} className="shrink-0" aria-hidden="true" />
				) : (
					<FiChevronUp size={14} className="shrink-0" aria-hidden="true" />
				)}
			</MenuButton>

			<Menu store={shortcutsMenu} gutter={8} className={menuClasses} autoFocusOnShow>
				{/* Chat group */}
				{chatShortcuts.length > 0 && (
					<>
						<div className="text-neutral-custom/70 px-3 pt-2 pb-1 text-[11px] tracking-wide uppercase">
							Chat shortcuts
						</div>
						{chatShortcuts.map(item => (
							<MenuItem key={item.action} className={menuItemClasses}>
								<span className="flex-1 text-left">{item.label}</span>
								<span className="text-neutral-custom ml-auto w-20 text-left text-[11px] whitespace-nowrap">
									{item.keys}
								</span>
							</MenuItem>
						))}
					</>
				)}

				{/* Insert group */}
				{insertShortcuts.length > 0 && (
					<>
						{chatShortcuts.length > 0 && <div className="border-base-200 mx-2 mt-1 border-t" />}
						<div className="text-neutral-custom/70 px-3 pt-2 pb-1 text-[11px] tracking-wide uppercase">
							Insert shortcuts
						</div>
						{insertShortcuts.map(item => (
							<MenuItem key={item.action} className={menuItemClasses}>
								<span className="flex-1 text-left">{item.label}</span>
								<span className="text-neutral-custom ml-auto w-20 text-left text-[11px] whitespace-nowrap">
									{item.keys}
								</span>
							</MenuItem>
						))}
					</>
				)}
			</Menu>

			{/* Additional input tips menu (with hover tooltips for descriptions) */}
			<MenuButton
				store={tipsMenu}
				className="btn btn-xs text-neutral-custom flex items-center gap-2 overflow-hidden border-none bg-transparent px-2 py-1 text-left shadow-none"
				title="Additional input tips"
				aria-label="Additional input tips"
			>
				<span className="text-xs font-normal whitespace-nowrap">Additional input tips</span>
				{tipsOpen ? (
					<FiChevronDown size={14} className="shrink-0" aria-hidden="true" />
				) : (
					<FiChevronUp size={14} className="shrink-0" aria-hidden="true" />
				)}
			</MenuButton>

			<Menu store={tipsMenu} gutter={8} className={menuClasses} autoFocusOnShow>
				{tips.map(tip => (
					<MenuItem key={tip.key} className={menuItemClasses}>
						<div className="tooltip tooltip-top w-full" data-tip={tip.description}>
							<div className="flex items-center gap-2">
								<span className="text-left font-medium">{tip.title}</span>
								<FiMoreHorizontal size={14} className="ml-auto opacity-60" aria-hidden="true" />
							</div>
						</div>
					</MenuItem>
				))}
			</Menu>
		</div>
	);
}
