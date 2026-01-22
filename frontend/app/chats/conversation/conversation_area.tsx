/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import {
	forwardRef,
	useCallback,
	useEffect,
	useImperativeHandle,
	useLayoutEffect,
	useMemo,
	useRef,
	useState,
} from 'react';

import type { Conversation, ConversationMessage } from '@/spec/conversation';
import { ContentItemKind, type ModelParam, OutputKind, type OutputUnion, RoleEnum, Status } from '@/spec/inference';
import { DefaultUIChatOptions, type UIChatOption } from '@/spec/modelpreset';
import { type ToolStoreChoice, ToolStoreChoiceType } from '@/spec/tool';

import type { ShortcutConfig } from '@/lib/keyboard_shortcuts';
import { getBlockQuotedLines } from '@/lib/text_utils';

import { useAtTopBottom } from '@/hooks/use_at_top_bottom';

import { ButtonScrollToBottom, ButtonScrollToTop } from '@/components/button_scroll_top_bottom';

import type { ChatTabState } from '@/chats/chat_tabs_persist';
import { HandleCompletion } from '@/chats/conversation/completion_helper';
import {
	buildUserConversationMessageFromEditor,
	deriveConversationToolsFromMessages,
	deriveWebSearchChoiceFromMessages,
	initConversationMessage,
} from '@/chats/conversation/hydration_helper';
import type { InputBoxHandle } from '@/chats/inputarea/input_box';
import type { EditorExternalMessage, EditorSubmitPayload } from '@/chats/inputarea/input_editor_utils';
import { InputPane } from '@/chats/inputarea/input_pane';
import { ChatMessage } from '@/chats/messages/message';

export type ConversationAreaHandle = {
	disposeTabRuntime: (tabId: string) => void;
	clearStreamForTab: (tabId: string) => void;
	syncComposerFromConversation: (tabId: string, conv: Conversation) => void;

	focusInput: (tabId: string) => void;
	openTemplateMenu: (tabId: string) => void;
	openToolMenu: (tabId: string) => void;
	openAttachmentMenu: (tabId: string) => void;

	setScrollTopForTab: (tabId: string, top: number) => void;
	resetScrollToTop: (tabId: string) => void;
	getScrollTopByTabSnapshot: () => Record<string, number>;
};

type ConversationAreaProps = {
	tabs: ChatTabState[];
	selectedTabId: string;

	// Used by InputPane (unchanged API)
	shortcutConfig: ShortcutConfig;

	// Persisted scroll restore seed (from storage)
	initialScrollTopByTab?: Record<string, number>;

	// State mutations remain in the page; this component owns "conversation runtime":
	// streaming buffers, abort controllers, input refs, scroll restoration, send/edit/resend.
	updateTab: (tabId: string, updater: (t: ChatTabState) => ChatTabState) => void;
	saveUpdatedConversation: (tabId: string, updatedConv: Conversation, titleWasExternallyChanged?: boolean) => void;
};

