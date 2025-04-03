import ChatMessageContent from '@/chats/chat_message_content';
import EditBox from '@/chats/chat_message_editbox';
import ChatMessageFooterArea from '@/chats/chat_message_footer';
import type { ConversationMessage } from '@/models/conversationmodel';
import { ConversationRoleEnum } from '@/models/conversationmodel';
import type { ChangeEvent, FC } from 'react';
import { useState } from 'react';
import { FiCompass, FiUser } from 'react-icons/fi';

interface ChatMessageProps {
	message: ConversationMessage;
	onEdit: (editedText: string) => void;
	onResend: () => void;
	streamedMessage: string;
}

const ChatMessage: FC<ChatMessageProps> = ({ message, onEdit, onResend, streamedMessage }) => {
	const isUser = message.role === ConversationRoleEnum.user;
	const align = !isUser ? 'items-end text-left' : 'items-start text-left';
	const leftColSpan = !isUser ? 'col-span-1 lg:col-span-2' : 'col-span-1';
	const rightColSpan = !isUser ? 'col-span-1' : 'col-span-1 lg:col-span-2';

	const [isEditing, setIsEditing] = useState(false);
	const [editText, setEditText] = useState(message.content);

	const handleEditClick = () => {
		setIsEditing(true);
	};

	const handleTextChange = (e: ChangeEvent<HTMLTextAreaElement>) => {
		setEditText(e.target.value);
	};

	const handleSubmit = (editedText: string) => {
		onEdit(editedText);
		setIsEditing(false);
	};

	const handleDiscard = () => {
		setIsEditing(false);
		setEditText(message.content);
	};

	return (
		<div className="grid grid-cols-12 grid-rows-[auto_auto] gap-2 mb-4" style={{ fontSize: '14px' }}>
			{/* Row 1: Icons and Message Content */}
			<div className={`${leftColSpan} flex justify-end row-start-1 row-end-1`}>
				{isUser && (
					<div className="flex w-10 h-10 items-center justify-center rounded-full self-end">
						<FiUser size={24} />
					</div>
				)}
			</div>
			<div className="col-span-10 lg:col-span-9 row-start-1 row-end-1 overflow-hidden shadow-lg rounded-2xl">
				{isEditing ? (
					<EditBox
						editText={editText}
						onTextChange={handleTextChange}
						onSubmit={handleSubmit}
						onDiscard={handleDiscard}
					/>
				) : (
					<div className="flex flex-col w-full">
						<ChatMessageContent
							content={streamedMessage || message.content}
							align={align}
							streamedMessage={streamedMessage}
							renderAsMarkdown={!isUser}
						/>
					</div>
				)}
			</div>
			<div className={`${rightColSpan} flex justify-start row-start-1 row-end-1`}>
				{!isUser && (
					<div className="flex w-10 h-10 items-center justify-center rounded-full self-end">
						<FiCompass size={24} />
					</div>
				)}
			</div>

			{/* Row 2: Footer */}
			<div className={`${leftColSpan} row-start-2 row-end-2`}></div>
			<div className="col-span-10 lg:col-span-9 row-start-2 row-end-2">
				{!isEditing && (
					<ChatMessageFooterArea
						isUser={isUser}
						cardContent={message.content}
						onEdit={handleEditClick}
						onResend={onResend}
						messageDetails={message.details ? message.details : ''}
						isStreaming={streamedMessage ? true : false}
					/>
				)}
			</div>
			<div className={`${rightColSpan} row-start-2 row-end-2`}></div>
		</div>
	);
};

export default ChatMessage;
