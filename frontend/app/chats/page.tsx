/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react';

import { useStoreState } from '@ariakit/react';
import { useTabStore } from '@ariakit/react/tab';

import type {
	Conversation,
	ConversationMessage,
	ConversationSearchItem,
	StoreConversation,
	StoreConversationMessage,
} from '@/spec/conversation';
import { ContentItemKind, type ModelParam, OutputKind, type OutputUnion, RoleEnum, Status } from '@/spec/inference';
import { DefaultUIChatOptions, type UIChatOption } from '@/spec/modelpreset';
import { type ToolStoreChoice, ToolStoreChoiceType } from '@/spec/tool';

import { defaultShortcutConfig, type ShortcutConfig, useChatShortcuts } from '@/lib/keyboard_shortcuts';
import { getBlockQuotedLines, sanitizeConversationTitle } from '@/lib/text_utils';
import { generateTitle } from '@/lib/title_utils';

import { useAtTopBottom } from '@/hooks/use_at_top_bottom';
import { useTitleBarContent } from '@/hooks/use_title_bar';

import { conversationStoreAPI } from '@/apis/baseapi';

import { ButtonScrollToBottom, ButtonScrollToTop } from '@/components/button_scroll_top_bottom';
import { PageFrame } from '@/components/page_frame';

import { ChatSearch, type ChatSearchHandle } from '@/chats/chat_search';
import { ChatTabsBar } from '@/chats/chat_tabs_bar';
import { HandleCompletion } from '@/chats/conversation/completion_helper';
import {
	buildUserConversationMessageFromEditor,
	deriveConversationToolsFromMessages,
	deriveWebSearchChoiceFromMessages,
	hydrateConversation,
	initConversation,
	initConversationMessage,
} from '@/chats/conversation/hydration_helper';
import { InputBox, type InputBoxHandle } from '@/chats/inputarea/input_box';
import type { EditorExternalMessage, EditorSubmitPayload } from '@/chats/inputarea/input_editor_utils';
import { ChatMessage } from '@/chats/messages/message';

const MAX_TABS = 8;

type ChatTabState = {
	tabId: string;
	conversation: Conversation;

	// streaming (state only for coarse UI like spinner)
	isBusy: boolean;

	// persistence/title behavior
	isPersisted: boolean;
	manualTitleLocked: boolean;

	// edit state
	editingMessageId: string | null;

	// LRU
	lastActivatedAt: number;
};

function createEmptyTab(): ChatTabState {
	return {
		tabId: crypto.randomUUID(),
		conversation: initConversation(),
		isBusy: false,
		isPersisted: false,
		manualTitleLocked: false,
		editingMessageId: null,
		lastActivatedAt: Date.now(),
	};
}

function isScratchTab(t: ChatTabState) {
	// "New conversation" = local scratch, not a stored convo
	return !t.isPersisted && t.conversation.messages.length === 0;
}

