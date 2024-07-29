'use client';
import ChatInputField from '@/chats/ChatInputField';
import ChatMessage from '@/chats/ChatMessage';
import ChatNavBar from '@/chats/ChatNavbar';
import ButtonScrollToBottom from '@/components/ButtonScrollToBottom';
import { messageSamplesList } from '@/lib/MessageSamples';
import { Chat, ChatCompletionRoleEnum, Message } from '@/lib/models/ChatTypes';
import { FC, createRef, useEffect, useRef, useState } from 'react';
import { v7 as uuidv7 } from 'uuid';

const initialItems = ['Chat 1', 'Chat 2', 'Chat 3', 'Chat 4', 'my convo 1', 'py search 1'];

const fetchSearchResults = async (query: string): Promise<string[]> => {
	// Replace with your actual search logic
	return initialItems.filter(item => item.toLowerCase().includes(query.toLowerCase()));
};

const ChatScreen: FC = () => {
	const [chat, setChat] = useState<Chat>({
		id: uuidv7(),
		title: 'New Chat',
		createTime: new Date(),
		modifiedTime: new Date(),
		messages: [],
	});

	const [inputHeight, setInputHeight] = useState(0);
	const chatContainerRef = useRef<HTMLDivElement>(null);

	const sendMessage = (messageContent: string) => {
		const trimmedText = messageContent.trim();
		if (trimmedText) {
			const newMessage: Message = {
				id: new Date().toISOString(),
				role: ChatCompletionRoleEnum.user,
				content: trimmedText,
			};

			setChat(prevChat => ({
				...prevChat,
				messages: [...prevChat.messages, newMessage],
				modifiedTime: new Date(),
			}));
		}
	};

	useEffect(() => {
		// Initialize chat with sample messages on mount
		setChat(prevChat => ({
			...prevChat,
			messages: messageSamplesList,
			modifiedTime: new Date(),
		}));
	}, []);

	useEffect(() => {
		if (chatContainerRef.current) {
			chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight;
		}
	}, [chat.messages]);

	return (
		<div className="flex flex-col items-center w-full h-full overflow-hidden">
			<div className="w-full flex justify-center bg-transparent fixed top-2">
				<div className="w-10/12 lg:w-2/3">
					<ChatNavBar
						onNewChat={() => {
							setChat(prevChat => ({
								...prevChat,
								messages: [],
								modifiedTime: new Date(),
							}));
						}}
						onExport={() => {
							// Add export functionality
						}}
						initialSearchItems={initialItems}
						onSearch={fetchSearchResults}
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
										icon: (
											<div className="w-8 h-8 flex items-center justify-center bg-base-100 rounded-full">
												{message.role === ChatCompletionRoleEnum.user ? 'U' : 'A'}
											</div>
										),
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
