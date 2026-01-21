/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { useCallback, useEffect, useMemo, useState } from 'react';

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

export function ChatTabsBar({
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

		elements.push(
			<Tab
				key={t.tabId}
				store={store}
				id={t.tabId}
				// render as <div> so we can safely place an <input> inside the tab (no <input> inside <button>)
				render={<div />}
				className={[
					'relative flex h-full w-44 items-center gap-1 overflow-hidden p-0',
					'select-none',
					'focus-visible:outline-primary focus-visible:outline focus-visible:outline-offset-2',
					// Firefox-ish feel: rounded top + active lifted
					isActive
						? 'bg-base-100 text-base-content border-base-300 rounded-t-xl border shadow-sm'
						: 'bg-base-200/80 text-base-content/80 hover:bg-base-200 border-0',
				].join(' ')}
			>
				{/* Busy indicator */}
				<div className="w-4 shrink-0">
					{t.isBusy ? <span className="loading loading-spinner loading-xs" aria-label="Busy" /> : null}
				</div>

				{/* Title / Rename */}
				<div className="min-w-0 flex-1 overflow-hidden p-0 text-sm">
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
						<div
							className={[
								'flex min-w-0 items-center gap-2 overflow-hidden p-0',
								isActive && t.renameEnabled ? 'cursor-text' : 'cursor-pointer',
							].join(' ')}
							title={t.renameEnabled ? t.title + ' (Click to rename)' : t.title}
							onClick={() => {
								// Only clicking the ACTIVE tab title opens rename
								if (!isActive) return;
								if (!t.renameEnabled) return;
								setEditingTabId(t.tabId);
								setDraftTitle(t.title);
							}}
						>
							<span className="truncate">{t.title}</span>

							{/* Pencil is indicator only (no separate click action) */}
							{isActive && t.renameEnabled ? (
								<span className="shrink-0 opacity-70">
									<FiEdit2 size={12} />
								</span>
							) : null}
						</div>
					)}
				</div>

				{/* Close button (available on ALL tabs) */}
				<div className="py-0 pr-1 pl-0">
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

	// New tab button immediately beside the ACTIVE tab.

	const disableNewTab = active?.isEmpty; // no multiple empty new tabs
	elements.push();

	return (
		<div className="w-full min-w-0 overflow-hidden">
			{/* Tab strip with bottom border line (no scrollbars) */}
			<div className="border-base-300 flex w-full items-center gap-2 overflow-hidden border-b bg-inherit">
				<div className="flex flex-1">
					<TabList
						store={store}
						aria-label="Chat tabs"
						className="flex min-w-0 flex-nowrap items-end gap-0 overflow-hidden py-0 pr-1"
					>
						{elements}
					</TabList>

					<div
						className="tooltip tooltip-left px-1 py-0"
						data-tip={
							disableNewTab
								? 'Already on a new/empty tab'
								: tabs.length >= maxTabs
									? 'New tab (LRU tab will close)'
									: 'New tab'
						}
					>
						<button
							type="button"
							className="btn btn-ghost btn-circle btn-xs shrink-0 p-0 opacity-80 hover:opacity-100"
							disabled={disableNewTab}
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
}