// eslint-disable-next-line no-restricted-exports
export default function ChatsPage() {
	// ---------------- Tabs state ----------------
	const initialTab = useMemo(() => createEmptyTab(), []);
	const [tabs, setTabs] = useState<ChatTabState[]>([initialTab]);
	const tabsRef = useRef(tabs);
	useEffect(() => {
		tabsRef.current = tabs;
	}, [tabs]);

	const tabStore = useTabStore({ defaultSelectedId: initialTab.tabId });
	const selectedTabId = useStoreState(tabStore, 'selectedId') ?? initialTab.tabId;

	const selectedTabIdRef = useRef(selectedTabId);
	useEffect(() => {
		selectedTabIdRef.current = selectedTabId;
	}, [selectedTabId]);

	const activeTab = useMemo(() => tabs.find(t => t.tabId === selectedTabId) ?? tabs[0], [tabs, selectedTabId]);

	// ---------------- Per-tab runtime refs ----------------
	// Abort controllers per tab
	const abortRefs = useRef(new Map<string, { current: AbortController | null }>());
	const requestIdByTab = useRef(new Map<string, string | null>());
	const tokensReceivedByTab = useRef(new Map<string, boolean | null>());

	const getAbortRef = useCallback((tabId: string) => {
		let refObj = abortRefs.current.get(tabId);
		if (!refObj) {
			refObj = { current: null };
			abortRefs.current.set(tabId, refObj);
		}
		return refObj;
	}, []);

	// Stream text stored in refs (NO state updates for background tabs)
	const streamTextRefs = useRef(new Map<string, { current: string }>());
	const getStreamTextRef = useCallback((tabId: string) => {
		let refObj = streamTextRefs.current.get(tabId);
		if (!refObj) {
			refObj = { current: '' };
			streamTextRefs.current.set(tabId, refObj);
		}
		return refObj;
	}, []);

	// Input refs per tab (per-tab composer instance)
	const inputRefs = useRef(new Map<string, InputBoxHandle | null>());
	const setInputRef = useCallback((tabId: string) => {
		return (inst: InputBoxHandle | null) => {
			inputRefs.current.set(tabId, inst);
		};
	}, []);

	// Scroll position restore per tab
	const scrollTopByTab = useRef(new Map<string, number>());

	const disposeTabRuntime = useCallback(
		(tabId: string) => {
			const a = getAbortRef(tabId);
			a.current?.abort();
			a.current = null;

			abortRefs.current.delete(tabId);
			requestIdByTab.current.delete(tabId);
			tokensReceivedByTab.current.delete(tabId);

			streamTextRefs.current.delete(tabId);
			inputRefs.current.delete(tabId);
			scrollTopByTab.current.delete(tabId);
		},
		[getAbortRef]
	);

	// ---------------- UI refs ----------------
	const chatContainerRef = useRef<HTMLDivElement>(null);
	const searchRef = useRef<ChatSearchHandle | null>(null);

	const { isAtBottom, isAtTop, isScrollable } = useAtTopBottom(chatContainerRef);

	// Active stream text state (ONLY for active tab, throttled)
	const [activeStreamText, setActiveStreamText] = useState<string>('');
	const pendingActiveStreamSync = useRef<number | null>(null);

	const syncActiveStreamTextNow = useCallback(() => {
		pendingActiveStreamSync.current = null;
		const tabId = selectedTabIdRef.current;
		const refObj = streamTextRefs.current.get(tabId);
		setActiveStreamText(refObj?.current ?? '');
	}, []);

	const scheduleActiveStreamSync = useCallback(
		(tabId: string) => {
			if (selectedTabIdRef.current !== tabId) return;
			if (pendingActiveStreamSync.current !== null) return;
			pendingActiveStreamSync.current = window.setTimeout(syncActiveStreamTextNow, 80);
		},
		[syncActiveStreamTextNow]
	);

	// When user switches tabs, show whatever stream text exists for that tab.
	useEffect(() => {
		setActiveStreamText(getStreamTextRef(selectedTabId).current);
	}, [getStreamTextRef, selectedTabId]);

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

	// Update LRU timestamp on activation (update only the selected tab)
	useEffect(() => {
		updateTab(selectedTabId, t => ({ ...t, lastActivatedAt: Date.now() }));
	}, [selectedTabId, updateTab]);

	const tabExists = useCallback((tabId: string) => tabsRef.current.some(t => t.tabId === tabId), []);

	const syncComposerFromConversation = useCallback((tabId: string, conv: Conversation) => {
		const input = inputRefs.current.get(tabId);
		if (!input) return;

		const tools = deriveConversationToolsFromMessages(conv.messages);
		const web = deriveWebSearchChoiceFromMessages(conv.messages);
		input.setConversationToolsFromChoices(tools);
		input.setWebSearchFromChoices(web);
	}, []);

	const scrollToBottom = useCallback((tabId: string) => {
		// There is only one visible scroll container; it corresponds to the active tab.
		if (selectedTabIdRef.current !== tabId) return;
		const el = chatContainerRef.current;
		if (!el) return;
		el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' });
	}, []);

	const scrollToBottomSoon = useCallback(
		(tabId: string) => {
			window.setTimeout(() => {
				scrollToBottom(tabId);
			}, 80);
		},
		[scrollToBottom]
	);

	// Save scroll position for active tab on scroll
	const onScrollActive = useCallback(() => {
		const tabId = selectedTabIdRef.current;
		const el = chatContainerRef.current;
		if (!el) return;
		scrollTopByTab.current.set(tabId, el.scrollTop);
	}, []);

	// Restore scroll position when changing tabs (LRU + scroll restore requirement)
	useLayoutEffect(() => {
		const el = chatContainerRef.current;
		if (!el) return;

		const top = scrollTopByTab.current.get(selectedTabId) ?? 0;

		// Use rAF to ensure DOM layout is committed for the new tab content
		requestAnimationFrame(() => {
			const el2 = chatContainerRef.current;
			if (!el2) return;
			el2.scrollTop = top;
		});
	}, [selectedTabId]);

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
		return candidates.reduce((lru, t) => (t.lastActivatedAt < lru.lastActivatedAt ? t : lru), candidates[0]);
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
			getAbortRef(scratch.tabId);
			getStreamTextRef(scratch.tabId);
			scrollTopByTab.current.set(scratch.tabId, 0);
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
	}, [disposeTabRuntime, getAbortRef, getStreamTextRef, pickLRUEvictionCandidate]);

	// Enforce scratch invariants after any tabs mutation (search load, send, close, etc.)
	useEffect(() => {
		ensureScratchRightmost();
	}, [tabs, ensureScratchRightmost]);

	// ---------------- Tab actions ----------------
	const openNewTab = useCallback(() => {
		// Always go to the (single) scratch tab; invariant effect keeps it rightmost.
		const scratch = tabsRef.current.find(isScratchTab) ?? null;
		const targetId = scratch?.tabId ?? selectedTabIdRef.current;
		tabStore.setSelectedId(targetId);
		requestAnimationFrame(() => inputRefs.current.get(targetId)?.focus());
	}, [tabStore]);

	const closeTab = useCallback(
		(tabId: string) => {
			const current = tabsRef.current;
			const isActive = tabId === selectedTabIdRef.current;

			disposeTabRuntime(tabId);

			let nextSelected = selectedTabIdRef.current;

			if (isActive) {
				const idx = current.findIndex(t => t.tabId === tabId);
				const right = idx >= 0 ? current[idx + 1] : undefined;
				const left = idx > 0 ? current[idx - 1] : undefined;
				nextSelected = (right ?? left)?.tabId ?? '';
			}

			setTabs(prev => {
				const filtered = prev.filter(t => t.tabId !== tabId);
				if (filtered.length === 0) {
					const fresh = createEmptyTab();
					getAbortRef(fresh.tabId);
					getStreamTextRef(fresh.tabId);
					nextSelected = fresh.tabId;
					scrollTopByTab.current.set(fresh.tabId, 0);
					return [fresh];
				}
				return filtered;
			});

			if (isActive) {
				const fallback = tabsRef.current.filter(t => t.tabId !== tabId)[0]?.tabId ?? initialTab.tabId;
				tabStore.setSelectedId(nextSelected || fallback);
			}
		},
		[disposeTabRuntime, getAbortRef, getStreamTextRef, initialTab.tabId, tabStore]
	);

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
			getStreamTextRef(tabId).current = '';
			if (selectedTabIdRef.current === tabId) setActiveStreamText('');

			updateTab(tabId, t => ({
				...t,
				conversation: hydrated,
				isPersisted: true,
				manualTitleLocked: false,
				editingMessageId: null,
				isBusy: false,
			}));

			// Reset scroll position for that tab
			scrollTopByTab.current.set(tabId, 0);
			if (selectedTabIdRef.current === tabId) {
				requestAnimationFrame(() => {
					const el = chatContainerRef.current;
					if (el) el.scrollTop = 0;
				});
			}

			requestAnimationFrame(() => {
				syncComposerFromConversation(tabId, hydrated);
			});
		},
		[getStreamTextRef, syncComposerFromConversation, updateTab]
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
			requestAnimationFrame(() => inputRefs.current.get(targetId)?.focus());
		},
		[loadConversationIntoTab, tabStore]
	);

	// ---------------- Streaming completion (per tab) ----------------
	const updateStreamingMessage = useCallback(
		async (tabId: string, updatedChatWithUserMessage: Conversation, options: UIChatOption) => {
			if (!tabExists(tabId)) return;

			const abortRef = getAbortRef(tabId);

			abortRef.current?.abort();
			tokensReceivedByTab.current.set(tabId, false);

			// mark busy (coarse UI only)
			updateTab(tabId, t => ({ ...t, isBusy: true }));

			abortRef.current = new AbortController();
			requestIdByTab.current.set(tabId, crypto.randomUUID());

			let allMessages = updatedChatWithUserMessage.messages;
			if (options.disablePreviousMessages) {
				allMessages = [updatedChatWithUserMessage.messages[updatedChatWithUserMessage.messages.length - 1]];
			}
			if (allMessages.length === 0) {
				updateTab(tabId, t => ({ ...t, isBusy: false }));
				return;
			}

			const currentUserMsg = allMessages[allMessages.length - 1];
			const history = allMessages.slice(0, allMessages.length - 1);

			// reset stream buffer (ref only)
			const streamRef = getStreamTextRef(tabId);
			streamRef.current = '';
			if (selectedTabIdRef.current === tabId) setActiveStreamText('');

			// assistant placeholder for streaming
			const assistantPlaceholder = initConversationMessage(RoleEnum.Assistant);
			const chatWithPlaceholder: Conversation = {
				...updatedChatWithUserMessage,
				messages: [...updatedChatWithUserMessage.messages, assistantPlaceholder],
				modifiedAt: new Date(),
			};

			// Show placeholder immediately (single state update)
			updateTab(tabId, t => ({
				...t,
				conversation: { ...chatWithPlaceholder, messages: [...chatWithPlaceholder.messages] },
			}));

			if (selectedTabIdRef.current === tabId) scrollToBottomSoon(tabId);

			const onStreamTextData = (textData: string) => {
				if (!textData) return;
				tokensReceivedByTab.current.set(tabId, true);

				// Append to ref only
				streamRef.current += textData;

				// Only active tab triggers throttled state updates for display
				scheduleActiveStreamSync(tabId);
			};

			const onStreamThinkingData = (thinkingData: string) => {
				if (!thinkingData) return;
				tokensReceivedByTab.current.set(tabId, true);

				const block = getBlockQuotedLines(thinkingData) + '\n';
				streamRef.current += block;
				scheduleActiveStreamSync(tabId);
			};

			try {
				const inputParams: ModelParam = {
					name: options.name,
					temperature: options.temperature,
					stream: options.stream,
					maxPromptLength: options.maxPromptLength,
					maxOutputLength: options.maxOutputLength,
					reasoning: options.reasoning,
					systemPrompt: options.systemPrompt,
					timeout: options.timeout,
					additionalParametersRawJSON: options.additionalParametersRawJSON,
				};

				let toolStoreChoices: ToolStoreChoice[] | undefined;
				const latestUser = updatedChatWithUserMessage.messages
					.slice()
					.reverse()
					.find(m => m.role === RoleEnum.User);
				if (latestUser?.toolStoreChoices && latestUser.toolStoreChoices.length > 0) {
					toolStoreChoices = latestUser.toolStoreChoices;
				}

				const { responseMessage, rawResponse } = await HandleCompletion(
					options.providerName,
					inputParams,
					currentUserMsg,
					history,
					toolStoreChoices,
					assistantPlaceholder,
					requestIdByTab.current.get(tabId) ?? undefined,
					abortRef.current.signal,
					onStreamTextData,
					onStreamThinkingData
				);

				if (!tabExists(tabId)) return;

				if (responseMessage) {
					let finalChat: Conversation = {
						...chatWithPlaceholder,
						messages: [...chatWithPlaceholder.messages.slice(0, -1), responseMessage],
						modifiedAt: new Date(),
					};

					if (rawResponse?.hydratedCurrentInputs && currentUserMsg.id) {
						const hydrated = rawResponse.hydratedCurrentInputs;
						finalChat = {
							...finalChat,
							messages: finalChat.messages.map(m =>
								m.id === currentUserMsg.id
									? {
											...m,
											inputs: hydrated,
										}
									: m
							),
						};
					}

					saveUpdatedConversation(tabId, finalChat);

					// Tool calls -> load into THIS tab's composer (even if hidden)
					if (responseMessage.uiToolCalls && responseMessage.uiToolCalls.length > 0) {
						const runnableCalls = responseMessage.uiToolCalls.filter(
							c => c.type === ToolStoreChoiceType.Function || c.type === ToolStoreChoiceType.Custom
						);
						if (runnableCalls.length > 0) {
							inputRefs.current.get(tabId)?.loadToolCalls(runnableCalls);
						}
					}
				}
			} catch (e) {
				if (!tabExists(tabId)) return;

				if ((e as DOMException).name === 'AbortError') {
					const tokensReceived = tokensReceivedByTab.current.get(tabId);

					if (!tokensReceived) {
						// remove placeholder
						updateTab(tabId, t => {
							const idx = t.conversation.messages.findIndex(m => m.id === assistantPlaceholder.id);
							if (idx === -1) return t;
							const msgs = t.conversation.messages.filter((_, i) => i !== idx);
							return {
								...t,
								conversation: { ...t.conversation, messages: msgs, modifiedAt: new Date() },
							};
						});
					} else {
						const partialText = streamRef.current + '\n\n>API Aborted after partial response...';

						const partialOutputs: OutputUnion[] = [
							{
								kind: OutputKind.OutputMessage,
								outputMessage: {
									id: assistantPlaceholder.id,
									role: RoleEnum.Assistant,
									status: Status.Completed,
									contents: [{ kind: ContentItemKind.Text, textItem: { text: partialText } }],
								},
							},
						];

						const partialMsg: ConversationMessage = {
							...assistantPlaceholder,
							status: Status.Completed,
							outputs: partialOutputs,
							uiContent: partialText,
						};

						const finalChat: Conversation = {
							...chatWithPlaceholder,
							messages: [...chatWithPlaceholder.messages.slice(0, -1), partialMsg],
							modifiedAt: new Date(),
						};

						saveUpdatedConversation(tabId, finalChat);
					}
				} else {
					console.error(e);
				}
			} finally {
				if (tabExists(tabId)) {
					// clear stream buffer
					streamRef.current = '';
					if (selectedTabIdRef.current === tabId) {
						setActiveStreamText('');
					}

					updateTab(tabId, t => ({ ...t, isBusy: false }));

					if (selectedTabIdRef.current === tabId) scrollToBottomSoon(tabId);
				}
			}
		},
		[
			getAbortRef,
			getStreamTextRef,
			scheduleActiveStreamSync,
			saveUpdatedConversation,
			scrollToBottomSoon,
			tabExists,
			updateTab,
		]
	);

	// ---------------- Per-tab send/edit/resend ----------------
	const sendMessageForTab = useCallback(
		async (tabId: string, payload: EditorSubmitPayload, options: UIChatOption) => {
			const tab = tabsRef.current.find(t => t.tabId === tabId);
			if (!tab) return;
			if (tab.isBusy) return;

			const trimmed = payload.text.trim();
			const hasNonEmptyText = trimmed.length > 0;
			const hasToolOutputs = payload.toolOutputs.length > 0;
			const hasAttachments = payload.attachments.length > 0;
			if (!hasNonEmptyText && !hasToolOutputs && !hasAttachments) return;

			const editingId = tab.editingMessageId ?? undefined;
			const userMsg = buildUserConversationMessageFromEditor(payload, editingId);

			if (tab.editingMessageId) {
				const idx = tab.conversation.messages.findIndex(m => m.id === tab.editingMessageId);
				if (idx !== -1) {
					const msgs = tab.conversation.messages.slice(0, idx + 1);
					msgs[idx] = userMsg;

					const updatedChat: Conversation = {
						...tab.conversation,
						messages: msgs,
						modifiedAt: new Date(),
					};

					updateTab(tabId, t => ({ ...t, editingMessageId: null }));
					saveUpdatedConversation(tabId, updatedChat);

					if (selectedTabIdRef.current === tabId) scrollToBottomSoon(tabId);

					void updateStreamingMessage(tabId, updatedChat, options).catch(console.error);
					return;
				}

				// message vanished -> clear edit state and append normally
				updateTab(tabId, t => ({ ...t, editingMessageId: null }));
			}

			const updated: Conversation = {
				...tab.conversation,
				messages: [...tab.conversation.messages, userMsg],
				modifiedAt: new Date(),
			};

			saveUpdatedConversation(tabId, updated);
			if (selectedTabIdRef.current === tabId) scrollToBottomSoon(tabId);

			void updateStreamingMessage(tabId, updated, options).catch(console.error);
		},
		[saveUpdatedConversation, scrollToBottomSoon, updateStreamingMessage, updateTab]
	);

	const beginEditMessageForTab = useCallback(
		(tabId: string, id: string) => {
			const tab = tabsRef.current.find(t => t.tabId === tabId);
			if (!tab) return;
			if (tab.isBusy) return;

			const msg = tab.conversation.messages.find(m => m.id === id);
			if (!msg) return;
			if (msg.role !== RoleEnum.User) return;

			const external: EditorExternalMessage = {
				text: msg.uiContent ?? '',
				attachments: msg.attachments,
				toolChoices: msg.toolStoreChoices,
				toolOutputs: msg.uiToolOutputs,
			};

			const input = inputRefs.current.get(tabId);
			input?.loadExternalMessage(external);
			input?.focus();

			updateTab(tabId, t => ({ ...t, editingMessageId: id }));
		},
		[updateTab]
	);

	const cancelEditingForTab = useCallback(
		(tabId: string) => {
			updateTab(tabId, t => ({ ...t, editingMessageId: null }));
		},
		[updateTab]
	);

	const handleResendForTab = useCallback(
		async (tabId: string, id: string) => {
			const tab = tabsRef.current.find(t => t.tabId === tabId);
			if (!tab) return;
			if (tab.isBusy) return;

			const idx = tab.conversation.messages.findIndex(m => m.id === id);
			if (idx === -1) return;

			const msg = tab.conversation.messages[idx];
			const input = inputRefs.current.get(tabId);

			if (msg.role === RoleEnum.User && msg.toolStoreChoices && msg.toolStoreChoices.length > 0) {
				input?.setConversationToolsFromChoices(msg.toolStoreChoices);
				input?.setWebSearchFromChoices(msg.toolStoreChoices);
			} else {
				input?.setConversationToolsFromChoices([]);
				input?.setWebSearchFromChoices([]);
			}

			const msgs = tab.conversation.messages.slice(0, idx + 1);
			const updated = { ...tab.conversation, messages: msgs, modifiedAt: new Date() };
			saveUpdatedConversation(tabId, updated);

			let opts = { ...DefaultUIChatOptions };
			if (input) opts = input.getUIChatOptions();

			await updateStreamingMessage(tabId, updated, opts);
		},
		[saveUpdatedConversation, updateStreamingMessage]
	);

	// ---------------- Shortcuts ----------------
	const [shortcutConfig] = useState<ShortcutConfig>(defaultShortcutConfig);

	useChatShortcuts({
		config: shortcutConfig,
		isBusy: false, // new tabs allowed even if another tab is busy
		handlers: {
			newChat: () => {
				openNewTab();
			},
			focusSearch: () => searchRef.current?.focusInput(),
			focusInput: () => inputRefs.current.get(selectedTabIdRef.current)?.focus(),
			insertTemplate: () => inputRefs.current.get(selectedTabIdRef.current)?.openTemplateMenu(),
			insertTool: () => inputRefs.current.get(selectedTabIdRef.current)?.openToolMenu(),
			insertAttachment: () => inputRefs.current.get(selectedTabIdRef.current)?.openAttachmentMenu(),
		},
	});

	// ---------------- Titlebar: search only (like before) ----------------
	useTitleBarContent(
		{
			center: (
				<div className="w-[min(720px,60vw)]">
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

	const activeRenderedMessages = useMemo(() => {
		if (!activeTab) return null;

		return activeTab.conversation.messages.map((msg, idx) => {
			const isLast = idx === activeTab.conversation.messages.length - 1;

			const isPending =
				activeTab.isBusy && isLast && msg.role === RoleEnum.Assistant && (msg.uiContent?.length ?? 0) === 0;

			const live = activeTab.isBusy && isLast && msg.role === RoleEnum.Assistant ? activeStreamText : '';

			return (
				<ChatMessage
					key={msg.id}
					message={msg}
					onEdit={() => {
						beginEditMessageForTab(activeTab.tabId, msg.id);
					}}
					onResend={() => void handleResendForTab(activeTab.tabId, msg.id)}
					streamedMessage={live}
					isPending={isPending}
					isBusy={activeTab.isBusy}
					isEditing={activeTab.editingMessageId === msg.id}
				/>
			);
		});
	}, [activeStreamText, activeTab, beginEditMessageForTab, handleResendForTab]);

	return (
		<PageFrame contentScrollable={false}>
			<div className="grid h-full w-full grid-rows-[auto_1fr_auto] overflow-hidden">
				{/* Row 1: TAB STRIP (where navbar used to be) + Download floater */}
				<div className="relative row-start-1 row-end-2 flex w-full justify-center p-0">
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

				{/* Row 2: MESSAGES (single scroll container; scroll position restored per tab) */}
				<div className="relative row-start-2 row-end-3 mt-2 min-h-0">
					<div
						ref={chatContainerRef}
						onScroll={onScrollActive}
						className="relative h-full w-full overflow-y-auto overscroll-contain py-1"
						style={{ scrollbarGutter: 'stable both-edges' }}
					>
						<div className="mx-auto w-11/12 xl:w-5/6">
							<div className="space-y-4">{activeRenderedMessages}</div>
						</div>
					</div>

					{/* Overlay scroll buttons (active tab only, since container is shared) */}
					<div className="pointer-events-none absolute right-4 bottom-16 z-10 xl:right-24">
						<div className="pointer-events-auto">
							{isScrollable && !isAtTop ? (
								<ButtonScrollToTop
									scrollContainerRef={chatContainerRef}
									iconSize={32}
									show={isScrollable && !isAtTop}
									className="btn btn-md border-none bg-transparent shadow-none"
								/>
							) : null}
						</div>
					</div>
					<div className="pointer-events-none absolute right-4 bottom-4 z-10 xl:right-24">
						<div className="pointer-events-auto">
							{isScrollable && !isAtBottom ? (
								<ButtonScrollToBottom
									scrollContainerRef={chatContainerRef}
									iconSize={32}
									show={isScrollable && !isAtBottom}
									className="btn btn-md border-none bg-transparent shadow-none"
								/>
							) : null}
						</div>
					</div>
				</div>

				{/* Row 3: INPUT (per tab; all mounted, only active visible) */}
				<div className="row-start-3 row-end-4 flex w-full min-w-0 justify-center">
					<div className="w-11/12 min-w-0 xl:w-5/6">
						{tabs.map(t => (
							<div key={t.tabId} className={t.tabId === selectedTabId ? 'block' : 'hidden'}>
								<InputBox
									ref={setInputRef(t.tabId)}
									onSend={(payload, options) => {
										return sendMessageForTab(t.tabId, payload, options);
									}}
									isBusy={t.isBusy}
									abortRef={getAbortRef(t.tabId)}
									shortcutConfig={shortcutConfig}
									editingMessageId={t.editingMessageId}
									cancelEditing={() => {
										cancelEditingForTab(t.tabId);
									}}
								/>
							</div>
						))}
					</div>
				</div>
			</div>
		</PageFrame>
	);
}
