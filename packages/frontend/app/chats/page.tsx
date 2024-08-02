'use client';
import { getConversation, listConversations, saveConversation, startConversation } from '@/api/conversation';
import ChatInputField from '@/chats/ChatInputField';
import ChatMessage from '@/chats/ChatMessage';
import ChatNavBar from '@/chats/ChatNavbar';
import { SearchItem } from '@/chats/ChatSearch';
import ButtonScrollToBottom from '@/components/ButtonScrollToBottom';
import { ChatCompletionRoleEnum } from 'aiprovider';
import { Conversation, ConversationMessage } from 'conversationmodel';
import { FC, createRef, useEffect, useRef, useState } from 'react';
import { v7 as uuidv7 } from 'uuid';

const ChatScreen: FC = () => {
	const [chat, setChat] = useState<Conversation>({
		id: uuidv7(),
		title: 'New Conversation',
		createdAt: new Date(),
		modifiedAt: new Date(),
		messages: [],
	});

	const [initialItems, setInitialItems] = useState<SearchItem[]>([]);
	const [inputHeight, setInputHeight] = useState(0);
	const chatContainerRef = useRef<HTMLDivElement>(null);

	const loadInitialItems = async () => {
		const { conversations } = await listConversations();
		setInitialItems(conversations);
	};

	useEffect(() => {
		loadInitialItems();
	}, []);

	const sendMessage = async (messageContent: string) => {
		const trimmedText = messageContent.trim();
		if (trimmedText) {
			const newMessage: ConversationMessage = {
				id: new Date().toISOString(),
				role: ChatCompletionRoleEnum.user,
				content: trimmedText,
				userId: '1',
			};

			setChat(prevChat => {
				const updatedChat = {
					...prevChat,
					messages: [...prevChat.messages, newMessage],
					modifiedAt: new Date(),
				};
				saveConversation(updatedChat);
				return updatedChat;
			});
		}
	};

	const handleNewChat = async () => {
		const newChat = await startConversation('New chat');
		setChat(newChat);
	};

	const handleSelectConversation = async (item: SearchItem) => {
		const selectedChat = await getConversation(item.id, item.title);
		if (selectedChat) {
			setChat(selectedChat);
		}
	};

	useEffect(() => {
		if (chatContainerRef.current) {
			chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight;
		}
	}, [chat.messages]);

	const fetchSearchResults = async (query: string): Promise<SearchItem[]> => {
		const { conversations } = await listConversations();
		return conversations.filter(item => item.title.toLowerCase().includes(query.toLowerCase()));
	};

	const getConversationForExport = async (): Promise<string> => {
		const selectedChat = await getConversation(chat.id, chat.title);
		const value = JSON.stringify(selectedChat, null, 2);
		return value;
	};

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
			<div className="flex flex-col items-center w-full flex-grow bg-transparent overflow-hidden mt-12">
				<h1 className="text-xl font-semibold text-center flex my-4">{chat.title}</h1>
				<div
					className="w-full flex-grow flex justify-center overflow-y-auto"
					ref={chatContainerRef}
					style={{ maxHeight: `calc(100vh - 144px - ${inputHeight}px)` }} // Adjust height dynamically
				>
					<div className="w-11/12 lg:w-2/3">
						<div className="w-full flex-1 space-y-4">
							{chat.messages.map(message => (
								<ChatMessage
									key={message.id}
									user={{
										id: '1',
										role: message.role,
										name: 'Joe',
									}}
									message={message}
									onEdit={() => {}}
									onResend={() => {}}
									onLike={() => {}}
									onDislike={() => {}}
									onSendFeedback={() => {}}
									feedbackController={createRef<HTMLInputElement>()}
								/>
							))}
						</div>
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
