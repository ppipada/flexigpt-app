import EditBox from '@/chats/chat_message_editbox';
import StaticMessage from '@/chats/chat_message_fixedbox';
import { ConversationMessage, ConversationRoleEnum } from '@/models/conversationmodel';
import { ChangeEvent, FC, useState } from 'react';
import { FiCompass, FiUser } from 'react-icons/fi';

interface ChatMessageProps {
	message: ConversationMessage;
	onEdit: (editedText: string) => void;
	streamedMessage: string;
}

const ChatMessage: FC<ChatMessageProps> = ({ message, onEdit, streamedMessage }) => {
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
		<div className="grid grid-cols-12 gap-2 mb-4 items-start" style={{ fontSize: '14px' }}>
			<div className={`${leftColSpan} flex justify-end`}>
				{isUser && (
					<div className="flex w-10 h-10 mt-2 items-center justify-center rounded-full">
						<FiUser size={24} />
					</div>
				)}
			</div>
			<div className="col-span-10 lg:col-span-9">
				{isEditing ? (
					<EditBox
						editText={editText}
						onTextChange={handleTextChange}
						onSubmit={handleSubmit}
						onDiscard={handleDiscard}
					/>
				) : (
					<StaticMessage
						message={message}
						onEdit={handleEditClick}
						streamedMessage={streamedMessage}
						isUser={isUser}
						align={align}
					/>
				)}
			</div>
			<div className={`${rightColSpan} flex justify-start`}>
				{!isUser && (
					<div className="flex w-10 h-10 mt-2 items-center justify-center rounded-full">
						<FiCompass size={24} />
					</div>
				)}
			</div>
		</div>
	);
};

export default ChatMessage;
