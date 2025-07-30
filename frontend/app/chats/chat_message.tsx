import type { ChangeEvent, FC } from 'react';
import { memo, useState } from 'react';

import { FiCompass, FiUser } from 'react-icons/fi';

import type { ConversationMessage } from '@/spec/conversation';
import { ConversationRoleEnum } from '@/spec/conversation';

import ChatMessageContent from '@/chats/chat_message_content';
import EditBox from '@/chats/chat_message_editbox';
import ChatMessageFooterArea from '@/chats/chat_message_footer';

interface ChatMessageProps {
	message: ConversationMessage;
	onEdit: (editedText: string) => void;
	onResend: () => void;
	streamedMessage: string;
	isPending: boolean;
}

const ChatMessageInner: FC<ChatMessageProps> = ({ message, onEdit, onResend, streamedMessage, isPending }) => {
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
	const handleSubmit = (txt: string) => {
		onEdit(txt);
		setIsEditing(false);
	};
	const handleDiscard = () => {
		setIsEditing(false);
		setEditText(message.content);
	};

	return (
		<div className="grid grid-cols-12 grid-rows-[auto_auto] gap-2 mb-4" style={{ fontSize: 14 }}>
			{/* Row 1 ── icon + message bubble */}
			<div className={`${leftColSpan} flex justify-end row-start-1 row-end-1`}>
				{isUser && (
					<div className="flex w-10 h-10 items-center justify-center rounded-full self-end">
						<FiUser size={24} />
					</div>
				)}
			</div>

			<div
				className={`col-span-10 lg:col-span-9 row-start-1 row-end-1 overflow-hidden rounded-2xl ${streamedMessage ? '' : 'shadow-lg'}`}
			>
				{isEditing ? (
					<EditBox
						editText={editText}
						onTextChange={handleTextChange}
						onSubmit={handleSubmit}
						onDiscard={handleDiscard}
					/>
				) : (
					<ChatMessageContent
						content={message.content}
						streamedText={streamedMessage}
						isStreaming={!!streamedMessage}
						isPending={isPending}
						align={align}
						renderAsMarkdown={!isUser}
					/>
				)}
			</div>

			<div className={`${rightColSpan} flex justify-start row-start-1 row-end-1`}>
				{!isUser && (
					<div className="flex w-10 h-10 items-center justify-center rounded-full self-end">
						<FiCompass size={24} />
					</div>
				)}
			</div>

			{/* Row 2 ── footer bar */}
			<div className={`${leftColSpan} row-start-2 row-end-2`} />
			<div className="col-span-10 lg:col-span-9 row-start-2 row-end-2">
				{!isEditing && (
					<ChatMessageFooterArea
						isUser={isUser}
						cardContent={message.content}
						onEdit={handleEditClick}
						onResend={onResend}
						messageDetails={message.details ?? ''}
						isStreaming={!!streamedMessage}
					/>
				)}
			</div>
			<div className={`${rightColSpan} row-start-2 row-end-2`} />
		</div>
	);
};

function propsAreEqual(prev: ChatMessageProps, next: ChatMessageProps) {
	if (prev.message.details !== next.message.details) {
		//
		// We need to check details as parent is updating details in place for previous message
		return false;
	}
	// We only care if THIS row’s streamed text changed.
	if (prev.streamedMessage !== next.streamedMessage) return false;

	if (prev.isPending !== next.isPending) return false;

	// If the *object reference* for the ConversationMessage changes
	// React must re-render (content edited, message appended).
	if (prev.message !== next.message) return false;

	// Everything else is the same: skip.
	return true;
}

export default memo(ChatMessageInner, propsAreEqual);
