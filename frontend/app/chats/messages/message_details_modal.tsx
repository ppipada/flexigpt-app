import { useEffect } from 'react';

import { createPortal } from 'react-dom';

import { FiInfo, FiX } from 'react-icons/fi';

import { MessageContent } from '@/chats/messages/message_content';

type MessageDetailsModalProps = {
	isOpen: boolean;
	onClose: () => void;
	messageID: string;
	title?: string;
	content: string;
	isBusy: boolean;
};

export function MessageDetailsModal({
	isOpen,
	onClose,
	messageID,
	title = 'Message Details',
	content,
	isBusy,
}: MessageDetailsModalProps) {
	// Close on Escape
	useEffect(() => {
		if (!isOpen) return;

		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				event.stopPropagation();
				onClose();
			}
		};

		window.addEventListener('keydown', handleKeyDown);
		return () => {
			window.removeEventListener('keydown', handleKeyDown);
		};
	}, [isOpen, onClose]);

	if (!isOpen) return null;

	return createPortal(
		<dialog className="modal modal-open">
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-[80vw] overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					{/* header */}
					<div className="mb-4 flex items-center justify-between">
						<h3 className="flex items-center gap-2 text-lg font-bold">
							<FiInfo size={16} />
							<span>{title}</span>
						</h3>
						<button className="btn btn-sm btn-circle bg-base-300" onClick={onClose} aria-label="Close">
							<FiX size={12} />
						</button>
					</div>

					{/* body */}

					<div className="mt-2">
						<MessageContent
							messageID={messageID}
							content={content}
							streamedText=""
							isStreaming={false}
							isBusy={isBusy}
							isPending={false}
							align="items-start text-left"
							renderAsMarkdown={true}
						/>
					</div>
				</div>
			</div>
		</dialog>,
		document.body
	);
}
