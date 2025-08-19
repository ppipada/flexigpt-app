import type { FC } from 'react';
import { useState } from 'react';

import { FiChevronDown, FiChevronUp, FiEdit, FiRepeat } from 'react-icons/fi';

import { stripThinkingFences } from '@/lib/text_utils';

import CopyButton from '@/components/copy_button';

import MessageContent from '@/chats/messagebox/message_content';

interface MessageFooterAreaProps {
	messageID: string;
	isUser: boolean;
	cardCopyContent: string;
	onEdit: () => void;
	onResend: () => void;
	messageDetails: string;
	isStreaming: boolean;
	isBusy: boolean;
}

const MessageFooterArea: FC<MessageFooterAreaProps> = ({
	messageID,
	isUser,
	cardCopyContent,
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
		<div className="px-1 py-1 pt-0">
			<div className="flex h-8 items-center justify-between">
				{!isStreaming && (
					<div className="flex items-center gap-1">
						<CopyButton
							value={stripThinkingFences(cardCopyContent)}
							// value={cardCopyContent}
							className="btn btn-sm flex items-center border-none bg-transparent shadow-none"
							size={16}
						/>
						{isUser && (
							<button
								className={`btn btn-sm flex items-center border-none !bg-transparent shadow-none ${isBusy ? 'btn-disabled' : ''}`}
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
								className={`btn btn-sm flex items-center border-none !bg-transparent shadow-none ${isBusy ? 'btn-disabled' : ''}`}
								onClick={onResend}
								aria-label="Resend Message"
								title="Resend Message"
								disabled={isBusy}
							>
								<FiRepeat size={16} />
							</button>
						)}
					</div>
				)}
				{isStreaming && (
					<div className="text-sm">
						<div className="flex items-center bg-transparent px-4 py-2">
							Streaming
							<span className="loading loading-dots loading-sm ml-4" />
						</div>
					</div>
				)}
				{!isStreaming && (
					<button
						className="btn btn-sm flex items-center border-none bg-transparent shadow-none"
						onClick={toggleExpanded}
						aria-label="Details"
						title="Details"
					>
						{isExpanded ? <FiChevronUp size={16} /> : <FiChevronDown size={16} />}
					</button>
				)}
			</div>
			{isExpanded && messageDetails && (
				<div className="bg-base-100 mt-2 overflow-hidden rounded-2xl shadow-lg">
					<MessageContent
						messageID={messageID}
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

export default MessageFooterArea;
