import { useEffect, useRef } from 'react';

import { createPortal } from 'react-dom';

import { FiCode, FiX } from 'react-icons/fi';

import { MessageContent } from '@/chats/messages/message_content';

type MessageDetailsModalProps = {
	isOpen: boolean;
	onClose: () => void;
	messageID: string;
	title: string;
	content: string;
	isBusy: boolean;
};

export function MessageDetailsModal({ isOpen, onClose, messageID, title, content, isBusy }: MessageDetailsModalProps) {
	const dialogRef = useRef<HTMLDialogElement | null>(null);

	// Open the dialog natively when isOpen becomes true
	useEffect(() => {
		if (!isOpen) return;

		const dialog = dialogRef.current;
		if (!dialog) return;

		// In StrictMode effects can run twice; guard against double showModal().
		if (!dialog.open) {
			dialog.showModal();
		}
	}, [isOpen]);

	// Sync parent state whenever the dialog is closed (Esc, backdrop, or dialog.close()).
	const handleDialogClose = () => {
		onClose();
	};

	if (!isOpen) return null;

	return createPortal(
		<dialog ref={dialogRef} className="modal" onClose={handleDialogClose}>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-[80vw] overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					{/* header */}
					<div className="mb-4 flex items-center justify-between">
						<h3 className="flex items-center gap-2 text-lg font-bold">
							<FiCode size={16} />
							<span>{title}</span>
						</h3>
						<button
							className="btn btn-sm btn-circle bg-base-300"
							onClick={() => dialogRef.current?.close()}
							aria-label="Close"
						>
							<FiX size={12} />
						</button>
					</div>

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
