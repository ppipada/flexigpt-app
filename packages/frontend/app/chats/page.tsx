'use client';
import { getCompletionMessage } from '@/backendapihelper/chat_helper';
import { getConversation, listAllConversations, saveConversation } from '@/backendapihelper/conversation_memoized_api';
import ChatInputField, { ChatInputFieldHandle, ChatOptions } from '@/chats/chat_input_field';
import ChatMessage from '@/chats/chat_message';
import ChatNavBar from '@/chats/chat_navbar';
import ButtonScrollToBottom from '@/components/button_scroll_to_bottom';
import { Conversation, ConversationItem, ConversationMessage, ConversationRoleEnum } from '@/models/conversationmodel';
import { FC, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { v7 as uuidv7 } from 'uuid';

function initConversation(title: string = 'New Conversation'): Conversation {
	return {
		id: uuidv7(),
		title: title.substring(0, 64),
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
		role: role,
		content: content,
	};
}

const ChatScreen: FC = () => {
	const [chat, setChat] = useState<Conversation>(initConversation());
	const [initialItems, setInitialItems] = useState<ConversationItem[]>([]);
	const [inputHeight, setInputHeight] = useState(0);
	const chatContainerRef = useRef<HTMLDivElement>(null);
	const isSubmittingRef = useRef<boolean>(false);
	const [streamedMessage, setStreamedMessage] = useState<string>('');
	const [isStreaming, setIsStreaming] = useState<boolean>(false);
	const chatInputRef = useRef<ChatInputFieldHandle>(null);

	const loadInitialItems = useCallback(async () => {
		const conversations = await listAllConversations();
		setInitialItems(conversations);
	}, []);

	useEffect(() => {
		loadInitialItems();
	}, [loadInitialItems]);

	const handleNewChat = useCallback(async () => {
		saveConversation(chat);
		setChat(initConversation());
	}, [chat]);

	const handleSelectConversation = useCallback(async (item: ConversationItem) => {
		const selectedChat = await getConversation(item.id, item.title);
		if (selectedChat) {
			setChat(selectedChat);
		}
	}, []);

	useEffect(() => {
		if (chatContainerRef.current) {
			chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight;
		}
	}, [chat.messages, streamedMessage]);

	const fetchSearchResults = useCallback(async (query: string): Promise<ConversationItem[]> => {
		const conversations = await listAllConversations();
		return conversations.filter(item => item.title.toLowerCase().includes(query.toLowerCase()));
	}, []);

	const getConversationForExport = useCallback(async (): Promise<string> => {
		const selectedChat = await getConversation(chat.id, chat.title);
		const value = JSON.stringify(selectedChat, null, 2);
		return value;
	}, [chat.id, chat.title]);

	const saveUpdatedChat = (updatedChat: Conversation) => {
		saveConversation(updatedChat);
		setChat(updatedChat);
	};

	const updateStreamingMessage = useCallback(async (updatedChatWithUserMessage: Conversation, options: ChatOptions) => {
		const inputParams: Record<string, any> = {
			provider: options.modelInfo?.provider,
			model: options.modelInfo?.name,
			temperature: options.modelInfo?.temperature,
		};
		let prevMessages = updatedChatWithUserMessage.messages;
		if (options.disablePreviousMessages) {
			prevMessages = [updatedChatWithUserMessage.messages[updatedChatWithUserMessage.messages.length - 1]];
		}

		setIsStreaming(true);
		setStreamedMessage('');

		await new Promise(resolve => setTimeout(resolve, 0));

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
		// log.info(JSON.stringify({ prevMessages, inputParams, convoMsg }, null, 2));
		const newMsg = await getCompletionMessage(convoMsg, prevMessages, inputParams, onStreamData);
		if (newMsg && newMsg.requestDetails) {
			if (updatedChatWithConvoMessage.messages.length > 1) {
				updatedChatWithConvoMessage.messages[updatedChatWithConvoMessage.messages.length - 2].details =
					newMsg.requestDetails;
			}
		}
		if (newMsg && newMsg.responseMessage) {
			const respMessage = newMsg.responseMessage;
			updatedChatWithConvoMessage.messages.pop();
			updatedChatWithConvoMessage.messages.push(respMessage);
			updatedChatWithConvoMessage.modifiedAt = new Date();
			saveUpdatedChat(updatedChatWithConvoMessage);
			setIsStreaming(false); // Mark streaming as complete
		}
		isSubmittingRef.current = false; // Reset submitting flag
	}, []);

	const sendMessage = async (messageContent: string, options: ChatOptions) => {
		if (isSubmittingRef.current) return;
		isSubmittingRef.current = true;

		const trimmedText = messageContent.trim();
		if (!trimmedText) {
			isSubmittingRef.current = false;
			return;
		}
		const newMessage = initConversationMessage(ConversationRoleEnum.user, trimmedText);

		const updatedChatWithUserMessage = {
			...chat,
			messages: [...chat.messages, newMessage],
			modifiedAt: new Date(),
		};
		if (updatedChatWithUserMessage.messages.length === 1) {
			const content = updatedChatWithUserMessage.messages[0].content.substring(0, 48);
			const capitalizedTitle = content.charAt(0).toUpperCase() + content.slice(1);
			updatedChatWithUserMessage.title = capitalizedTitle;
		}
		saveUpdatedChat(updatedChatWithUserMessage);

		updateStreamingMessage(updatedChatWithUserMessage, options);
	};

	const handleEdit = useCallback(
		async (editedText: string, messageId: string) => {
			const messageIndex = chat.messages.findIndex(msg => msg.id === messageId);
			if (messageIndex === -1) return;

			const newMessages = chat.messages.slice(0, messageIndex + 1);
			newMessages[messageIndex].content = editedText;

			const updatedChat = {
				...chat,
				messages: newMessages,
				modifiedAt: new Date(),
			};
			saveUpdatedChat(updatedChat);
			let currentChatOptions: ChatOptions = {
				disablePreviousMessages: false,
			};
			if (chatInputRef && chatInputRef.current) {
				currentChatOptions = chatInputRef.current.getChatOptions();
			}

			await updateStreamingMessage(updatedChat, currentChatOptions);
		},
		[chat, updateStreamingMessage]
	);

	const memoizedChatMessages = useMemo(
		() =>
			chat.messages.map((message, index) => (
				<ChatMessage
					key={message.id}
					message={message}
					onEdit={editedText => handleEdit(editedText, message.id)}
					streamedMessage={
						isStreaming && index === chat.messages.length - 1 && message.role === ConversationRoleEnum.assistant
							? streamedMessage
							: ''
					}
				/>
			)),
		[chat.messages, streamedMessage, isStreaming, handleEdit]
	);

	return (
		<div className="flex flex-col items-center w-full h-full overflow-hidden">
			<div className="w-full flex justify-center bg-transparent fixed top-2">
				<div className="w-11/12 lg:w-4/5 xl:w-3/4">
					<ChatNavBar
						onNewChat={handleNewChat}
						getConversationForExport={getConversationForExport}
						initialSearchItems={initialItems}
						onSearch={fetchSearchResults}
						onSelectConversation={handleSelectConversation}
						chatTitle={chat.title}
					/>
				</div>
			</div>
			<div className="flex flex-col items-center w-full flex-grow bg-transparent overflow-hidden mt-32">
				<div
					className="w-full flex-grow flex justify-center overflow-y-auto"
					ref={chatContainerRef}
					style={{ maxHeight: `calc(100vh - 196px - ${inputHeight}px)` }} // Adjust height dynamically
				>
					<div className="w-11/12 lg:w-4/5 xl:w-3/4">
						<div className="w-full flex-1 space-y-4">{memoizedChatMessages}</div>
					</div>
				</div>
				<div className="w-full flex justify-center bg-transparent fixed bottom-0 mb-3">
					<div className="w-11/12 lg:w-4/5 xl:w-3/4">
						<ChatInputField ref={chatInputRef} onSend={sendMessage} setInputHeight={setInputHeight} />
					</div>
				</div>
			</div>
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
