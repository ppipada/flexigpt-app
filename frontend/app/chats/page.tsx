import { useCallback, useEffect, useRef, useState } from 'react';

import type { ModelParams } from '@/spec/aiprovider';
import type { Conversation, ConversationMessage, ConversationSearchItem } from '@/spec/conversation';
import { ConversationRoleEnum } from '@/spec/conversation';
import type { ToolChoice } from '@/spec/tool';

import { defaultShortcutConfig, type ShortcutConfig, useChatShortcuts } from '@/lib/keyboard_shortcuts';
import { getBlockQuotedLines, sanitizeConversationTitle } from '@/lib/text_utils';
import { generateTitle } from '@/lib/title_utils';
import { getUUIDv7 } from '@/lib/uuid_utils';

import { useAtBottom } from '@/hooks/use_at_bottom';

import { BuildCompletionDataFromConversation, HandleCompletion } from '@/apis/aiprovider_helper';
import { conversationStoreAPI } from '@/apis/baseapi';
import { type ChatOption, DefaultChatOptions } from '@/apis/chatoption_helper';

import { ButtonScrollToBottom } from '@/components/button_scroll_to_bottom';
import { PageFrame } from '@/components/page_frame';

import { editorAttachmentToConversation } from '@/chats/attachments/editor_attachment_utils';
import type { EditorAttachedToolChoice } from '@/chats/attachments/tool_editor_utils';
import { InputBox, type InputBoxHandle } from '@/chats/chat_input_box';
import type { EditorExternalMessage, EditorSubmitPayload } from '@/chats/chat_input_editor';
import { ChatNavBar, type ChatNavBarHandle } from '@/chats/chat_navbar';
import { ChatMessage } from '@/chats/messages/message';

function initConversation(title = 'New Conversation'): Conversation {
	return {
		id: getUUIDv7(),
		title: generateTitle(title).title,
		createdAt: new Date(),
		modifiedAt: new Date(),
		messages: [],
	};
}

function initConversationMessage(role: ConversationRoleEnum, content: string): ConversationMessage {
	const d = new Date();
	return {
		id: d.toISOString(),
		createdAt: new Date(),
		role,
		content,
	};
}

