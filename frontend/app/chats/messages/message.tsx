import { memo, useState } from 'react';

import { FiCompass, FiUser } from 'react-icons/fi';

import type { ConversationMessage } from '@/spec/conversation';
import { RoleEnum } from '@/spec/modelpreset';

import { MessageAttachmentsBar } from '@/chats/messages/message_attachments_bar';
import { MessageContent } from '@/chats/messages/message_content';
import { MessageFooterArea } from '@/chats/messages/message_footer';

interface ChatMessageProps {
	message: ConversationMessage;
	onEdit: () => void;
	onResend: () => void;
	streamedMessage: string;
	isPending: boolean;
	isBusy: boolean;
	isEditing: boolean;
}

function propsAreEqual(prev: ChatMessageProps, next: ChatMessageProps) {
	if (prev.message.details !== next.message.details) {
		//
		// We need to check details as parent is updating details in place for previous message
		return false;
	}
	if (prev.message.usage !== next.message.usage) {
		return false;
	}
	// We only care if THIS row’s streamed text changed.
	if (prev.streamedMessage !== next.streamedMessage) return false;

	if (prev.isPending !== next.isPending) return false;
	if (prev.isBusy !== next.isBusy) return false;
	if (prev.isEditing !== next.isEditing) return false;

	// If the *object reference* for the ConversationMessage changes
	// react must re-render (content edited, message appended).
	if (prev.message !== next.message) return false;

	// Everything else is the same: skip.
	return true;
}

export const ChatMessage = memo(function ChatMessage({
	message,
	onEdit,
	onResend,
	streamedMessage,
	isPending,
	isBusy,
	isEditing,
}: ChatMessageProps) {
	const isUser = message.role === RoleEnum.User;
	const align = !isUser ? 'items-end text-left' : 'items-start text-left';
	const leftColSpan = !isUser ? 'col-span-1 lg:col-span-2' : 'col-span-1';
	const rightColSpan = !isUser ? 'col-span-1' : 'col-span-1 lg:col-span-2';

	const [renderMarkdown, setRenderMarkdown] = useState(!isUser);

	const bubbleBase = 'bg-base-100 col-span-10 row-start-1 row-end-1 overflow-x-auto rounded-2xl lg:col-span-9';
	const bubbleExtra = [streamedMessage ? '' : 'shadow-lg', isEditing ? 'ring-2 ring-primary/70' : '']
		.filter(Boolean)
		.join(' ');

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

			<div className={`${bubbleBase} ${bubbleExtra}`}>
				<MessageContent
					messageID={message.id}
					content={message.content}
					streamedText={streamedMessage}
					isStreaming={!!streamedMessage}
					isBusy={isBusy}
					isPending={isPending}
					align={align}
					renderAsMarkdown={renderMarkdown}
				/>
				<div className="flex w-full min-w-0 overflow-x-hidden p-0">
					<MessageAttachmentsBar attachments={message.attachments} toolChoices={message.toolChoices} />
				</div>
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
				<MessageFooterArea
					messageID={message.id}
					isUser={isUser}
					cardCopyContent={message.content}
					onEdit={onEdit}
					onResend={onResend}
					messageDetails={message.details ?? ''}
					isStreaming={!!streamedMessage}
					isBusy={isBusy}
					disableMarkdown={!renderMarkdown}
					onDisableMarkdownChange={checked => {
						setRenderMarkdown(!checked);
					}}
					usage={message.usage}
				/>
			</div>
			<div className={`${rightColSpan} row-start-2 row-end-2`} />
		</div>
	);
}, propsAreEqual);
