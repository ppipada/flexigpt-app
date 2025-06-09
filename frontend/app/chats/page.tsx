import type { FC } from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';

import { v7 as uuidv7 } from 'uuid';

import type { ModelParams } from '@/models/aiprovidermodel';
import type { Conversation, ConversationItem, ConversationMessage } from '@/models/conversationmodel';
import { ConversationRoleEnum } from '@/models/conversationmodel';
import { type ChatOptions, DefaultChatOptions } from '@/models/settingmodel';

import { GetCompletionMessage } from '@/apis/aiprovider_helper';
import { conversationStoreAPI } from '@/apis/baseapi';

import ButtonScrollToBottom from '@/components/button_scroll_to_bottom';

import ChatInputField, { type ChatInputFieldHandle } from '@/chats/chat_input_field';
import ChatMessage from '@/chats/chat_message';
import ChatNavBar from '@/chats/chat_navbar';

function deriveTitle(text: string) {
	const raw = text.substring(0, 48);
	return raw.charAt(0).toUpperCase() + raw.slice(1);
}

function initConversation(title = 'New Conversation'): Conversation {
	return {
		id: uuidv7(),
		title: deriveTitle(title),
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
	const [isStreaming, setIsStreaming] = useState(false);

	const chatContainerRef = useRef<HTMLDivElement>(null);
	const chatInputRef = useRef<ChatInputFieldHandle>(null);
	const isSubmittingRef = useRef(false);
	// Has the current conversation already been persisted?
	const isChatPersistedRef = useRef(false);

	// Focus on mount.
	useEffect(() => {
		chatInputRef.current?.focus();
	}, []);

	const bumpSearchKey = () => {
		setSearchRefreshKey(k => k + 1);
	};

	const handleNewChat = useCallback(async () => {
		if (chat.messages.length === 0) {
			chatInputRef.current?.focus();
			return;
		}
		conversationStoreAPI.saveConversation(chat);
		setChat(initConversation());
		// New non-persisted conversation started.
		isChatPersistedRef.current = false;
		chatInputRef.current?.focus();
	}, [chat]);

	const saveUpdatedChat = (updatedChat: Conversation) => {
		let titleChanged = false;
		if (updatedChat.messages.length === 1) {
			const newTitle = deriveTitle(updatedChat.messages[0].content);
			if (newTitle && newTitle !== updatedChat.title) {
				updatedChat.title = newTitle;
				titleChanged = true;
			}
		}

		conversationStoreAPI.saveConversation(updatedChat);
		setChat(updatedChat);

		// Detect "first save", now searchable.
		if (!isChatPersistedRef.current && updatedChat.messages.length > 0) {
			// Title is fresh; no extra checks needed.
			bumpSearchKey();
			isChatPersistedRef.current = true;
		} else if (titleChanged) {
			// Title change (e.g. first-message edit).
			bumpSearchKey();
		}
	};

	// Scroll to bottom.
	useEffect(() => {
		if (chatContainerRef.current) {
			chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight;
		}
	}, [chat.messages, streamedMessage]);

	const handleSelectConversation = useCallback(async (item: ConversationItem) => {
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
		async (updatedChatWithUserMessage: Conversation, options: ChatOptions) => {
			let prevMessages = updatedChatWithUserMessage.messages;
			if (options.disablePreviousMessages) {
				prevMessages = [updatedChatWithUserMessage.messages[updatedChatWithUserMessage.messages.length - 1]];
			}

			setIsStreaming(true);
			setStreamedMessage('');

			await new Promise(res => setTimeout(res, 0));

			const convoMsg = initConversationMessage(ConversationRoleEnum.assistant, '');
			const updatedChatWithConvoMessage = {
				...updatedChatWithUserMessage,
				messages: [...updatedChatWithUserMessage.messages, convoMsg],
				modifiedAt: new Date(),
			};

			const onStreamData = (data: string) => {
				setStreamedMessage(prev => {
					if (prev === '') {
						updatedChatWithConvoMessage.messages[updatedChatWithConvoMessage.messages.length - 1].content = data;
						updatedChatWithConvoMessage.modifiedAt = new Date();
						setChat(updatedChatWithConvoMessage);
						return data;
					}
					return prev + data;
				});
			};

			const inputParams: ModelParams = {
				name: options.name,
				temperature: options.temperature,
				stream: options.stream,
				maxPromptLength: options.maxPromptLength,
				maxOutputLength: options.maxOutputLength,
				reasoning: options.reasoning,
				systemPrompt: options.systemPrompt,
				timeout: options.timeout,
				additionalParameters: options.additionalParameters,
			};

			const newMsg = await GetCompletionMessage(options.provider, inputParams, convoMsg, prevMessages, onStreamData);

			if (newMsg.requestDetails) {
				if (updatedChatWithConvoMessage.messages.length > 1) {
					updatedChatWithConvoMessage.messages[updatedChatWithConvoMessage.messages.length - 2].details =
						newMsg.requestDetails;
				}
			}

			if (newMsg.responseMessage) {
				const respMessage = newMsg.responseMessage;
				updatedChatWithConvoMessage.messages.pop();
				updatedChatWithConvoMessage.messages.push(respMessage);
				updatedChatWithConvoMessage.modifiedAt = new Date();
				saveUpdatedChat(updatedChatWithConvoMessage);
				setIsStreaming(false);
			}

			isSubmittingRef.current = false;
		},
		[saveUpdatedChat]
	);

	const sendMessage = async (text: string, options: ChatOptions) => {
		if (isSubmittingRef.current) return;
		isSubmittingRef.current = true;

		const trimmed = text.trim();
		if (!trimmed) {
			isSubmittingRef.current = false;
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
			const idx = chat.messages.findIndex(m => m.id === id);
			if (idx === -1) return;

			const msgs = chat.messages.slice(0, idx + 1);
			msgs[idx].content = edited;

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
		const live =
			isStreaming && idx === chat.messages.length - 1 && msg.role === ConversationRoleEnum.assistant
				? streamedMessage
				: '';

		return (
			<ChatMessage
				key={msg.id}
				message={msg}
				onEdit={txt => handleEdit(txt, msg.id)}
				onResend={() => handleResend(msg.id)}
				streamedMessage={live}
			/>
		);
	});

	return (
		<div className="flex flex-col items-center w-full h-full overflow-hidden">
			{/* NAVBAR */}
			<div className="w-full flex justify-center fixed top-2 z-10">
				<div className="w-11/12 lg:w-4/5 xl:w-3/4">
					<ChatNavBar
						onNewChat={handleNewChat}
						getConversationForExport={getConversationForExport}
						onSelectConversation={handleSelectConversation}
						chatTitle={chat.title}
						searchRefreshKey={searchRefreshKey}
					/>
				</div>
			</div>

			{/* MESSAGES */}
			<div className="flex flex-col items-center w-full grow overflow-hidden mt-32">
				<div
					className="w-full grow flex justify-center overflow-y-auto"
					ref={chatContainerRef}
					style={{ maxHeight: `calc(100vh - 196px - ${inputHeight}px)` }}
				>
					<div className="w-11/12 lg:w-4/5 xl:w-3/4">
						<div className="w-full flex-1 space-y-4">{renderedMessages}</div>
					</div>
				</div>

				{/* INPUT */}
				<div className="w-full flex justify-center fixed bottom-0 mb-3">
					<div className="w-11/12 lg:w-4/5 xl:w-3/4">
						<ChatInputField ref={chatInputRef} onSend={sendMessage} setInputHeight={setInputHeight} />
					</div>
				</div>
			</div>

			{/* SCROLL-TO-BOTTOM BUTTON */}
			<div className="fixed bottom-0 right-0 mb-16 mr-0 lg:mr-16 z-10">
				<ButtonScrollToBottom
					scrollContainerRef={chatContainerRef}
					size={32}
					className="btn btn-md bg-transparent border-none flex items-center shadow-none"
				/>
			</div>
		</div>
	);
};

export default ChatScreen;
