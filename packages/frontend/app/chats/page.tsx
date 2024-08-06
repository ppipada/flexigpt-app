'use client';
import { getConversation, listAllConversations, saveConversation } from '@/api/conversation_memoized_api';
import { getCompletionMessage } from '@/chats/chat_helper';
import ChatInputField from '@/chats/chat_input_field';
import ChatMessage from '@/chats/chat_message';
import ChatNavBar from '@/chats/chat_navbar';
import ButtonScrollToBottom from '@/components/button_scroll_to_bottom';
import {
	Conversation,
	ConversationItem,
	ConversationRoleEnum,
	initConversation,
	initConversationMessage,
} from 'conversationmodel';
import { FC, createRef, useCallback, useEffect, useMemo, useRef, useState } from 'react';

const ChatScreen: FC = () => {
	const [chat, setChat] = useState<Conversation>(initConversation());
	const [initialItems, setInitialItems] = useState<ConversationItem[]>([]);
	const [inputHeight, setInputHeight] = useState(0);
	const chatContainerRef = useRef<HTMLDivElement>(null);
	const isSubmittingRef = useRef<boolean>(false);
	const [streamedMessage, setStreamedMessage] = useState<string>('');
	const [isStreaming, setIsStreaming] = useState<boolean>(false);

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

	const sendMessage = async (messageContent: string) => {
		if (isSubmittingRef.current) return;
		isSubmittingRef.current = true;
		const trimmedText = messageContent.trim();
		if (trimmedText) {
			const newMessage = initConversationMessage(ConversationRoleEnum.user, trimmedText);

			// First update: adding the user's message
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
			saveConversation(updatedChatWithUserMessage);
			setChat(updatedChatWithUserMessage);
			setIsStreaming(true);
			setStreamedMessage('');

			// Wait for the state to update
			await new Promise(resolve => setTimeout(resolve, 0));

			// Set a empty message for streaming data
			const convoMsg = initConversationMessage(ConversationRoleEnum.assistant, '');
			const updatedChatWithConvoMessage = {
				...updatedChatWithUserMessage,
				messages: [...updatedChatWithUserMessage.messages, convoMsg],
				modifiedAt: new Date(),
			};

			const onStreamData = (data: string) => {
				setStreamedMessage(prev => {
					// Add assistant message to chat only when first stream data arrives
					if (prev === '') {
						updatedChatWithConvoMessage.messages[updatedChatWithConvoMessage.messages.length - 1].content = data;
						updatedChatWithConvoMessage.modifiedAt = new Date();
						setChat(updatedChatWithConvoMessage);
						return data;
					}
					return prev + data;
				});
			};

			const newMsg = await getCompletionMessage(convoMsg, updatedChatWithUserMessage.messages, {}, onStreamData);

			if (newMsg) {
				// Remove last message and substitue the complete thing
				updatedChatWithConvoMessage.messages.pop();
				updatedChatWithConvoMessage.messages.push(newMsg);
				updatedChatWithConvoMessage.modifiedAt = new Date();
				// Save conversation after streaming complete
				saveConversation(updatedChatWithConvoMessage);
				// Set chat with final message
				setChat(updatedChatWithConvoMessage);
				setStreamedMessage(''); // Reset streamed message state
				setIsStreaming(false); // Mark streaming as complete
			}
		}
		isSubmittingRef.current = false;
	};

	const memoizedChatMessages = useMemo(
		() =>
			chat.messages.map((message, index) => (
				<ChatMessage
					key={message.id}
					message={message}
					onEdit={() => {}}
					onResend={() => {}}
					onLike={() => {}}
					onDislike={() => {}}
					onSendFeedback={() => {}}
					feedbackController={createRef<HTMLInputElement>()}
					streamedMessage={
						isStreaming && index === chat.messages.length - 1 && message.role === ConversationRoleEnum.assistant
							? streamedMessage
							: ''
					}
				/>
			)),
		[chat.messages, streamedMessage, isStreaming]
	);

	return (
		<div className="flex flex-col items-center w-full h-full overflow-hidden">
			<div className="w-full flex justify-center bg-transparent fixed top-2">
				<div className="w-10/12 lg:w-2/3">
					<ChatNavBar
						onNewChat={handleNewChat}
						getConversationForExport={getConversationForExport}
						initialSearchItems={initialItems}
						onSearch={fetchSearchResults}
						onSelectConversation={handleSelectConversation}
					/>
				</div>
			</div>
			<div className="flex flex-col items-center w-full flex-grow bg-transparent overflow-hidden mt-16">
				<h1 className="text-xl font-semibold text-center flex my-4">{chat.title}</h1>
				<div
					className="w-full flex-grow flex justify-center overflow-y-auto"
					ref={chatContainerRef}
					style={{ maxHeight: `calc(100vh - 156px - ${inputHeight}px)` }} // Adjust height dynamically
				>
					<div className="w-11/12 lg:w-2/3">
						<div className="w-full flex-1 space-y-4">{memoizedChatMessages}</div>
					</div>
				</div>
				<div className="w-full flex justify-center bg-transparent fixed bottom-0 mb-3">
					<div className="w-10/12 lg:w-2/3">
						<ChatInputField onSend={sendMessage} setInputHeight={setInputHeight} />
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
