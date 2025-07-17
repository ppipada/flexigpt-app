import React from 'react';

import { FiAlertTriangle } from 'react-icons/fi';

interface ConfirmationModalProps {
	isOpen: boolean;
	onClose: () => void;
	onConfirm: () => void;
	title: string;
	message: string;
	confirmButtonText: string;
}

const DeleteConfirmationModal: React.FC<ConfirmationModalProps> = ({
	isOpen,
	onClose,
	onConfirm,
	title,
	message,
	confirmButtonText,
}) => {
	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box rounded-2xl">
				<h3 className="font-bold text-lg flex items-center">
					<FiAlertTriangle className="text-warning mr-2" /> {title}
				</h3>
				<p className="py-4">{message}</p>
				<div className="modal-action">
					<button className="btn btn-ghost rounded-2xl" onClick={onClose}>
						Cancel
					</button>
					<button className="btn btn-error rounded-2xl" onClick={onConfirm}>
						{confirmButtonText}
					</button>
				</div>
			</div>
		</dialog>
	);
};

export default DeleteConfirmationModal;
