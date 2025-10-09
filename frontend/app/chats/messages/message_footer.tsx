import type { FC } from 'react';
import { useState } from 'react';

import { FiChevronDown, FiChevronUp, FiEdit2, FiRepeat } from 'react-icons/fi';

import { stripCustomMDFences } from '@/lib/text_utils';

import CopyButton from '@/components/copy_button';

import MessageContent from '@/chats/messages/message_content';

interface MessageFooterAreaProps {
	messageID: string;
	isUser: boolean;
	cardCopyContent: string;
	onEdit: () => void;
	onResend: () => void;
	messageDetails: string;
	isStreaming: boolean;
	isBusy: boolean;
	disableMarkdown: boolean;
	onDisableMarkdownChange: (checked: boolean) => void;
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
	disableMarkdown,
	onDisableMarkdownChange,
}) => {
	const [isExpanded, setIsExpanded] = useState(false);

	const toggleExpanded = () => {
		setIsExpanded(!isExpanded);
	};

	return (
		<div className="px-1 py-1 pt-0">
			<div className="flex h-8 items-center justify-between">
				<div className="flex">
					<button
						className={`btn btn-sm flex items-center border-none !bg-transparent shadow-none ${isBusy ? 'btn-disabled' : ''}`}
						onClick={toggleExpanded}
						aria-label="Details"
						title="Details"
						disabled={isBusy}
					>
						{isExpanded ? <FiChevronUp size={16} /> : <FiChevronDown size={16} />}
					</button>
					<label
						className={`ml-1 flex items-center space-x-2 truncate ${isBusy ? 'cursor-not-allowed opacity-50' : ''}`}
						title="Disable Markdown"
					>
						<input
							type="checkbox"
							checked={disableMarkdown}
							onChange={e => {
								onDisableMarkdownChange(e.target.checked);
							}}
							className="checkbox checkbox-xs ml-1 rounded-full"
							spellCheck="false"
							disabled={isBusy}
						/>
						<span className="text-neutral-custom text-xs text-nowrap">Disable Markdown</span>
					</label>
				</div>

				{isStreaming && (
					<div className="text-sm">
						<div className="flex items-center bg-transparent px-4 py-2">
							Streaming
							<span className="loading loading-dots loading-sm ml-4" />
						</div>
					</div>
				)}

				<div className="flex items-center gap-1">
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
					{isUser && (
						<button
							className={`btn btn-sm flex items-center border-none !bg-transparent shadow-none ${isBusy ? 'btn-disabled' : ''}`}
							onClick={onEdit}
							aria-label="Edit Message"
							title="Edit Message"
							disabled={isBusy}
						>
							<FiEdit2 size={16} />
						</button>
					)}

					<CopyButton
						value={stripCustomMDFences(cardCopyContent)}
						className={`btn btn-sm flex items-center border-none !bg-transparent shadow-none ${isBusy ? 'btn-disabled' : ''}`}
						size={16}
						disabled={isBusy}
					/>
				</div>
			</div>
			{isExpanded && messageDetails && (
				<div className="bg-base-100 mt-2 overflow-hidden rounded-2xl shadow-lg">
					<MessageContent
						messageID={messageID}
						content={messageDetails}
						streamedText=""
						isStreaming={false}
						isBusy={isBusy}
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
