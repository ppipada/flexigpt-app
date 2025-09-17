// components/ActionDeniedAlert.tsx
import type { FC } from 'react';

import { createPortal } from 'react-dom';

import { FiAlertTriangle } from 'react-icons/fi';

interface ActionDeniedAlertProps {
	isOpen: boolean;
	onClose: () => void;
	message: string;
	title?: string;
}

const ActionDeniedAlert: FC<ActionDeniedAlertProps> = ({ isOpen, onClose, message, title = 'Action Not Allowed' }) => {
	if (!isOpen) return null;

	return createPortal(
		<dialog className="modal modal-open">
			<div className="modal-box bg-base-200 w-11/12 max-w-md">
				<div className="mb-4 flex items-center">
					<FiAlertTriangle size={24} className="text-warning mr-3" />
					<h3 className="text-lg font-bold">{title}</h3>
				</div>
				<p className="py-2">{message}</p>
				<div className="modal-action">
					<button className="btn btn-primary" onClick={onClose}>
						OK
					</button>
				</div>
			</div>
			<form method="dialog" className="modal-backdrop">
				<button onClick={onClose}>close</button>
			</form>
		</dialog>,
		document.body
	);
};

export default ActionDeniedAlert;
