import type { FC } from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';

import type { ModelParams } from '@/spec/aiprovider';
import type { Conversation, ConversationMessage, ConversationSearchItem } from '@/spec/conversation';
import { ConversationRoleEnum } from '@/spec/conversation';

import { generateTitle, getBlockQuotedLines } from '@/lib/text_utils';
import { getUUIDv7 } from '@/lib/uuid_utils';

import { useAtBottom } from '@/hooks/use_at_bottom';

import { GetCompletionMessage } from '@/apis/aiprovider_helper';
import { conversationStoreAPI } from '@/apis/baseapi';
import { type ChatOption, DefaultChatOptions } from '@/apis/chatoption_helper';

import ButtonScrollToBottom from '@/components/button_scroll_to_bottom';

import ChatInputField, { type ChatInputFieldHandle } from '@/chats/chat_input_field';
import ChatMessage from '@/chats/chat_message';
import ChatNavBar from '@/chats/chat_navbar';

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

const ChatScreen: FC = () => {
	const [chat, setChat] = useState<Conversation>(initConversation());
	const [searchRefreshKey, setSearchRefreshKey] = useState(0);
	const [inputHeight, setInputHeight] = useState(0);

	const [streamedMessage, setStreamedMessage] = useState('');
	const [isBusy, setIsBusy] = useState(false);

	/* Currently running request */
	const abortRef = useRef<AbortController | null>(null);
	const requestIdRef = useRef<string | null>(null); // will go to backend for cancellation

	const chatInputRef = useRef<ChatInputFieldHandle>(null);
	const chatContainerRef = useRef<HTMLDivElement>(null);
	const tokensReceivedRef = useRef<boolean | null>(false);

	// Has the current conversation already been persisted?
	const isChatPersistedRef = useRef(false);
	const { isAtBottom, isScrollable, checkScroll } = useAtBottom(chatContainerRef);

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
		chatInputRef.current?.focus();
	};

	// Persist `updatedChat` using the cheapest API that is still correct.
	// •  First time we ever write this conversation              -> putConversation
	// •  Title has changed (search index must be updated)        -> putConversation
	// •  Otherwise                                               -> putMessagesToConversation
	// putMessagesToConversation REQUIRES the *full* message list, so we just pass `updatedChat.messages` every time.
	const saveUpdatedChat = (updatedChat: Conversation) => {
		let newTitle = updatedChat.title;
		if (updatedChat.messages.length <= 4) {
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
		}

		const titleChanged = newTitle !== updatedChat.title;
		if (titleChanged) {
			updatedChat.title = newTitle;
		}

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

		// update local React state
		setChat({ ...updatedChat, messages: [...updatedChat.messages] });
	};

	const handleSelectConversation = useCallback(async (item: ConversationSearchItem) => {
		const selectedChat = await conversationStoreAPI.getConversation(item.id, item.title);
		if (selectedChat) {
			setChat(selectedChat);
			isChatPersistedRef.current = true;
		}
	}, []);

	const getConversationForExport = useCallback(async (): Promise<string> => {
		const selectedChat = await conversationStoreAPI.getConversation(chat.id, chat.title);
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

					/* Always copy the current assistant text into the conversation message
					       object so that we do not lose it when the stream aborts.                */
					const lastIdx = updatedChatWithConvoMessage.messages.length - 1;
					updatedChatWithConvoMessage.messages[lastIdx].content = next;
					updatedChatWithConvoMessage.modifiedAt = new Date();

					/* A full re-render is only needed for the very first token.               */
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
				const newMsg = await GetCompletionMessage(
					options.providerName,
					inputParams,
					convoMsg,
					prevMessages,
					requestIdRef.current,
					abortRef.current.signal,
					onStreamTextData,
					onStreamThinkingData
				);

				if (newMsg.requestDetails) {
					if (updatedChatWithConvoMessage.messages.length > 1) {
						const prevIdx = updatedChatWithConvoMessage.messages.length - 2;
						updatedChatWithConvoMessage.messages = updatedChatWithConvoMessage.messages.map((m, i) =>
							i === prevIdx ? { ...m, details: newMsg.requestDetails } : m
						);
					}
				}

				if (newMsg.responseMessage) {
					const respMessage = { ...newMsg.responseMessage };
					// Create FRESH objects so React sees the change even in non-streaming
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

	const sendMessage = async (text: string, options: ChatOption) => {
		if (isBusy) return;

		const trimmed = text.trim();
		if (!trimmed) {
			return;
		}

		const newMsg = initConversationMessage(ConversationRoleEnum.user, trimmed);
		const updated = {
			...chat,
			messages: [...chat.messages, newMsg],
			modifiedAt: new Date(),
		};
		saveUpdatedChat(updated);
		updateStreamingMessage(updated, options);
	};

	const handleEdit = useCallback(
		async (edited: string, id: string) => {
			if (isBusy) return;
			const idx = chat.messages.findIndex(m => m.id === id);
			if (idx === -1) return;

			const msgs = chat.messages.slice(0, idx + 1);
			msgs[idx] = { ...msgs[idx], content: edited };

			const updated: Conversation = {
				...chat,
				messages: msgs,
				modifiedAt: new Date(),
			};
			saveUpdatedChat(updated);

			let opts = { ...DefaultChatOptions };
			if (chatInputRef.current) opts = chatInputRef.current.getChatOptions();
			await updateStreamingMessage(updated, opts);
		},
		[chat, saveUpdatedChat, updateStreamingMessage]
	);

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
				onEdit={txt => handleEdit(txt, msg.id)}
				onResend={() => handleResend(msg.id)}
				streamedMessage={live}
				isPending={isPending}
				isBusy={isBusy}
			/>
		);
	});

	return (
		<div className="flex flex-col items-center w-full h-full overflow-hidden">
			{/* NAVBAR */}
			<div className="w-full flex justify-center fixed top-2">
				<div className="w-11/12 lg:w-5/6">
					<ChatNavBar
						onNewChat={handleNewChat}
						getConversationForExport={getConversationForExport}
						onSelectConversation={handleSelectConversation}
						chatTitle={chat.title}
						chatID={chat.id}
						searchRefreshKey={searchRefreshKey}
						disabled={isBusy}
					/>
				</div>
			</div>

			{/* MESSAGES */}
			<div className="flex flex-col items-center w-full grow overflow-hidden mt-32">
				<div
					className="w-full grow flex justify-center overflow-y-auto"
					ref={chatContainerRef}
					style={{ maxHeight: `calc(100vh - 208px - ${inputHeight}px)` }}
				>
					<div className="w-11/12 lg:w-5/6">
						<div className="w-full flex-1 space-y-4 mb-4">{renderedMessages}</div>
					</div>
				</div>
				{/* INPUT */}
				<div className="w-full flex justify-center fixed bottom-0 mb-3">
					<div className="w-11/12 lg:w-5/6">
						<ChatInputField
							ref={chatInputRef}
							onSend={sendMessage}
							setInputHeight={setInputHeight}
							isBusy={isBusy}
							abortRef={abortRef}
						/>
					</div>
				</div>
			</div>

			{/* SCROLL-TO-BOTTOM BUTTON */}
			<div className="fixed bottom-0 right-0 mb-28 mr-0 lg:mr-16">
				<ButtonScrollToBottom
					scrollContainerRef={chatContainerRef}
					iconSize={32}
					isAtBottom={isAtBottom}
					isScrollable={isScrollable}
					className="btn btn-md bg-transparent border-none flex items-center shadow-none"
				/>
			</div>
		</div>
	);
};

export default ChatScreen;
