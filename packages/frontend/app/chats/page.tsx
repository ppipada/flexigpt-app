'use client';
import { getConversation, listAllConversations, saveConversation } from '@/api/conversation_memoized_api';
import ChatInputField from '@/chats/chat_input_field';
import ChatMessage from '@/chats/chat_message';
import ChatNavBar from '@/chats/chat_navbar';
import { SearchItem } from '@/chats/chat_search';
import ButtonScrollToBottom from '@/components/button_scroll_to_bottom';
import { ChatCompletionRequestMessage, ChatCompletionRoleEnum, providerSet } from 'aiprovider';

import {
	Conversation,
	ConversationMessage,
	ConversationRoleEnum,
	initConversation,
	initConversationMessage,
} from 'conversationmodel';
// import { log } from 'logger';
import { FC, createRef, useCallback, useEffect, useMemo, useRef, useState } from 'react';

const roleMap: Record<ConversationRoleEnum, ChatCompletionRoleEnum> = {
	[ConversationRoleEnum.system]: ChatCompletionRoleEnum.system,
	[ConversationRoleEnum.user]: ChatCompletionRoleEnum.user,
	[ConversationRoleEnum.assistant]: ChatCompletionRoleEnum.assistant,
	[ConversationRoleEnum.function]: ChatCompletionRoleEnum.function,
	[ConversationRoleEnum.feedback]: ChatCompletionRoleEnum.user,
};

function convertConversationToChatMessages(
	conversationMessages?: ConversationMessage[]
): ChatCompletionRequestMessage[] {
	if (!conversationMessages) {
		return [];
	}
	const chatMessages: ChatCompletionRequestMessage[] = [];
	conversationMessages.forEach(convoMsg => {
		chatMessages.push({ role: roleMap[convoMsg.role], content: convoMsg.content });
	});
	return chatMessages;
}

async function getCompletionMessage(
	prompt: string,
	messages?: Array<ConversationMessage>,
	inputParams?: { [key: string]: any }
): Promise<ConversationMessage | undefined> {
	const chatMsgs = convertConversationToChatMessages(messages);
	const providerResp = await providerSet
		.getProviderAPI(providerSet.getDefaultProvider())
		.getCompletion(prompt, chatMsgs, inputParams);
	let respContent = '';
	let respDetails: string | undefined;
	if (providerResp) {
		if (providerResp.respContent) {
			respContent = providerResp.respContent;
		}
		if (providerResp.fullResponse) {
			respDetails = '```json' + JSON.stringify(providerResp.fullResponse, null, 2) + '```';
		}
	}
	const newMsg = initConversationMessage(ConversationRoleEnum.assistant, respContent);
	newMsg.details = respDetails;
	return newMsg;
}

const ChatScreen: FC = () => {
	const [chat, setChat] = useState<Conversation>(initConversation());
	const [initialItems, setInitialItems] = useState<SearchItem[]>([]);
	const [inputHeight, setInputHeight] = useState(0);
	const chatContainerRef = useRef<HTMLDivElement>(null);
	const isSubmittingRef = useRef<boolean>(false);

	const loadInitialItems = useCallback(async () => {
		const conversations = await listAllConversations();
		setInitialItems(conversations);
	}, []);

	useEffect(() => {
		loadInitialItems();
	}, [loadInitialItems]);

	const sendMessage = useCallback(
		async (messageContent: string) => {
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
					updatedChatWithUserMessage.title = `${updatedChatWithUserMessage.messages[0].content.substring(0, 32)}...`;
				}
				saveConversation(updatedChatWithUserMessage);
				setChat(updatedChatWithUserMessage);

				// Wait for the state to update
				await new Promise(resolve => setTimeout(resolve, 0));

				const newMsg = await getCompletionMessage(trimmedText, updatedChatWithUserMessage.messages);
				if (newMsg) {
					// Second update: adding the assistant's message
					const updatedChatWithAssistantMessage = {
						...updatedChatWithUserMessage,
						messages: [...updatedChatWithUserMessage.messages, newMsg],
						modifiedAt: new Date(),
					};
					saveConversation(updatedChatWithAssistantMessage);
					setChat(updatedChatWithAssistantMessage);
				}
			}
			isSubmittingRef.current = false;
		},
		[chat]
	);

	const handleNewChat = useCallback(async () => {
		saveConversation(chat);
		setChat(initConversation());
	}, [chat]);

	const handleSelectConversation = useCallback(async (item: SearchItem) => {
		const selectedChat = await getConversation(item.id, item.title);
		if (selectedChat) {
			setChat(selectedChat);
		}
	}, []);

	useEffect(() => {
		if (chatContainerRef.current) {
			chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight;
		}
	}, [chat.messages]);

	const fetchSearchResults = useCallback(async (query: string): Promise<SearchItem[]> => {
		const conversations = await listAllConversations();
		return conversations.filter(item => item.title.toLowerCase().includes(query.toLowerCase()));
	}, []);

	const getConversationForExport = useCallback(async (): Promise<string> => {
		const selectedChat = await getConversation(chat.id, chat.title);
		const value = JSON.stringify(selectedChat, null, 2);
		return value;
	}, [chat.id, chat.title]);

	const memoizedChatMessages = useMemo(
		() =>
			chat.messages.map(message => (
				<ChatMessage
					key={message.id}
					message={message}
					onEdit={() => {}}
					onResend={() => {}}
					onLike={() => {}}
					onDislike={() => {}}
					onSendFeedback={() => {}}
					feedbackController={createRef<HTMLInputElement>()}
				/>
			)),
		[chat.messages]
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
