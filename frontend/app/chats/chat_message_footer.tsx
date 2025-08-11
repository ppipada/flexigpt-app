import type { FC } from 'react';
import { useState } from 'react';

import { FiChevronDown, FiChevronUp, FiEdit, FiRepeat } from 'react-icons/fi';

import CopyButton from '@/components/copy_button';

import ChatMessageContent from '@/chats/chat_message_content';

interface ChatMessageFooterAreaProps {
	isUser: boolean;
	cardContent: string;
	onEdit: () => void;
	onResend: () => void;
	messageDetails: string;
	isStreaming: boolean;
	isBusy: boolean;
}

const ChatMessageFooterArea: FC<ChatMessageFooterAreaProps> = ({
	isUser,
	cardContent,
	onEdit,
	onResend,
	messageDetails,
	isStreaming,
	isBusy,
}) => {
	const [isExpanded, setIsExpanded] = useState(false);

	const toggleExpanded = () => {
		setIsExpanded(!isExpanded);
	};

	return (
		<div className="p-1">
			<div className="flex justify-between items-center h-8">
				<div className="flex items-center gap-1">
					<CopyButton
						value={cardContent}
						className="btn btn-sm bg-transparent border-none flex items-center shadow-none"
						size={16}
					/>
					{isUser && (
						<button
							className={`btn btn-sm !bg-transparent border-none flex items-center shadow-none ${isBusy ? 'btn-disabled' : ''}`}
							onClick={onEdit}
							aria-label="Edit Message"
							title="Edit Message"
							disabled={isBusy}
						>
							<FiEdit size={16} />
						</button>
					)}
					{isUser && (
						<button
							className={`btn btn-sm !bg-transparent border-none flex items-center shadow-none ${isBusy ? 'btn-disabled' : ''}`}
							onClick={onResend}
							aria-label="Resend Message"
							title="Resend Message"
							disabled={isBusy}
						>
							<FiRepeat size={16} />
						</button>
					)}
				</div>
				{isStreaming && (
					<div className="text-sm">
						<div className="bg-transparent px-4 py-2 flex items-center">
							Streaming
							<span className="ml-4 loading loading-dots loading-sm" />
						</div>
					</div>
				)}
				<button
					className="btn btn-sm bg-transparent border-none flex items-center shadow-none"
					onClick={toggleExpanded}
					aria-label="Details"
					title="Details"
				>
					{isExpanded ? <FiChevronUp size={16} /> : <FiChevronDown size={16} />}
				</button>
			</div>
			{isExpanded && messageDetails && (
				<div className="bg-base-100 mt-2 overflow-hidden shadow-lg rounded-2xl">
					<ChatMessageContent
						content={messageDetails}
						streamedText=""
						isStreaming={false}
						isPending={false}
						align="items-start text-left"
						renderAsMarkdown={true}
					/>
				</div>
			)}
		</div>
	);
};

export default ChatMessageFooterArea;
