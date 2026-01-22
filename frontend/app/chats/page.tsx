/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { useStoreState } from '@ariakit/react';
import { useTabStore } from '@ariakit/react/tab';

import type {
	Conversation,
	ConversationSearchItem,
	StoreConversation,
	StoreConversationMessage,
} from '@/spec/conversation';
import { RoleEnum } from '@/spec/inference';

import { defaultShortcutConfig, type ShortcutConfig, useChatShortcuts } from '@/lib/keyboard_shortcuts';
import { sanitizeConversationTitle } from '@/lib/text_utils';
import { generateTitle } from '@/lib/title_utils';

import { useTitleBarContent } from '@/hooks/use_title_bar';

import { conversationStoreAPI } from '@/apis/baseapi';

import { PageFrame } from '@/components/page_frame';

import { ChatSearch, type ChatSearchHandle } from '@/chats/chat_search';
import { ChatTabsBar } from '@/chats/chat_tabs_bar';
import {
	buildInitialChatsModel,
	type ChatTabState,
	createEmptyTab,
	type InitialChatsModel,
	isScratchTab,
	MAX_TABS,
	writePersistedChatsPageState,
} from '@/chats/chat_tabs_persist';
import { ConversationArea, type ConversationAreaHandle } from '@/chats/conversation/conversation_area';
import { hydrateConversation, initConversation } from '@/chats/conversation/hydration_helper';

