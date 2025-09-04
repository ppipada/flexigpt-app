import type { ChangeEvent, FC } from 'react';
import { memo, useState } from 'react';

import { FiCompass, FiUser } from 'react-icons/fi';

import type { ConversationMessage } from '@/spec/conversation';
import { ConversationRoleEnum } from '@/spec/conversation';

import MessageContent from '@/chats/messages/message_content';
import MessageEditBox from '@/chats/messages/message_editbox';
import MessageFooterArea from '@/chats/messages/message_footer';

interface ChatMessageProps {
	message: ConversationMessage;
	onEdit: (editedText: string) => void;
	onResend: () => void;
	streamedMessage: string;
	isPending: boolean;
	isBusy: boolean;
}

const ChatMessageInner: FC<ChatMessageProps> = ({ message, onEdit, onResend, streamedMessage, isPending, isBusy }) => {
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
		<div className="mb-4 grid grid-cols-12 grid-rows-[auto_auto] gap-2" style={{ fontSize: 14 }}>
			{/* Row 1 ── icon + message bubble */}
			<div className={`${leftColSpan} row-start-1 row-end-1 flex justify-end`}>
				{isUser && (
					<div className="flex h-10 w-10 items-center justify-center self-end rounded-full">
						<FiUser size={24} />
					</div>
				)}
			</div>

			<div
				className={`bg-base-100 col-span-10 row-start-1 row-end-1 overflow-x-auto rounded-2xl lg:col-span-9 ${streamedMessage ? '' : 'shadow-lg'}`}
			>
				{isEditing ? (
					<MessageEditBox
						editText={editText}
						onTextChange={handleTextChange}
						onSubmit={handleSubmit}
						onDiscard={handleDiscard}
					/>
				) : (
					<MessageContent
						messageID={message.id}
						content={message.content}
						streamedText={streamedMessage}
						isStreaming={!!streamedMessage}
						isBusy={isBusy}
						isPending={isPending}
						align={align}
						renderAsMarkdown={!isUser}
					/>
				)}
			</div>

			<div className={`${rightColSpan} row-start-1 row-end-1 flex justify-start`}>
				{!isUser && (
					<div className="flex h-10 w-10 items-center justify-center self-end rounded-full">
						<FiCompass size={24} />
					</div>
				)}
			</div>

			{/* Row 2 ── footer bar */}
			<div className={`${leftColSpan} row-start-2 row-end-2`} />
			<div className="col-span-10 row-start-2 row-end-2 lg:col-span-9">
				{!isEditing && (
					<MessageFooterArea
						messageID={message.id}
						isUser={isUser}
						cardCopyContent={message.content}
						onEdit={handleEditClick}
						onResend={onResend}
						messageDetails={message.details ?? ''}
						isStreaming={!!streamedMessage}
						isBusy={isBusy}
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
	if (prev.isBusy !== next.isBusy) return false;

	// If the *object reference* for the ConversationMessage changes
	// React must re-render (content edited, message appended).
	if (prev.message !== next.message) return false;

	// Everything else is the same: skip.
	return true;
}

export default memo(ChatMessageInner, propsAreEqual);
