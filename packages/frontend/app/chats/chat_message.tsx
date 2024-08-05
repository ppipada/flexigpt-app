import { ChatMessageContent } from '@/chats/chat_message_content';
import ChatMessageFooterArea from '@/chats/chat_message_footer';
import { ConversationMessage, ConversationRoleEnum } from 'conversationmodel';
import { FC, RefObject } from 'react';
import { FiCompass, FiUser } from 'react-icons/fi';

interface ChatMessageProps {
	message: ConversationMessage;
	onEdit: () => void;
	onResend: () => void;
	onLike: () => void;
	onDislike: () => void;
	onSendFeedback: () => void;
	feedbackController: RefObject<HTMLInputElement>;
}

const ChatMessage: FC<ChatMessageProps> = ({
	message,
	onEdit,
	onResend,
	onLike,
	onDislike,
	onSendFeedback,
	feedbackController,
}) => {
	const isUser = message.role === ConversationRoleEnum.user;
	const align = !isUser ? 'items-end text-left' : 'items-start text-left';
	const leftColSpan = !isUser ? 'col-span-1 lg:col-span-2' : 'col-span-1';
	const rightColSpan = !isUser ? 'col-span-1' : 'col-span-1 lg:col-span-2';

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
				<div className="flex flex-col w-full">
					<ChatMessageContent content={message.content} align={align} />
					<ChatMessageFooterArea
						isUser={isUser}
						onEdit={onEdit}
						onResend={onResend}
						onLike={onLike}
						onDislike={onDislike}
						onSendFeedback={onSendFeedback}
						feedbackController={feedbackController}
						messageDetails={message.details ? message.details : ''}
						cardContent={message.content}
					/>
				</div>
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
