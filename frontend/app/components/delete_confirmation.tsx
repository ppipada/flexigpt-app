import { createPortal } from 'react-dom';

import { FiAlertTriangle } from 'react-icons/fi';

interface ConfirmationModalProps {
	isOpen: boolean;
	onClose: () => void;
	onConfirm: () => void;
	title: string;
	message: string;
	confirmButtonText: string;
}

export function DeleteConfirmationModal({
	isOpen,
	onClose,
	onConfirm,
	title,
	message,
	confirmButtonText,
}: ConfirmationModalProps) {
	if (!isOpen) return null;

	return createPortal(
		<dialog className="modal modal-open">
			<div className="modal-box bg-base-200 rounded-2xl">
				<h3 className="flex items-center text-lg font-bold">
					<FiAlertTriangle className="text-warning mr-2" /> {title}
				</h3>
				<p className="py-4">{message}</p>
				<div className="modal-action">
					<button className="btn bg-base-300 rounded-2xl" onClick={onClose}>
						Cancel
					</button>
					<button className="btn btn-error rounded-2xl" onClick={onConfirm}>
						{confirmButtonText}
					</button>
				</div>
			</div>
		</dialog>,
		document.body
	);
}
