/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import { useCallback, useEffect, useRef, useState } from 'react';

import {
	type Conversation,
	type ConversationMessage,
	type ConversationSearchItem,
	type StoreConversation,
	type StoreConversationMessage,
} from '@/spec/conversation';
import { ContentItemKind, type ModelParam, OutputKind, type OutputUnion, RoleEnum, Status } from '@/spec/inference';
import { type ToolStoreChoice, ToolStoreChoiceType } from '@/spec/tool';

import { defaultShortcutConfig, type ShortcutConfig, useChatShortcuts } from '@/lib/keyboard_shortcuts';
import { getBlockQuotedLines, sanitizeConversationTitle } from '@/lib/text_utils';
import { generateTitle } from '@/lib/title_utils';

import { useAtTopBottom } from '@/hooks/use_at_top_bottom';

import { conversationStoreAPI } from '@/apis/baseapi';

import { ButtonScrollToBottom, ButtonScrollToTop } from '@/components/button_scroll_top_bottom';
import { PageFrame } from '@/components/page_frame';

import { HandleCompletion } from '@/chats/chat_completion_helper';
import {
	buildUserConversationMessageFromEditor,
	deriveConversationToolsFromMessages,
	hydrateConversation,
	initConversation,
	initConversationMessage,
} from '@/chats/chat_hydration_helper';
import { InputBox, type InputBoxHandle } from '@/chats/chat_input_box';
import type { EditorExternalMessage, EditorSubmitPayload } from '@/chats/chat_input_editor';
import { ChatNavBar, type ChatNavBarHandle } from '@/chats/chat_navbar';
import { type ChatOption, DefaultChatOptions } from '@/chats/chat_option_helper';
import { ChatMessage } from '@/chats/messages/message';

