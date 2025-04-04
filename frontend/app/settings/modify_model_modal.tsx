import type { ModelSetting } from '@/models/settingmodel';
import type { FC } from 'react';
import React, { useState } from 'react';
import { FiX } from 'react-icons/fi';

interface ModifyModelModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (modelData: ModelSetting) => void;
	initialData?: ModelSetting;
	existingModels: ModelSetting[];
}

const ModifyModelModal: FC<ModifyModelModalProps> = ({ isOpen, onClose, onSubmit, initialData, existingModels }) => {
	const [name, setName] = useState(initialData?.name || '');
	const [temperature, setTemperature] = useState(initialData?.temperature || 0.1);
	const [isEnabled, setIsEnabled] = useState(initialData?.isEnabled ?? true);
	const [error, setError] = useState('');

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		// Validate model name uniqueness
		if (!initialData || initialData.name !== name) {
			if (existingModels.some(m => m.name === name)) {
				setError('Model name must be unique');
				return;
			}
		}

		const modelData: ModelSetting = {
			name: name,
			displayName: name,
			isEnabled: isEnabled,
			temperature: temperature,
		};

		onSubmit(modelData);
	};

	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<form method="dialog" className="modal-box w-5/6 max-w-2xl" onSubmit={handleSubmit}>
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{initialData ? 'Edit Model' : 'Add Model'}</h3>
					<button className="btn btn-sm btn-circle btn-ghost" onClick={onClose}>
						<FiX />
					</button>
				</div>
				<div className="space-y-4">
					{/* Model Name */}
					<div className="form-control">
						<label className="label">
							<span className="label-text">Model Name</span>
						</label>
						<input
							type="text"
							className={`input input-bordered ${error ? 'input-error' : ''}`}
							value={name}
							onChange={e => {
								setName(e.target.value);
								setError('');
							}}
							required
							spellCheck="false"
						/>
						{error && <p className="text-sm text-error mt-1">{error}</p>}
					</div>
					{/* Temperature */}
					<div className="form-control">
						<label className="label">
							<span className="label-text">Temperature</span>
						</label>
						<input
							type="number"
							step="0.01"
							min="0"
							max="1"
							className="input input-bordered"
							value={temperature}
							onChange={e => {
								setTemperature(parseFloat(e.target.value));
							}}
							required
							spellCheck="false"
						/>
					</div>
					{/* Enabled */}
					<div className="form-control">
						<label className="cursor-pointer label">
							<span className="label-text">Enabled</span>
							<input
								type="checkbox"
								className="toggle toggle-primary"
								checked={isEnabled}
								onChange={() => {
									setIsEnabled(!isEnabled);
								}}
							/>
						</label>
					</div>
				</div>
				<div className="modal-action">
					<button type="button" className="btn" onClick={onClose}>
						Cancel
					</button>
					<button type="submit" className="btn btn-primary">
						{initialData ? 'Save' : 'Add'}
					</button>
				</div>
			</form>
		</dialog>
	);
};

export default ModifyModelModal;