function attachedToolToConversationToolChoice(tool: EditorAttachedToolChoice): ToolChoice {
	return {
		bundleID: tool.bundleID,
		toolSlug: tool.toolSlug,
		toolVersion: tool.toolVersion,
		displayName: tool.displayName,
		description: tool.description,
		toolID: tool.toolID,
	};
}

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
	const { isAtBottom, isScrollable, checkScroll } = useAtBottom(chatContainerRef);
	const manualTitleLockedRef = useRef(false);

	const [shortcutConfig] = useState<ShortcutConfig>(defaultShortcutConfig);
	const [editingMessageId, setEditingMessageId] = useState<string | null>(null);

	// Focus on mount.
	useEffect(() => {
		chatInputRef.current?.focus();
	}, []);

	// Get the scroll down button active via regular check.
	useEffect(() => {
		if (!isBusy) return;

		const interval = setInterval(() => {
			checkScroll();
		}, 100);

		return () => {
			clearInterval(interval);
		};
	}, [isBusy, checkScroll]);

	// Scroll to bottom. Tell the browser to bring the sentinel into view.
	const scrollToBottom = () => {
		if (chatContainerRef.current) {
			chatContainerRef.current.scrollTo({ top: chatContainerRef.current.scrollHeight, behavior: 'smooth' });
		}
	};
	useEffect(() => {
		const t = setTimeout(scrollToBottom, 100);
		return () => {
			clearTimeout(t);
		};
	}, [chat.messages]);

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
		conversationStoreAPI.putConversation(chat);
		setChat(initConversation());
		// New non-persisted conversation started.
		isChatPersistedRef.current = false;
		manualTitleLockedRef.current = false;

		setEditingMessageId(null);

		chatInputRef.current?.focus();
	};

	// Persist `updatedChat` using the cheapest API that is still correct.
	// •  First time we ever write this conversation              -> putConversation
	// •  Title has changed (search index must be updated)        -> putConversation
	// •  Otherwise                                               -> putMessagesToConversation
	// putMessagesToConversation REQUIRES the *full* message list, so we just pass `updatedChat.messages` every time.
	const saveUpdatedChat = (updatedChat: Conversation, titleWasExternallyChanged = false) => {
		let newTitle = updatedChat.title;
		if (!manualTitleLockedRef.current && updatedChat.messages.length <= 4) {
			const userMessages = updatedChat.messages.filter(m => m.role === ConversationRoleEnum.user);
			if (userMessages.length === 1) {
				// Always generate title from first user message
				newTitle = generateTitle(userMessages[0].content).title;
			} else if (userMessages.length === 2) {
				// Generate titles from both messages, pick the one with higher score
				const titleCondidate1 = generateTitle(userMessages[0].content);
				const titleCondidate2 = generateTitle(userMessages[1].content);
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
			conversationStoreAPI.putConversation(updatedChat);
			isChatPersistedRef.current = true;
			bumpSearchKey(); // now searchable
		} else if (titleChanged) {
			// metadata (title) changed -> need the heavy PUT
			conversationStoreAPI.putConversation(updatedChat);
			bumpSearchKey();
		} else {
			// normal case -> only messages changed
			conversationStoreAPI.putMessagesToConversation(updatedChat.id, updatedChat.title, updatedChat.messages);
		}

		// update local state
		setChat({ ...updatedChat, messages: [...updatedChat.messages] });
	};

	const handleSelectConversation = useCallback(async (item: ConversationSearchItem) => {
		const selectedChat = await conversationStoreAPI.getConversation(item.id, item.title, true);
		if (selectedChat) {
			setChat(selectedChat);
			isChatPersistedRef.current = true;
			manualTitleLockedRef.current = false; // fresh load – don’t assume manual lock
			setEditingMessageId(null);
		}
	}, []);

	/* ----------------------------------------------------------------
	 * manual rename handler
	 * ----------------------------------------------------------------*/
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

	const getConversationForExport = useCallback(async (): Promise<string> => {
		const selectedChat = await conversationStoreAPI.getConversation(chat.id, chat.title, true);
		return JSON.stringify(selectedChat, null, 2);
	}, [chat.id, chat.title]);

	const updateStreamingMessage = useCallback(
		async (updatedChatWithUserMessage: Conversation, options: ChatOption) => {
			// Abort older request (if any) and reset UI
			abortRef.current?.abort();
			tokensReceivedRef.current = false;
			setIsBusy(true);

			abortRef.current = new AbortController();

			/* Generate a unique request-ID so that the Go side can cancel, too */
			requestIdRef.current = crypto.randomUUID();

			let prevMessages = updatedChatWithUserMessage.messages;
			if (options.disablePreviousMessages) {
				prevMessages = [updatedChatWithUserMessage.messages[updatedChatWithUserMessage.messages.length - 1]];
			}

			setStreamedMessage('');

			await new Promise(res => setTimeout(res, 0));

			const convoMsg = initConversationMessage(ConversationRoleEnum.assistant, '');
			const updatedChatWithConvoMessage = {
				...updatedChatWithUserMessage,
				messages: [...updatedChatWithUserMessage.messages, convoMsg],
				modifiedAt: new Date(),
			};

			// Make the empty assistant bubble appear immediately
			setChat({ ...updatedChatWithConvoMessage, messages: [...updatedChatWithConvoMessage.messages] });

			const onStreamTextData = (textData: string) => {
				if (textData) tokensReceivedRef.current = true;

				setStreamedMessage(prev => {
					const next = prev + textData;

					// Always copy the current assistant text into the conversation message object so that we do not lose it when the stream aborts.
					const lastIdx = updatedChatWithConvoMessage.messages.length - 1;
					updatedChatWithConvoMessage.messages[lastIdx].content = next;
					updatedChatWithConvoMessage.modifiedAt = new Date();

					// Also keep the local convoMsg (passed to HandleCompletion) in sync
					convoMsg.content = next;

					// A full re-render is only needed for the very first token.
					if (prev === '') setChat({ ...updatedChatWithConvoMessage });

					return next;
				});
			};

			const onStreamThinkingData = (thinkingData: string) => {
				if (!thinkingData) {
					return;
				}
				tokensReceivedRef.current = true;
				setStreamedMessage(prev => {
					const data = thinkingData ? getBlockQuotedLines(thinkingData) + '\n' : '';
					const next = prev + data;
					const last = updatedChatWithConvoMessage.messages.length - 1;
					updatedChatWithConvoMessage.messages[last].content = next;
					updatedChatWithConvoMessage.modifiedAt = new Date();
					// Also keep the local convoMsg (passed to HandleCompletion) in sync
					convoMsg.content = next;

					if (prev === '') setChat({ ...updatedChatWithConvoMessage });

					return next;
				});
			};

			try {
				const inputParams: ModelParams = {
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
				const completionData = await BuildCompletionDataFromConversation(
					options.providerName,
					inputParams,
					prevMessages
				);

				if (updatedChatWithConvoMessage.messages.length > 1) {
					const prevIdx = updatedChatWithConvoMessage.messages.length - 2;
					if (updatedChatWithConvoMessage.messages[prevIdx].role === ConversationRoleEnum.user) {
						if (completionData.messages && completionData.messages.length > 0) {
							const msg = completionData.messages[completionData.messages.length - 1];
							const atts = msg.attachments ?? [];
							if (atts.length > 0) {
								updatedChatWithConvoMessage.messages = updatedChatWithConvoMessage.messages.map((m, i) =>
									i === prevIdx ? { ...m, attachments: atts } : m
								);
								saveUpdatedChat({ ...updatedChatWithConvoMessage });
							}
						}
					}
				}
				const newMsg = await HandleCompletion(
					options.providerName,
					completionData,
					convoMsg,
					requestIdRef.current,
					abortRef.current.signal,
					onStreamTextData,
					onStreamThinkingData
				);

				// Attach debug details to the *previous* user message.
				// Prefer the provider-supplied HTTP request details (already
				// sanitized, includes curl). If those are not available (e.g.
				// the Go side failed before making the HTTP call), fall back to
				// the locally built completionData.
				if (updatedChatWithConvoMessage.messages.length > 1) {
					const prevIdx = updatedChatWithConvoMessage.messages.length - 2;
					const prevMessage = updatedChatWithConvoMessage.messages[prevIdx];

					if (prevMessage.role === ConversationRoleEnum.user) {
						let detailsMd: string | undefined;

						if (newMsg.requestDetails) {
							const reqData = 'API request data:\n' + newMsg.requestDetails;
							detailsMd = reqData;
						}

						updatedChatWithConvoMessage.messages = updatedChatWithConvoMessage.messages.map((m, i) =>
							i === prevIdx
								? {
										...m,
										// Overwrite details on resend; keep the most relevant
										// debug info visible.
										details: detailsMd,
									}
								: m
						);
					}
				}

				if (newMsg.responseMessage) {
					const respMessage = { ...newMsg.responseMessage };
					// Create FRESH objects so react sees the change even in non-streaming
					// mode, where `streamedMessage` never changes.
					const finalChat: Conversation = {
						...updatedChatWithConvoMessage,
						messages: [...updatedChatWithConvoMessage.messages.slice(0, -1), respMessage],
						modifiedAt: new Date(),
					};

					saveUpdatedChat(finalChat);
				}
			} catch (e) {
				if ((e as DOMException).name === 'AbortError') {
					// ESlint cannot see async updates in streaming, hence disable the lint warning.
					// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
					if (!tokensReceivedRef.current) {
						removeAssistantPlaceholder(convoMsg.id);
					} else {
						const last = updatedChatWithConvoMessage.messages.length - 1;

						updatedChatWithConvoMessage.messages[last].content += '\n\n>API Aborted after partial response...';
						updatedChatWithConvoMessage.modifiedAt = new Date();
						saveUpdatedChat({ ...updatedChatWithConvoMessage });
					}
				} else {
					console.error(e);
				}
			} finally {
				setStreamedMessage('');
				setIsBusy(false);
			}
		},
		[saveUpdatedChat]
	);

	const removeAssistantPlaceholder = (msgId: string) => {
		setChat(c => ({
			...c,
			messages: c.messages.filter(m => m.id !== msgId),
			modifiedAt: new Date(),
		}));
	};

	const sendMessage = async (payload: EditorSubmitPayload, options: ChatOption) => {
		if (isBusy) return;

		const trimmed = payload.text.trim();
		if (!trimmed) {
			return;
		}

		// If we are editing an existing user message, replace it and drop later messages.
		if (editingMessageId) {
			const idx = chat.messages.findIndex(m => m.id === editingMessageId);
			// If somehow the message vanished, fall back to "append new".
			if (idx !== -1) {
				const base = chat.messages[idx];

				const toolChoices =
					payload.attachedTools.length > 0
						? payload.attachedTools.map(attachedToolToConversationToolChoice)
						: undefined;

				const attachments =
					payload.attachments.length > 0 ? payload.attachments.map(editorAttachmentToConversation) : undefined;

				const updatedUserMsg: ConversationMessage = {
					...base,
					content: trimmed,
					toolChoices: toolChoices && toolChoices.length > 0 ? toolChoices : undefined,
					attachments: attachments && attachments.length > 0 ? attachments : undefined,
				};

				const msgs = chat.messages.slice(0, idx + 1);
				msgs[idx] = updatedUserMsg;

				const updatedChat: Conversation = {
					...chat,
					messages: msgs,
					modifiedAt: new Date(),
				};

				setEditingMessageId(null);
				saveUpdatedChat(updatedChat);
				await updateStreamingMessage(updatedChat, options);
				return;
			}

			// If not found, clear edit state and proceed as a normal send.
			setEditingMessageId(null);
		}

		const newMsg = initConversationMessage(ConversationRoleEnum.user, trimmed);
		const toolChoices =
			payload.attachedTools.length > 0 ? payload.attachedTools.map(attachedToolToConversationToolChoice) : undefined;
		if (toolChoices && toolChoices.length > 0) {
			newMsg.toolChoices = toolChoices;
		}
		const attachments =
			payload.attachments.length > 0 ? payload.attachments.map(editorAttachmentToConversation) : undefined;
		if (attachments && attachments.length > 0) {
			newMsg.attachments = attachments;
		}
		const updated = {
			...chat,
			messages: [...chat.messages, newMsg],
			modifiedAt: new Date(),
		};
		saveUpdatedChat(updated);
		updateStreamingMessage(updated, options);
	};

	const beginEditMessage = useCallback(
		(id: string) => {
			if (isBusy) return;

			const msg = chat.messages.find(m => m.id === id);
			if (!msg) return;
			if (msg.role !== ConversationRoleEnum.user) return;

			const external: EditorExternalMessage = {
				text: msg.content,
				attachments: msg.attachments,
				toolChoices: msg.toolChoices,
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

	// const handleEdit = useCallback(
	// 	async (edited: string, id: string) => {
	// 		if (isBusy) return;
	// 		const idx = chat.messages.findIndex(m => m.id === id);
	// 		if (idx === -1) return;

	// 		const msgs = chat.messages.slice(0, idx + 1);
	// 		msgs[idx] = { ...msgs[idx], content: edited };

	// 		const updated: Conversation = {
	// 			...chat,
	// 			messages: msgs,
	// 			modifiedAt: new Date(),
	// 		};
	// 		saveUpdatedChat(updated);

	// 		let opts = { ...DefaultChatOptions };
	// 		if (chatInputRef.current) opts = chatInputRef.current.getChatOptions();
	// 		await updateStreamingMessage(updated, opts);
	// 	},
	// 	[chat, saveUpdatedChat, updateStreamingMessage]
	// );

	const handleResend = useCallback(
		async (id: string) => {
			if (isBusy) return;
			const idx = chat.messages.findIndex(m => m.id === id);
			if (idx === -1) return;

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
			isBusy &&
			idx === chat.messages.length - 1 &&
			msg.role === ConversationRoleEnum.assistant &&
			msg.content.length === 0;
		const live =
			isBusy && idx === chat.messages.length - 1 && msg.role === ConversationRoleEnum.assistant ? streamedMessage : '';

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

					{/* Overlay the button; not part of the scrollable content */}
					<div className="pointer-events-none absolute right-4 bottom-4 z-10 xl:right-24">
						<div className="pointer-events-auto">
							<ButtonScrollToBottom
								scrollContainerRef={chatContainerRef}
								iconSize={32}
								show={isScrollable && !isAtBottom}
								className="btn btn-md border-none bg-transparent shadow-none"
							/>
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