// eslint-disable-next-line no-restricted-exports
export default function ChatsPage() {
	// Compute initial model once (important: stable IDs; no double-randomUUID)
	const initialModelRef = useRef<InitialChatsModel | null>(null);
	if (!initialModelRef.current) initialModelRef.current = buildInitialChatsModel();
	const initialModel = initialModelRef.current;

	// ---------------- Tabs state ----------------
	const lastActivatedAtRef = useRef(new Map<string, number>());
	const touchTab = useCallback((tabId: string) => {
		lastActivatedAtRef.current.set(tabId, Date.now());
	}, []);
	const [tabs, setTabs] = useState<ChatTabState[]>(initialModel.tabs);

	const tabsRef = useRef(tabs);
	useEffect(() => {
		tabsRef.current = tabs;
	}, [tabs]);

	const tabStore = useTabStore({ defaultSelectedId: initialModel.selectedTabId });
	const selectedTabId = useStoreState(tabStore, 'selectedId') ?? initialModel.selectedTabId;

	const selectedTabIdRef = useRef(selectedTabId);
	useEffect(() => {
		selectedTabIdRef.current = selectedTabId;
	}, [selectedTabId]);

	const activeTab = useMemo(() => tabs.find(t => t.tabId === selectedTabId) ?? tabs[0], [tabs, selectedTabId]);

	// ---------------- Conversation area (conversation runtime + UI) ----------------
	const conversationAreaRef = useRef<ConversationAreaHandle | null>(null);

	// Seed runtime maps from persisted state (once) - LRU only (scroll is now in ConversationArea)
	const seededRuntimeFromStorageRef = useRef(false);
	if (!seededRuntimeFromStorageRef.current) {
		seededRuntimeFromStorageRef.current = true;
		for (const [id, ts] of Object.entries(initialModel.lastActivatedAtByTab)) {
			if (typeof ts === 'number') lastActivatedAtRef.current.set(id, ts);
		}
	}

	const disposeTabRuntime = useCallback((tabId: string) => {
		conversationAreaRef.current?.disposeTabRuntime(tabId);
		lastActivatedAtRef.current.delete(tabId);
	}, []);

	// ---------------- UI refs ----------------
	const searchRef = useRef<ChatSearchHandle | null>(null);

	// ---------------- Helpers ----------------
	const updateTab = useCallback((tabId: string, updater: (t: ChatTabState) => ChatTabState) => {
		setTabs(prev => {
			const idx = prev.findIndex(t => t.tabId === tabId);
			if (idx === -1) return prev; // closed while async
			const next = prev.slice();
			next[idx] = updater(next[idx]);
			return next;
		});
	}, []);

	useEffect(() => {
		touchTab(selectedTabId);
	}, [selectedTabId, touchTab]);

	// If the selected tab id becomes invalid (e.g. restored state had stale ids), correct it.
	useEffect(() => {
		if (tabs.length === 0) return;
		if (tabs.some(t => t.tabId === selectedTabId)) return;
		tabStore.setSelectedId(tabs[0].tabId);
	}, [selectedTabId, tabStore, tabs]);

	// ---------------- Search refresh key ----------------
	const [searchRefreshKey, setSearchRefreshKey] = useState(0);
	const bumpSearchKey = useCallback(async () => {
		await new Promise(resolve => setTimeout(resolve, 50));
		setSearchRefreshKey(k => k + 1);
	}, []);

	// ---------------- Persistence ----------------
	const saveUpdatedConversation = useCallback(
		(tabId: string, updatedConv: Conversation, titleWasExternallyChanged = false) => {
			const tab = tabsRef.current.find(t => t.tabId === tabId);
			if (!tab) return;

			let newTitle = updatedConv.title;

			const allowAutoTitle = !titleWasExternallyChanged && !tab.manualTitleLocked;
			if (allowAutoTitle && updatedConv.messages.length <= 4) {
				const userMessages = updatedConv.messages.filter(m => m.role === RoleEnum.User);
				if (userMessages.length === 1) {
					newTitle = generateTitle(userMessages[0].uiContent).title;
				} else if (userMessages.length === 2) {
					const c1 = generateTitle(userMessages[0].uiContent);
					const c2 = generateTitle(userMessages[1].uiContent);
					newTitle = c2.score > c1.score ? c2.title : c1.title;
				}
				newTitle = sanitizeConversationTitle(newTitle);
			}

			const titleChangedByFunction = newTitle !== updatedConv.title;
			if (titleChangedByFunction) updatedConv.title = newTitle;

			const titleChanged = titleWasExternallyChanged || titleChangedByFunction;

			if (!tab.isPersisted) {
				conversationStoreAPI.putConversation(updatedConv as StoreConversation);
				void bumpSearchKey();
			} else if (titleChanged) {
				conversationStoreAPI.putConversation(updatedConv as StoreConversation);
				void bumpSearchKey();
			} else {
				conversationStoreAPI.putMessagesToConversation(
					updatedConv.id,
					updatedConv.title,
					updatedConv.messages as StoreConversationMessage[]
				);
			}

			updateTab(tabId, t => ({
				...t,
				conversation: { ...updatedConv, messages: [...updatedConv.messages] },
				isPersisted: true,
				manualTitleLocked: titleWasExternallyChanged ? true : t.manualTitleLocked,
			}));
		},
		[bumpSearchKey, updateTab]
	);

	// ---------------- LRU eviction ----------------
	const pickLRUEvictionCandidate = useCallback((current: ChatTabState[], activeId: string) => {
		if (current.length < MAX_TABS) return null;
		// Never evict scratch; prefer not to evict busy, but allow if needed.
		const base = current.filter(t => t.tabId !== activeId && !isScratchTab(t));
		const nonBusy = base.filter(t => !t.isBusy);
		const candidates = nonBusy.length > 0 ? nonBusy : base;

		if (candidates.length === 0) return null;
		const ts = (id: string) => lastActivatedAtRef.current.get(id) ?? 0;
		return candidates.reduce((lru, t) => (ts(t.tabId) < ts(lru.tabId) ? t : lru), candidates[0]);
	}, []);

	const ensureScratchRightmost = useCallback(() => {
		const current = tabsRef.current;
		const scratchTabs = current.filter(isScratchTab);

		// Pick which scratch to keep (prefer selected if it's scratch, else keep the rightmost)
		const selected = current.find(t => t.tabId === selectedTabIdRef.current);
		const keep = (selected && isScratchTab(selected) ? selected : scratchTabs[scratchTabs.length - 1]) ?? null;

		const removeIds = new Set(scratchTabs.filter(t => !keep || t.tabId !== keep.tabId).map(t => t.tabId));
		if (removeIds.size > 0) {
			for (const id of removeIds) disposeTabRuntime(id);
		}

		let next = current.filter(t => !removeIds.has(t.tabId));

		if (!keep) {
			// Need to create a new scratch tab
			if (next.length >= MAX_TABS) {
				const victim = pickLRUEvictionCandidate(next, selectedTabIdRef.current);
				if (victim) {
					disposeTabRuntime(victim.tabId);
					next = next.filter(t => t.tabId !== victim.tabId);
				}
			}
			const scratch = createEmptyTab();
			touchTab(scratch.tabId);

			conversationAreaRef.current?.setScrollTopForTab(scratch.tabId, 0);
			next = [...next, scratch];
		} else {
			// Ensure it's rightmost
			const idx = next.findIndex(t => t.tabId === keep.tabId);
			if (idx !== -1 && idx !== next.length - 1) {
				next = [...next.slice(0, idx), ...next.slice(idx + 1), keep];
			}
		}

		// Commit only if changed (cheap shallow check)
		if (next.length !== current.length || next.some((t, i) => t.tabId !== current[i]?.tabId)) {
			setTabs(next);
		}
	}, [disposeTabRuntime, pickLRUEvictionCandidate, touchTab]);

	const scratchKey = useMemo(() => {
		// changes when scratch tabs count/identity changes OR when tab count changes
		const scratchIds = tabs
			.filter(isScratchTab)
			.map(t => t.tabId)
			.join('|');
		return `${tabs.length}:${scratchIds}`;
	}, [tabs]);

	// Enforce scratch invariants after any tabs mutation (search load, send, close, etc.)
	useEffect(() => {
		ensureScratchRightmost();
	}, [scratchKey, ensureScratchRightmost]);

	// ---------------- Persist open tabs + selected + scroll/LRU ----------------
	const scrollTopSnapshotRef = useRef<Record<string, number>>(initialModel.scrollTopByTab ?? {});

	const persistNow = useCallback(() => {
		const tabsSnapshot = tabsRef.current.slice(0, MAX_TABS);

		const scrollObj =
			conversationAreaRef.current?.getScrollTopByTabSnapshot() ??
			scrollTopSnapshotRef.current ??
			({} as Record<string, number>);
		scrollTopSnapshotRef.current = scrollObj;

		const lruObj: Record<string, number> = {};
		for (const [k, v] of lastActivatedAtRef.current.entries()) lruObj[k] = v;

		writePersistedChatsPageState({
			v: 1,
			selectedTabId: selectedTabIdRef.current,
			tabs: tabsSnapshot.map(t => ({
				tabId: t.tabId,
				conversationId: t.conversation.id,
				title: t.conversation.title,
				isPersisted: t.isPersisted,
				manualTitleLocked: t.manualTitleLocked,
			})),
			scrollTopByTab: scrollObj,
			lastActivatedAtByTab: lruObj,
		});
	}, []);

	const persistTimerRef = useRef<number | null>(null);
	const schedulePersist = useCallback(() => {
		if (persistTimerRef.current) window.clearTimeout(persistTimerRef.current);
		persistTimerRef.current = window.setTimeout(() => {
			persistTimerRef.current = null;
			persistNow();
		}, 250);
	}, [persistNow]);

	useEffect(() => {
		schedulePersist();
	}, [schedulePersist, tabs, selectedTabId]);

	useEffect(() => {
		const onPageHide = () => {
			persistNow();
		};
		const onVis = () => {
			if (document.visibilityState === 'hidden') persistNow();
		};
		window.addEventListener('pagehide', onPageHide);
		document.addEventListener('visibilitychange', onVis);
		return () => {
			window.removeEventListener('pagehide', onPageHide);
			document.removeEventListener('visibilitychange', onVis);
		};
	}, [persistNow]);

	// On unmount: flush persistence (route navigation safety)
	useEffect(() => {
		return () => {
			persistNow();
		};
	}, [persistNow]);

	// ---------------- Tab actions ----------------
	const openNewTab = useCallback(() => {
		// Always go to the (single) scratch tab; invariant effect keeps it rightmost.
		const scratch = tabsRef.current.find(isScratchTab) ?? null;
		const targetId = scratch?.tabId ?? selectedTabIdRef.current;
		tabStore.setSelectedId(targetId);
		requestAnimationFrame(() => conversationAreaRef.current?.focusInput(targetId));
	}, [tabStore]);

	const closeTab = useCallback(
		(tabId: string) => {
			const current = tabsRef.current;
			const wasActive = tabId === selectedTabIdRef.current;

			// 1) Stop any async work / cleanup runtime refs immediately
			disposeTabRuntime(tabId);

			// 2) Compute next tabs synchronously from a single snapshot (no mutation inside setState)
			const idx = current.findIndex(t => t.tabId === tabId);
			let nextTabs = current.filter(t => t.tabId !== tabId);

			// 3) Ensure we never end up with 0 tabs
			if (nextTabs.length === 0) {
				const fresh = createEmptyTab();
				conversationAreaRef.current?.setScrollTopForTab(fresh.tabId, 0);
				nextTabs = [fresh];
			}

			// 4) If the closed tab was active, select nearest neighbor (prefer right, else left)
			let nextSelectedId = selectedTabIdRef.current;
			if (wasActive) {
				const right = idx >= 0 ? current[idx + 1] : undefined;
				const left = idx > 0 ? current[idx - 1] : undefined;
				nextSelectedId =
					(right && right.tabId !== tabId ? right.tabId : left && left.tabId !== tabId ? left.tabId : '') ||
					nextTabs[0].tabId;
			}

			// 5) Commit state + selection (selection set after tabs update to avoid invalid ids)
			setTabs(nextTabs);
			if (wasActive) {
				requestAnimationFrame(() => {
					// still valid? (paranoid guard)
					const ok =
						tabsRef.current.some(t => t.tabId === nextSelectedId) || nextTabs.some(t => t.tabId === nextSelectedId);
					tabStore.setSelectedId(ok ? nextSelectedId : nextTabs[0].tabId);
				});
			}
		},
		[disposeTabRuntime, tabStore]
	);

	const cycleTabBy = useCallback(
		(delta: number) => {
			const current = tabsRef.current;
			if (current.length < 2) return;

			const activeId = selectedTabIdRef.current;
			const idx = current.findIndex(t => t.tabId === activeId);
			const from = idx >= 0 ? idx : 0;
			const nextIndex = (from + delta + current.length) % current.length;
			const nextId = current[nextIndex]?.tabId;
			if (!nextId) return;

			tabStore.setSelectedId(nextId);
			requestAnimationFrame(() => conversationAreaRef.current?.focusInput(nextId));
		},
		[tabStore]
	);

	const selectNextTab = useCallback(() => {
		cycleTabBy(1);
	}, [cycleTabBy]);
	const selectPrevTab = useCallback(() => {
		cycleTabBy(-1);
	}, [cycleTabBy]);

	// ---------------- Rename ----------------
	const renameTabTitle = useCallback(
		(tabId: string, newTitle: string) => {
			const tab = tabsRef.current.find(t => t.tabId === tabId);
			if (!tab) return;

			const sanitized = sanitizeConversationTitle(newTitle.trim());
			if (!sanitized || sanitized === tab.conversation.title) return;

			const updatedConv: Conversation = {
				...tab.conversation,
				title: sanitized,
				modifiedAt: new Date(),
			};

			saveUpdatedConversation(tabId, updatedConv, true);
		},
		[saveUpdatedConversation]
	);

	// ---------------- Search select behavior ----------------
	const loadConversationIntoTab = useCallback(
		async (tabId: string, item: ConversationSearchItem) => {
			const selectedChat = await conversationStoreAPI.getConversation(item.id, item.title, true);
			if (!selectedChat) return;

			const hydrated = hydrateConversation(selectedChat);

			// Reset stream buffer for that tab
			conversationAreaRef.current?.clearStreamForTab(tabId);

			updateTab(tabId, t => ({
				...t,
				conversation: hydrated,
				isPersisted: true,
				manualTitleLocked: false,
				editingMessageId: null,
				isBusy: false,
			}));

			// Reset scroll position for that tab
			conversationAreaRef.current?.resetScrollToTop(tabId);

			requestAnimationFrame(() => {
				conversationAreaRef.current?.syncComposerFromConversation(tabId, hydrated);
			});
		},
		[updateTab]
	);

	const handleSelectConversation = useCallback(
		async (item: ConversationSearchItem) => {
			// If already open, just activate it
			const already = tabsRef.current.find(t => t.conversation.id === item.id);
			if (already) {
				tabStore.setSelectedId(already.tabId);
				return;
			}
			// Always load into scratch tab; after load it becomes a normal tab and
			// the invariant effect will create a new scratch on the right.
			const scratch = tabsRef.current.find(isScratchTab);
			const targetId = scratch?.tabId ?? selectedTabIdRef.current;
			tabStore.setSelectedId(targetId);
			await loadConversationIntoTab(targetId, item);
			requestAnimationFrame(() => conversationAreaRef.current?.focusInput(targetId));
		},
		[loadConversationIntoTab, tabStore]
	);

	// ---------------- Rehydrate tabs on mount (from Go-backed store) ----------------
	useEffect(() => {
		if (!initialModelRef.current?.restoredFromStorage) return;

		let cancelled = false;

		(async () => {
			const snapshot = tabsRef.current.slice(0, MAX_TABS);
			for (const t of snapshot) {
				if (cancelled) return;
				if (!t.isPersisted) continue;
				if (!t.conversation.id) continue;

				try {
					const stored = await conversationStoreAPI.getConversation(t.conversation.id, t.conversation.title, true);
					if (cancelled) return;

					if (!stored) {
						// Conversation missing -> degrade to scratch; scratch invariant will clean up.
						updateTab(t.tabId, prev => ({
							...prev,
							isBusy: false,
							isPersisted: false,
							manualTitleLocked: false,
							editingMessageId: null,
							conversation: initConversation(),
						}));
						continue;
					}

					const hydrated = hydrateConversation(stored);

					// Clear transient streaming buffer for restored tabs
					conversationAreaRef.current?.clearStreamForTab(t.tabId);

					updateTab(t.tabId, prev => ({
						...prev,
						isBusy: false,
						editingMessageId: null,
						isPersisted: true,
						conversation: hydrated,
					}));

					requestAnimationFrame(() => {
						conversationAreaRef.current?.syncComposerFromConversation(t.tabId, hydrated);
					});
				} catch (e) {
					console.error(e);
				}
			}
		})().catch(console.error);

		return () => {
			cancelled = true;
		};
	}, [updateTab]);

	// ---------------- Shortcuts ----------------
	const [shortcutConfig] = useState<ShortcutConfig>(defaultShortcutConfig);

	useChatShortcuts({
		config: shortcutConfig,
		isBusy: false, // new tabs allowed even if another tab is busy
		handlers: {
			newChat: () => {
				openNewTab();
			},
			closeChat: () => {
				closeTab(selectedTabIdRef.current);
			},
			nextChat: () => {
				selectNextTab();
			},
			previousChat: () => {
				selectPrevTab();
			},
			focusSearch: () => searchRef.current?.focusInput(),
			focusInput: () => conversationAreaRef.current?.focusInput(selectedTabIdRef.current),
			insertTemplate: () => conversationAreaRef.current?.openTemplateMenu(selectedTabIdRef.current),
			insertTool: () => conversationAreaRef.current?.openToolMenu(selectedTabIdRef.current),
			insertAttachment: () => conversationAreaRef.current?.openAttachmentMenu(selectedTabIdRef.current),
		},
	});

	// ---------------- Titlebar: search only (like before) ----------------
	useTitleBarContent(
		{
			center: (
				<div className="mx-auto flex w-4/5 items-center justify-center">
					<ChatSearch
						ref={searchRef}
						compact={true}
						onSelectConversation={handleSelectConversation}
						refreshKey={searchRefreshKey}
						currentConversationId={activeTab?.conversation.id ?? ''}
					/>
				</div>
			),
		},
		[activeTab?.conversation.id, handleSelectConversation, searchRefreshKey]
	);

	// ---------------- Export (active tab) ----------------
	const getConversationForExport = useCallback(async (): Promise<string> => {
		const t = tabsRef.current.find(x => x.tabId === selectedTabIdRef.current);
		if (!t) return JSON.stringify(null, null, 2);

		const selectedChat = await conversationStoreAPI.getConversation(t.conversation.id, t.conversation.title, true);
		return JSON.stringify(selectedChat ?? null, null, 2);
	}, []);

	// ---------------- Render helpers ----------------
	const tabBarItems = useMemo(
		() =>
			tabs.map(t => ({
				tabId: t.tabId,
				title: t.conversation.title,
				isBusy: t.isBusy,
				isEmpty: t.conversation.messages.length === 0,
				renameEnabled: t.conversation.messages.length > 0,
			})),
		[tabs]
	);

	return (
		<PageFrame contentScrollable={false}>
			<div className="grid h-full w-full grid-rows-[auto_1fr_auto] overflow-hidden">
				{/* Row 1: TAB STRIP (where navbar used to be) + Download floater */}
				<div className="relative row-start-1 row-end-2 min-h-0 min-w-0 p-0">
					<ChatTabsBar
						store={tabStore}
						selectedTabId={selectedTabId}
						tabs={tabBarItems}
						maxTabs={MAX_TABS}
						onNewTab={openNewTab}
						onCloseTab={closeTab}
						onRenameTab={renameTabTitle}
						getConversationForExport={getConversationForExport}
					/>
				</div>

				{/* Rows 2 + 3: Conversation UI/runtime (messages + per-tab input panes) */}
				<ConversationArea
					ref={conversationAreaRef}
					tabs={tabs}
					selectedTabId={selectedTabId}
					shortcutConfig={shortcutConfig}
					initialScrollTopByTab={initialModel.scrollTopByTab}
					updateTab={updateTab}
					saveUpdatedConversation={saveUpdatedConversation}
				/>
			</div>
		</PageFrame>
	);
}
