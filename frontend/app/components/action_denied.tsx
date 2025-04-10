// components/ActionDeniedAlert.tsx
import type { FC } from 'react';
import { FiAlertTriangle } from 'react-icons/fi';

interface ActionDeniedAlertProps {
	isOpen: boolean;
	onClose: () => void;
	message: string;
	title?: string;
}

const ActionDeniedAlert: FC<ActionDeniedAlertProps> = ({ isOpen, onClose, message, title = 'Action Not Allowed' }) => {
	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box w-11/12 max-w-md">
				<div className="flex items-center mb-4">
					<FiAlertTriangle size={24} className="text-warning mr-3" />
					<h3 className="font-bold text-lg">{title}</h3>
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
		</dialog>
	);
};

export default ActionDeniedAlert;
