import React from 'react';
import { FiAlertTriangle } from 'react-icons/fi';

interface DeleteDocStoreModalProps {
	isOpen: boolean;
	onClose: () => void;
	onConfirm: () => void;
	storeName: string;
}

const DeleteDocStoreModal: React.FC<DeleteDocStoreModalProps> = ({ isOpen, onClose, onConfirm, storeName }) => {
	if (!isOpen) return null;

	return (
		<div className="modal modal-open">
			<div className="modal-box rounded-2xl">
				<h3 className="font-bold text-lg flex items-center">
					<FiAlertTriangle className="text-warning mr-2" /> Confirm Document Store Deletion
				</h3>
				<p className="py-4">
					Are you sure you want to delete the document store {storeName}? This action cannot be undone.
				</p>
				<div className="modal-action">
					<button className="btn btn-ghost rounded-2xl" onClick={onClose}>
						Cancel
					</button>
					<button className="btn btn-error rounded-2xl" onClick={onConfirm}>
						Delete Document Store
					</button>
				</div>
			</div>
		</div>
	);
};

export default DeleteDocStoreModal;
