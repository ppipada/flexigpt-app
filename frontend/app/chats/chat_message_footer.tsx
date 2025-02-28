import { ChatMessageContent } from '@/chats/chat_message_content';
import CopyButton from '@/components/copy_button';
import { FC, useState } from 'react';
import { FiChevronDown, FiChevronUp, FiEdit, FiRepeat } from 'react-icons/fi';

interface ChatMessageFooterAreaProps {
	isUser: boolean;
	cardContent: string;
	onEdit: () => void;
	onResend: () => void;
	messageDetails: string;
	isStreaming: boolean;
}

const ChatMessageFooterArea: FC<ChatMessageFooterAreaProps> = ({
	isUser,
	cardContent,
	onEdit,
	onResend,
	messageDetails,
	isStreaming,
}) => {
	const [isExpanded, setIsExpanded] = useState(false);

	const toggleExpanded = () => {
		setIsExpanded(!isExpanded);
	};

	// const handleFeedbackClick = () => {
	// 	if (feedbackController.current) {
	// 		feedbackController.current.value = '';
	// 	}
	// 	// Show the feedback dialog
	// 	const feedback = window.prompt('Enter your feedback:');
	// 	if (feedback) {
	// 		// Assuming there's some way to send the feedback
	// 		onSendFeedback();
	// 	}
	// };

	return (
		<div className="p-1">
			<div className="flex justify-between items-center h-8">
				<div className="flex items-center">
					<>
						<CopyButton
							value={cardContent}
							className="btn btn-sm bg-transparent border-none flex items-center shadow-none"
							size={16}
						/>
						{isUser && (
							<button
								className="btn btn-sm bg-transparent border-none flex items-center shadow-none"
								onClick={onEdit}
								aria-label="Edit message"
							>
								<FiEdit size={16} />
							</button>
						)}
						{isUser && (
							<button
								className="btn btn-sm bg-transparent border-none flex items-center shadow-none"
								onClick={onResend}
								aria-label="Resend message"
							>
								<FiRepeat size={16} />
							</button>
						)}
						{/* <button
							className="btn btn-sm bg-transparent border-none flex items-center shadow-none"
							onClick={handleFeedbackClick}
							aria-label="Submit feedback"
						>
							<FiMessageCircle size={16} />
						</button> */}
					</>
				</div>
				{isStreaming && (
					<div className="text-sm">
						<span>Streaming...</span>
					</div>
				)}
				<button
					className="btn btn-sm bg-transparent border-none flex items-center shadow-none"
					onClick={toggleExpanded}
					aria-label="Details"
				>
					{isExpanded ? <FiChevronUp size={16} /> : <FiChevronDown size={16} />}
				</button>
			</div>
			{isExpanded && messageDetails && (
				<div className="mt-2 overflow-hidden shadow-lg rounded-2xl">
					<ChatMessageContent content={messageDetails} align="items-start text-left" streamedMessage="" />
				</div>
			)}
		</div>
	);
};

export default ChatMessageFooterArea;
