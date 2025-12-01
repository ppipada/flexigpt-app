import { useState } from 'react';

import { FiCode, FiEdit2, FiRepeat } from 'react-icons/fi';

import { stripCustomMDFences } from '@/lib/text_utils';

import { CopyButton } from '@/components/copy_button';

import { MessageDetailsModal } from '@/chats/messages/message_details_modal';

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

export function MessageFooterArea({
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
}: MessageFooterAreaProps) {
	const [isDetailsOpen, setIsDetailsOpen] = useState(false);

	const hasDetails = !!messageDetails;

	const toggleDetailsModal = () => {
		if (!hasDetails || isBusy) return;
		setIsDetailsOpen(prev => !prev);
	};

	return (
		<>
			<div className="px-1 py-1 pt-0">
				<div className="flex h-8 items-center justify-between">
					<div className="flex">
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
							<span className="text-base-content text-xs text-nowrap">Disable Markdown</span>
						</label>
						<button
							className={`btn text-neutral-custom flex items-center gap-2 border-none bg-transparent! text-xs shadow-none ${
								isBusy || !hasDetails ? 'btn-disabled' : ''
							}`}
							onClick={toggleDetailsModal}
							aria-label="Details"
							title={hasDetails ? 'Show details' : 'No details'}
							disabled={isBusy || !hasDetails}
						>
							<FiCode size={16} />
						</button>
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
								className={`btn btn-sm flex items-center border-none bg-transparent! shadow-none ${
									isBusy ? 'btn-disabled' : ''
								}`}
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
								className={`btn btn-sm flex items-center border-none bg-transparent! shadow-none ${
									isBusy ? 'btn-disabled' : ''
								}`}
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
							className={`btn btn-sm flex items-center border-none bg-transparent! shadow-none ${
								isBusy ? 'btn-disabled' : ''
							}`}
							size={16}
							disabled={isBusy}
						/>
					</div>
				</div>
			</div>

			{/* Details modal (works for both user & assistant messages) */}
			<MessageDetailsModal
				isOpen={isDetailsOpen && hasDetails}
				onClose={() => {
					setIsDetailsOpen(false);
				}}
				messageID={messageID}
				title={isUser ? 'User message details' : 'Assistant message details'}
				content={messageDetails}
				isBusy={isBusy}
			/>
		</>
	);
}