// eslint-disable-next-line no-restricted-exports
export default function ChatsPage() {
	const [chat, setChat] = useState<Conversation>(initConversation());
	const [searchRefreshKey, setSearchRefreshKey] = useState(0);

	const [streamedMessage, setStreamedMessage] = useState('');
	const [isBusy, setIsBusy] = useState(false);

	/* Currently running request */
	const abortRef = useRef<AbortController | null>(null);
	const requestIdRef = useRef<string | null>(null); // will go to backend for cancellation

	const chatInputRef = useRef<InputBoxHandle>(null);
	const chatContainerRef = useRef<HTMLDivElement>(null);
	const navRef = useRef<ChatNavBarHandle | null>(null);

	const tokensReceivedRef = useRef<boolean | null>(false);

	// Has the current conversation already been persisted?
	const isChatPersistedRef = useRef(false);
	const { isAtBottom, isAtTop, isScrollable } = useAtTopBottom(chatContainerRef);
	const manualTitleLockedRef = useRef(false);

	const [shortcutConfig] = useState<ShortcutConfig>(defaultShortcutConfig);
	const [editingMessageId, setEditingMessageId] = useState<string | null>(null);

	// Focus on mount.
	useEffect(() => {
		chatInputRef.current?.focus();
	}, []);

	// Scroll helper used only when the user explicitly starts a new turn
	// (send / resend / edited send). We *do not* auto-scroll on every
	// message list change any more.
	const scrollToBottom = useCallback(() => {
		const el = chatContainerRef.current;
		if (!el) return;
		el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' });
	}, []);

	const scrollToBottomSoon = useCallback(() => {
		// Give React a short window to commit DOM changes
		window.setTimeout(scrollToBottom, 80);
	}, [scrollToBottom]);

	const bumpSearchKey = async () => {
		await new Promise(resolve => setTimeout(resolve, 50));
		setSearchRefreshKey(k => k + 1);
	};

	const handleNewChat = async () => {
		if (isBusy) return;
		if (chat.messages.length === 0) {
			chatInputRef.current?.focus();
			return;
		}
		// Put the old conversation _fully_ before moving to new chat.
		// Handles edge cases like someone calls newchat when streaming is ongoing,
		// or if there is some bug below that forgets to call saveUpdated chat
		conversationStoreAPI.putConversation(chat as StoreConversation);
		setChat(initConversation());
		chatInputRef.current?.setConversationToolsFromChoices([]);

		// New non-persisted conversation started.
		isChatPersistedRef.current = false;
		manualTitleLockedRef.current = false;

		setEditingMessageId(null);

		chatInputRef.current?.focus();
	};

	const handleSelectConversation = useCallback(async (item: ConversationSearchItem) => {
		const selectedChat = await conversationStoreAPI.getConversation(item.id, item.title, true);
		if (selectedChat) {
			// Hydrate store-level data into full UI Conversation
			const hydrated = hydrateConversation(selectedChat);
			setChat(hydrated);

			isChatPersistedRef.current = true;
			manualTitleLockedRef.current = false;
			setEditingMessageId(null);

			const initialTools = deriveConversationToolsFromMessages(hydrated.messages);
			chatInputRef.current?.setConversationToolsFromChoices(initialTools);
		}
	}, []);

	const getConversationForExport = useCallback(async (): Promise<string> => {
		const selectedChat = await conversationStoreAPI.getConversation(chat.id, chat.title, true);
		return JSON.stringify(selectedChat ?? null, null, 2);
	}, [chat.id, chat.title]);

	// Persist `updatedChat` using the cheapest API that is still correct.
	// •  First time we ever write this conversation              -> putConversation
	// •  Title has changed (search index must be updated)        -> putConversation
	// •  Otherwise                                               -> putMessagesToConversation
	// putMessagesToConversation REQUIRES the *full* message list, so we just pass `updatedChat.messages` every time.
	const saveUpdatedChat = (updatedChat: Conversation, titleWasExternallyChanged = false) => {
		let newTitle = updatedChat.title;
		if (!manualTitleLockedRef.current && updatedChat.messages.length <= 4) {
			const userMessages = updatedChat.messages.filter(m => m.role === RoleEnum.User);
			if (userMessages.length === 1) {
				// Always generate title from first user message
				const t = generateTitle(userMessages[0].uiContent);
				newTitle = t.title;
			} else if (userMessages.length === 2) {
				// Generate titles from both messages, pick the one with higher score
				const titleCondidate1 = generateTitle(userMessages[0].uiContent);
				const titleCondidate2 = generateTitle(userMessages[1].uiContent);
				newTitle = titleCondidate2.score > titleCondidate1.score ? titleCondidate2.title : titleCondidate1.title;
			}
			newTitle = sanitizeConversationTitle(newTitle);
		}

		const titleChangedByFunction = newTitle !== updatedChat.title;
		if (titleChangedByFunction) {
			updatedChat.title = newTitle;
		}

		const titleChanged = titleWasExternallyChanged || titleChangedByFunction;

		// Decide which API to call
		if (!isChatPersistedRef.current) {
			// 1st save -> create the record + metadata work
			conversationStoreAPI.putConversation(updatedChat as StoreConversation);
			isChatPersistedRef.current = true;
			bumpSearchKey(); // now searchable
		} else if (titleChanged) {
			// metadata (title) changed -> need the heavy PUT
			conversationStoreAPI.putConversation(updatedChat as StoreConversation);
			bumpSearchKey();
		} else {
			// normal case -> only messages changed
			conversationStoreAPI.putMessagesToConversation(
				updatedChat.id,
				updatedChat.title,
				updatedChat.messages as StoreConversationMessage[]
			);
		}

		// update local state
		setChat({ ...updatedChat, messages: [...updatedChat.messages] });
	};

	const handleRenameTitle = useCallback(
		(newTitle: string) => {
			const sanitized = sanitizeConversationTitle(newTitle.trim());
			if (!sanitized || sanitized === chat.title) return;

			manualTitleLockedRef.current = true;

			const updatedChat: Conversation = {
				...chat,
				title: sanitized,
				modifiedAt: new Date(),
			};

			saveUpdatedChat(updatedChat, true);
		},
		[chat, saveUpdatedChat]
	);

	const updateStreamingMessage = useCallback(
		async (updatedChatWithUserMessage: Conversation, options: ChatOption) => {
			// Abort older request (if any) and reset UI
			abortRef.current?.abort();
			tokensReceivedRef.current = false;
			setIsBusy(true);

			abortRef.current = new AbortController();
			requestIdRef.current = crypto.randomUUID();

			let allMessages = updatedChatWithUserMessage.messages;
			if (options.disablePreviousMessages) {
				// Only keep the last user turn as context.
				allMessages = [updatedChatWithUserMessage.messages[updatedChatWithUserMessage.messages.length - 1]];
			}

			if (allMessages.length === 0) {
				setIsBusy(false);
				return;
			}

			const currentUserMsg = allMessages[allMessages.length - 1];
			const history = allMessages.slice(0, allMessages.length - 1);

			// Local accumulator for this request
			let aggregatedStreamText = '';
			setStreamedMessage('');

			// Create assistant placeholder for streaming.
			const assistantPlaceholder = initConversationMessage(RoleEnum.Assistant);
			const chatWithPlaceholder: Conversation = {
				...updatedChatWithUserMessage,
				messages: [...updatedChatWithUserMessage.messages, assistantPlaceholder],
				modifiedAt: new Date(),
			};

			// Show empty assistant bubble immediately.
			setChat({ ...chatWithPlaceholder, messages: [...chatWithPlaceholder.messages] });
			scrollToBottomSoon();

			const onStreamTextData = (textData: string) => {
				if (!textData) return;
				tokensReceivedRef.current = true;
				aggregatedStreamText += textData;
				setStreamedMessage(prev => prev + textData);
			};

			const onStreamThinkingData = (thinkingData: string) => {
				if (!thinkingData) return;
				tokensReceivedRef.current = true;
				const block = getBlockQuotedLines(thinkingData) + '\n';
				aggregatedStreamText += block;
				setStreamedMessage(prev => prev + block);
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
					requestIdRef.current ?? undefined,
					abortRef.current.signal,
					onStreamTextData,
					onStreamThinkingData
				);

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

					saveUpdatedChat(finalChat);

					// Expose *runnable* assistant-suggested tool calls (function/custom)
					// as chips in the composer. Web-search tool calls are provider-managed
					// and remain informational in the message history only.
					if (responseMessage.uiToolCalls && responseMessage.uiToolCalls.length > 0) {
						const runnableCalls = responseMessage.uiToolCalls.filter(
							c => c.type === ToolStoreChoiceType.Function || c.type === ToolStoreChoiceType.Custom
						);
						if (runnableCalls.length > 0) {
							chatInputRef.current?.loadToolCalls(runnableCalls);
						}
					}
				}
			} catch (e) {
				if ((e as DOMException).name === 'AbortError') {
					if (!tokensReceivedRef.current) {
						setChat(c => {
							const idx = c.messages.findIndex(m => m.id === assistantPlaceholder.id);
							if (idx === -1) return c;
							const msgs = c.messages.filter((_, i) => i !== idx);
							return { ...c, messages: msgs, modifiedAt: new Date() };
						});
					} else {
						const partialText = aggregatedStreamText + '\n\n>API Aborted after partial response...';

						const partialOutputs: OutputUnion[] = [
							{
								kind: OutputKind.OutputMessage,
								outputMessage: {
									id: assistantPlaceholder.id,
									role: RoleEnum.Assistant,
									status: Status.Completed,
									contents: [
										{
											kind: ContentItemKind.Text,
											textItem: { text: partialText },
										},
									],
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
						saveUpdatedChat(finalChat);
					}
				} else {
					console.error(e);
				}
			} finally {
				setStreamedMessage('');
				setIsBusy(false);
			}
		},
		[saveUpdatedChat, scrollToBottomSoon]
	);

	const sendMessage = async (payload: EditorSubmitPayload, options: ChatOption) => {
		if (isBusy) return;

		const trimmed = payload.text.trim();
		// Allow "empty-text" turns ONLY when there is at least attachments or tool outputs.
		// Tool choices alone (explicit or conversation-level) are NOT enough.
		const hasNonEmptyText = trimmed.length > 0;
		const hasToolOutputs = payload.toolOutputs.length > 0;
		const hasAttachments = payload.attachments.length > 0;

		if (!hasNonEmptyText && !hasToolOutputs && !hasAttachments) {
			return;
		}

		const editingId = editingMessageId ?? undefined;
		const userMsg = buildUserConversationMessageFromEditor(payload, editingId);

		// If we are editing an existing user message, replace it and drop later messages.
		if (editingMessageId) {
			const idx = chat.messages.findIndex(m => m.id === editingMessageId);
			// If somehow the message vanished, fall back to "append new".
			if (idx !== -1) {
				const msgs = chat.messages.slice(0, idx + 1);
				msgs[idx] = userMsg;

				const updatedChat: Conversation = {
					...chat,
					messages: msgs,
					modifiedAt: new Date(),
				};

				setEditingMessageId(null);
				saveUpdatedChat(updatedChat);

				// Fire-and-forget streaming; this will set isBusy and manage
				// aborts, but we do not block the composer on it.
				void (async () => {
					try {
						await updateStreamingMessage(updatedChat, options);
					} catch (err) {
						console.error(err);
					}
				})();
				return;
			}

			// If not found, clear edit state and proceed as a normal send.
			setEditingMessageId(null);
		}

		const updated: Conversation = {
			...chat,
			messages: [...chat.messages, userMsg],
			modifiedAt: new Date(),
		};

		saveUpdatedChat(updated);

		// Same fire-and-forget behavior for normal sends.
		void (async () => {
			try {
				await updateStreamingMessage(updated, options);
			} catch (err) {
				console.error(err);
			}
		})();
	};

	const beginEditMessage = useCallback(
		(id: string) => {
			if (isBusy) return;

			const msg = chat.messages.find(m => m.id === id);
			if (!msg) return;
			if (msg.role !== RoleEnum.User) return;

			const external: EditorExternalMessage = {
				text: msg.uiContent ?? '',
				attachments: msg.attachments,
				toolChoices: msg.toolStoreChoices,
				toolOutputs: msg.uiToolOutputs,
			};

			chatInputRef.current?.loadExternalMessage(external);
			chatInputRef.current?.focus();
			setEditingMessageId(id);
		},
		[chat.messages, isBusy]
	);

	const cancelEditing = useCallback(() => {
		setEditingMessageId(null);
	}, []);

	const handleResend = useCallback(
		async (id: string) => {
			if (isBusy) return;
			const idx = chat.messages.findIndex(m => m.id === id);
			if (idx === -1) return;
			const msg = chat.messages[idx];
			if (msg.role === RoleEnum.User && msg.toolStoreChoices && msg.toolStoreChoices.length > 0) {
				chatInputRef.current?.setConversationToolsFromChoices(msg.toolStoreChoices);
			} else {
				chatInputRef.current?.setConversationToolsFromChoices([]);
			}
			const msgs = chat.messages.slice(0, idx + 1);
			const updated = { ...chat, messages: msgs, modifiedAt: new Date() };
			saveUpdatedChat(updated);

			let opts = { ...DefaultChatOptions };
			if (chatInputRef.current) opts = chatInputRef.current.getChatOptions();
			await updateStreamingMessage(updated, opts);
		},
		[chat, saveUpdatedChat, updateStreamingMessage]
	);

	const renderedMessages = chat.messages.map((msg, idx) => {
		const isPending =
			isBusy && idx === chat.messages.length - 1 && msg.role === RoleEnum.Assistant && msg.uiContent.length === 0;
		const live = isBusy && idx === chat.messages.length - 1 && msg.role === RoleEnum.Assistant ? streamedMessage : '';

		return (
			<ChatMessage
				key={msg.id}
				message={msg}
				onEdit={() => {
					beginEditMessage(msg.id);
				}}
				onResend={() => handleResend(msg.id)}
				streamedMessage={live}
				isPending={isPending}
				isBusy={isBusy}
				isEditing={editingMessageId === msg.id}
			/>
		);
	});

	useChatShortcuts({
		config: shortcutConfig,
		isBusy,
		handlers: {
			newChat: () => {
				void handleNewChat();
			},
			focusSearch: () => {
				navRef.current?.focusSearch();
			},
			focusInput: () => {
				chatInputRef.current?.focus();
			},
			insertTemplate: () => {
				chatInputRef.current?.openTemplateMenu();
			},
			insertTool: () => {
				chatInputRef.current?.openToolMenu();
			},
			insertAttachment: () => {
				chatInputRef.current?.openAttachmentMenu();
			},
		},
	});

	return (
		<PageFrame contentScrollable={false}>
			<div className="grid h-full w-full grid-rows-[auto_1fr_auto] overflow-hidden">
				{/* Row 1: NAVBAR */}
				<div className="row-start-1 row-end-2 flex w-full justify-center">
					<ChatNavBar
						ref={navRef}
						onNewChat={handleNewChat}
						onRenameTitle={handleRenameTitle}
						getConversationForExport={getConversationForExport}
						onSelectConversation={handleSelectConversation}
						chatTitle={chat.title}
						chatID={chat.id}
						searchRefreshKey={searchRefreshKey}
						disabled={isBusy}
						renameEnabled={chat.messages.length > 0}
						shortcutConfig={shortcutConfig}
					/>
				</div>

				{/* Row 2: MESSAGES (the only scrollable area) */}
				<div className="relative row-start-2 row-end-3 min-h-0">
					{/* Make the full row the scroll container so the scrollbar is at far right */}
					<div
						ref={chatContainerRef}
						className="relative h-full w-full overflow-y-auto overscroll-contain py-1"
						style={{ scrollbarGutter: 'stable both-edges' }}
					>
						{/* Center the content inside the full-width scroll container */}
						<div className="mx-auto w-11/12 xl:w-5/6">
							<div className="space-y-4">{renderedMessages}</div>
						</div>
					</div>

					{/* Overlay the buttons; not part of the scrollable content */}
					<div className="pointer-events-none absolute right-4 bottom-16 z-10 xl:right-24">
						<div className="pointer-events-auto">
							{isScrollable && !isAtTop && (
								<ButtonScrollToTop
									scrollContainerRef={chatContainerRef}
									iconSize={32}
									show={isScrollable && !isAtTop}
									className="btn btn-md border-none bg-transparent shadow-none"
								/>
							)}
						</div>
					</div>
					<div className="pointer-events-none absolute right-4 bottom-4 z-10 xl:right-24">
						<div className="pointer-events-auto">
							{isScrollable && !isAtBottom && (
								<ButtonScrollToBottom
									scrollContainerRef={chatContainerRef}
									iconSize={32}
									show={isScrollable && !isAtBottom}
									className="btn btn-md border-none bg-transparent shadow-none"
								/>
							)}
						</div>
					</div>
				</div>

				{/* Row 3: INPUT (auto; grows with content) */}
				<div className="row-start-3 row-end-4 flex w-full min-w-0 justify-center">
					<div className="w-11/12 min-w-0 xl:w-5/6">
						<InputBox
							ref={chatInputRef}
							onSend={sendMessage}
							isBusy={isBusy}
							abortRef={abortRef}
							shortcutConfig={shortcutConfig}
							editingMessageId={editingMessageId}
							cancelEditing={cancelEditing}
						/>
					</div>
				</div>
			</div>
		</PageFrame>
	);
}
