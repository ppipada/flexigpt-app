import { ChatMessageContent } from '@/chats/chat_message_content';
import ChatMessageFooterArea from '@/chats/chat_message_footer';
import { ConversationMessage } from '@/models/conversationmodel';
import { FC } from 'react';

interface StaticMessageProps {
	message: ConversationMessage;
	onEdit: () => void;
	onResend: () => void;
	streamedMessage: string;
	isUser: boolean;
	align: string;
}

const StaticMessage: FC<StaticMessageProps> = ({ message, onEdit, onResend, streamedMessage, isUser, align }) => (
	<div className="flex flex-col w-full">
		<ChatMessageContent content={streamedMessage || message.content} align={align} streamedMessage={streamedMessage} />
		<ChatMessageFooterArea
			isUser={isUser}
			cardContent={message.content}
			onEdit={onEdit}
			onResend={onResend}
			messageDetails={message.details ? message.details : ''}
			isStreaming={streamedMessage ? true : false}
		/>
	</div>
);

export default StaticMessage;
