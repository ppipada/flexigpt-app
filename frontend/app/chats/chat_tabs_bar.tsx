/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { memo, useCallback, useEffect, useMemo, useState } from 'react';

import { FiEdit2, FiPlus, FiX } from 'react-icons/fi';

import type { TabStore } from '@ariakit/react/tab';
import { Tab, TabList } from '@ariakit/react/tab';

import { sanitizeConversationTitle } from '@/lib/text_utils';

import { DownloadButton } from '@/components/download_button';

interface ChatTabBarItem {
	tabId: string;
	title: string;
	isBusy: boolean;
	isEmpty: boolean; // messages.length === 0
	renameEnabled: boolean; // messages.length > 0
}

interface ChatTabsBarProps {
	store: TabStore;
	selectedTabId: string;
	tabs: ChatTabBarItem[];

	maxTabs: number;
	onNewTab: () => void;
	onCloseTab: (tabId: string) => void;
	onRenameTab: (tabId: string, newTitle: string) => void;
	getConversationForExport: () => Promise<string>;
}

export const ChatTabsBar = memo(function ChatTabsBar({
	store,
	selectedTabId,
	tabs,
	maxTabs,
	onNewTab,
	onCloseTab,
	onRenameTab,
	getConversationForExport,
}: ChatTabsBarProps) {
	const active = useMemo(() => tabs.find(t => t.tabId === selectedTabId) ?? tabs[0], [tabs, selectedTabId]);

	const [editingTabId, setEditingTabId] = useState<string | null>(null);
	const [draftTitle, setDraftTitle] = useState('');

	useEffect(() => {
		if (editingTabId) return;
		setDraftTitle(active?.title ?? '');
	}, [active?.title, editingTabId]);

	// Only active tab can be renamed; switching tabs cancels edit mode.
	useEffect(() => {
		if (editingTabId && editingTabId !== selectedTabId) setEditingTabId(null);
	}, [editingTabId, selectedTabId]);

	const finishRename = useCallback(() => {
		if (!editingTabId) return;

		const cleaned = sanitizeConversationTitle(draftTitle.trim());
		if (cleaned && cleaned !== (active?.title ?? '')) {
			onRenameTab(editingTabId, cleaned);
		}
		setEditingTabId(null);
	}, [active?.title, draftTitle, editingTabId, onRenameTab]);

	const elements: React.ReactNode[] = [];

	for (const t of tabs) {
		const isActive = t.tabId === selectedTabId;
		const canRename = isActive && t.renameEnabled && !t.isBusy;

		elements.push(
			<Tab
				key={t.tabId}
				store={store}
				id={t.tabId}
				// render as <div> so we can safely place an <input> inside the tab (no <input> inside <button>)
				render={<div />}
				className={[
					'relative flex h-full w-44 items-center p-0',
					'select-none',
					'focus-visible:outline-primary focus-visible:outline focus-visible:outline-offset-2',
					// Firefox-ish feel: rounded top + active lifted
					isActive
						? 'bg-base-100 text-base-content border-base-300 rounded-t-xl border shadow-sm'
						: 'bg-base-200/80 text-base-content/80 hover:bg-base-200 border-0',
				].join(' ')}
			>
				{/* Title / Rename */}
				<div className="min-w-0 flex-1 px-2 text-sm">
					{isActive && editingTabId === t.tabId ? (
						<input
							autoFocus
							value={draftTitle}
							onChange={e => {
								setDraftTitle(e.target.value);
							}}
							onBlur={finishRename}
							onKeyDown={e => {
								e.stopPropagation();
								if (e.key === 'Enter') finishRename();
								if (e.key === 'Escape') setEditingTabId(null);
							}}
							onMouseDown={e => {
								e.stopPropagation();
							}}
							className="input input-sm bg-base-100 w-full p-0"
						/>
					) : (
						<div className="flex min-w-0" title={t.title}>
							<span className="truncate">{t.title}</span>
						</div>
					)}
				</div>

				{/* Right end: spinner OR rename icon in same slot, then close */}
				<div className="flex items-center gap-1 pr-1">
					{(t.isBusy || canRename) && (
						<div className="flex w-6 shrink-0 items-center justify-center">
							{t.isBusy ? (
								<span className="loading loading-spinner loading-xs" aria-label="Busy" />
							) : canRename ? (
								<button
									type="button"
									className="btn btn-ghost btn-xs btn-circle p-0 opacity-70 hover:opacity-100"
									aria-label="Rename tab"
									title="Rename"
									onMouseDown={e => {
										e.stopPropagation();
									}}
									onClick={e => {
										e.stopPropagation();
										setEditingTabId(t.tabId);
										setDraftTitle(t.title);
									}}
								>
									<FiEdit2 size={14} />
								</button>
							) : null}
						</div>
					)}
					<button
						type="button"
						className="btn btn-ghost btn-xs btn-circle shrink-0 p-0 opacity-80 hover:opacity-100"
						aria-label="Close tab"
						title="Close tab"
						onMouseDown={e => {
							e.stopPropagation();
						}}
						onClick={e => {
							e.stopPropagation();
							onCloseTab(t.tabId);
						}}
					>
						<FiX size={14} />
					</button>
				</div>
			</Tab>
		);
	}

	return (
		<div className="w-full min-w-0">
			{/* Tab strip with bottom border line */}
			<div className="border-base-300 flex w-full items-center gap-2 overflow-hidden border-b bg-inherit">
				<div className="flex min-w-0 flex-1 flex-nowrap overflow-auto">
					<TabList store={store} aria-label="Chat tabs" className="flex min-w-0 items-end gap-0 py-0 pr-1">
						{elements}
					</TabList>

					<div
						className="tooltip tooltip-left px-1 py-0"
						data-tip={tabs.length >= maxTabs ? `New chat (max ${maxTabs} tabs)` : 'New chat'}
					>
						<button
							type="button"
							className="btn btn-ghost btn-circle btn-xs shrink-0 p-0 opacity-80 hover:opacity-100"
							onClick={onNewTab}
							aria-label="New tab"
							title="New tab"
						>
							<FiPlus size={18} />
						</button>
					</div>
				</div>
				<div className="tooltip tooltip-left mx-2 p-0" data-tip="Export Conversation As JSON">
					<DownloadButton
						language="json"
						valueFetcher={getConversationForExport}
						size={18}
						fileprefix="conversation"
						className="btn btn-ghost btn-circle btn-xs shrink-0 p-0 opacity-80 hover:opacity-100"
						aria-label="Export Chat"
						title="Export Chat"
					/>
				</div>
			</div>
		</div>
	);
});