export const ConversationArea = forwardRef<ConversationAreaHandle, ConversationAreaProps>(function ConversationArea(
	{ tabs, selectedTabId, shortcutConfig, initialScrollTopByTab, updateTab, saveUpdatedConversation },
	ref
) {
	// ---------------- Tabs ref (for async safety) ----------------
	const tabsRef = useRef(tabs);
	useEffect(() => {
		tabsRef.current = tabs;
	}, [tabs]);

	const selectedTabIdRef = useRef(selectedTabId);
	useEffect(() => {
		selectedTabIdRef.current = selectedTabId;
	}, [selectedTabId]);

	const activeTab = useMemo(() => tabs.find(t => t.tabId === selectedTabId) ?? tabs[0], [tabs, selectedTabId]);

	const tabExists = useCallback((tabId: string) => tabsRef.current.some(t => t.tabId === tabId), []);

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

	// Seed scroll positions from persisted state (once)
	const seededScrollFromStorageRef = useRef(false);
	if (!seededScrollFromStorageRef.current) {
		seededScrollFromStorageRef.current = true;
		if (initialScrollTopByTab) {
			for (const [id, top] of Object.entries(initialScrollTopByTab)) {
				if (typeof top === 'number') scrollTopByTab.current.set(id, top);
			}
		}
	}

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

			if (selectedTabIdRef.current === tabId) setActiveStreamText('');
		},
		[getAbortRef]
	);

	const clearStreamForTab = useCallback(
		(tabId: string) => {
			getStreamTextRef(tabId).current = '';
			if (selectedTabIdRef.current === tabId) setActiveStreamText('');
		},
		[getStreamTextRef]
	);

	const syncComposerFromConversation = useCallback((tabId: string, conv: Conversation) => {
		const input = inputRefs.current.get(tabId);
		if (!input) return;

		const tools = deriveConversationToolsFromMessages(conv.messages);
		const web = deriveWebSearchChoiceFromMessages(conv.messages);
		input.setConversationToolsFromChoices(tools);
		input.setWebSearchFromChoices(web);
	}, []);

	const focusInput = useCallback((tabId: string) => inputRefs.current.get(tabId)?.focus(), []);
	const openTemplateMenu = useCallback((tabId: string) => inputRefs.current.get(tabId)?.openTemplateMenu(), []);
	const openToolMenu = useCallback((tabId: string) => inputRefs.current.get(tabId)?.openToolMenu(), []);
	const openAttachmentMenu = useCallback((tabId: string) => inputRefs.current.get(tabId)?.openAttachmentMenu(), []);

	const setScrollTopForTab = useCallback((tabId: string, top: number) => {
		scrollTopByTab.current.set(tabId, top);
	}, []);

	// ---------------- UI refs ----------------
	const chatContainerRef = useRef<HTMLDivElement>(null);
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

	useEffect(() => {
		return () => {
			if (pendingActiveStreamSync.current !== null) {
				window.clearTimeout(pendingActiveStreamSync.current);
				pendingActiveStreamSync.current = null;
			}
		};
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

	// Restore scroll position when changing tabs (scroll restore requirement)
	useLayoutEffect(() => {
		const el = chatContainerRef.current;
		if (!el) return;

		const top = scrollTopByTab.current.get(selectedTabId) ?? 0;

		// Keep map consistent even if user never scrolls.
		scrollTopByTab.current.set(selectedTabId, top);

		// Use rAF to ensure DOM layout is committed for the new tab content
		requestAnimationFrame(() => {
			const el2 = chatContainerRef.current;
			if (!el2) return;
			el2.scrollTop = top;
		});
	}, [selectedTabId]);

	const resetScrollToTop = useCallback((tabId: string) => {
		scrollTopByTab.current.set(tabId, 0);
		if (selectedTabIdRef.current !== tabId) return;
		requestAnimationFrame(() => {
			const el = chatContainerRef.current;
			if (el) el.scrollTop = 0;
		});
	}, []);

	const getScrollTopByTabSnapshot = useCallback(() => {
		const obj: Record<string, number> = {};
		for (const [k, v] of scrollTopByTab.current.entries()) obj[k] = v;
		return obj;
	}, []);

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
			saveUpdatedConversation,
			scheduleActiveStreamSync,
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

	// ---------------- Expose imperative API to ChatsPage ----------------
	useImperativeHandle(
		ref,
		() => ({
			disposeTabRuntime,
			clearStreamForTab,
			syncComposerFromConversation,
			focusInput,
			openTemplateMenu,
			openToolMenu,
			openAttachmentMenu,
			setScrollTopForTab,
			resetScrollToTop,
			getScrollTopByTabSnapshot,
		}),
		[
			clearStreamForTab,
			disposeTabRuntime,
			focusInput,
			getScrollTopByTabSnapshot,
			openAttachmentMenu,
			openTemplateMenu,
			openToolMenu,
			resetScrollToTop,
			setScrollTopForTab,
			syncComposerFromConversation,
		]
	);

	// ---------------- Render helpers ----------------
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
		<>
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
						<InputPane
							key={t.tabId}
							tabId={t.tabId}
							active={t.tabId === selectedTabId}
							isBusy={t.isBusy}
							editingMessageId={t.editingMessageId}
							setInputRef={setInputRef}
							getAbortRef={getAbortRef}
							shortcutConfig={shortcutConfig}
							sendMessage={sendMessageForTab}
							cancelEditing={cancelEditingForTab}
						/>
					))}
				</div>
			</div>
		</>
	);
});
