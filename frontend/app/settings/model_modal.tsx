import type { ModelName } from '@/models/aiprovidermodel';
import type { ModelSetting } from '@/models/settingmodel';
import type { FC } from 'react';
import React, { useEffect, useState } from 'react';
import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

interface ModifyModelModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (modelName: ModelName, modelData: ModelSetting) => void;
	initialModelName?: ModelName; // if provided, indicates edit mode
	initialData?: ModelSetting;
	existingModels: Record<ModelName, ModelSetting>;
}

const ModifyModelModal: FC<ModifyModelModalProps> = ({
	isOpen,
	onClose,
	onSubmit,
	initialModelName,
	initialData,
	existingModels,
}) => {
	const isEditMode = !!initialModelName;

	const [modelName, setModelName] = useState<ModelName>(initialModelName || ('' as ModelName));
	const [formData, setFormData] = useState<ModelSetting>({
		displayName: '',
		isEnabled: true,
		stream: false,
		promptLength: 2048,
		outputLength: 1024,
		temperature: 0.7,
		reasoningSupport: false,
		systemPrompt: '',
		timeout: 60,
	});

	const [errors, setErrors] = useState<{
		modelName?: string;
		displayName?: string;
		temperature?: string;
		promptLength?: string;
		outputLength?: string;
		timeout?: string;
	}>({});

	useEffect(() => {
		if (isOpen) {
			setModelName(initialModelName || ('' as ModelName));
			setFormData(
				initialData || {
					displayName: '',
					isEnabled: true,
					stream: false,
					promptLength: 2048,
					outputLength: 1024,
					temperature: 0.7,
					reasoningSupport: false,
					systemPrompt: '',
					timeout: 60,
				}
			);
			setErrors({});
		}
	}, [isOpen, initialModelName, initialData]);

	const validateField = (
		field: 'modelName' | 'displayName' | 'temperature' | 'promptLength' | 'outputLength' | 'timeout',
		value: any
	) => {
		const newErrors = { ...errors };
		delete newErrors[field];

		if (field === 'modelName') {
			if (!value.trim()) {
				newErrors.modelName = 'Model name is required.';
			} else if (!isEditMode && existingModels.hasOwnProperty(value)) {
				newErrors.modelName = 'Model name must be unique.';
			}
		}

		if (field === 'displayName' && !value.trim()) {
			newErrors.displayName = 'Display name is required.';
		}

		if (field === 'temperature') {
			if (value === '' || value === null || isNaN(value)) {
				newErrors.temperature = 'Temperature is required.';
			} else if (value < 0 || value > 1) {
				newErrors.temperature = 'Temperature must be between 0 and 1.';
			}
		}

		if (field === 'promptLength') {
			if (value === '' || value === null || isNaN(value)) {
				newErrors.promptLength = 'Prompt length is required.';
			} else if (value < 1) {
				newErrors.promptLength = 'Prompt length must be positive.';
			}
		}

		if (field === 'outputLength') {
			if (value === '' || value === null || isNaN(value)) {
				newErrors.outputLength = 'Output length is required.';
			} else if (value < 1) {
				newErrors.outputLength = 'Output length must be positive.';
			}
		}

		if (field === 'timeout') {
			if (value === '' || value === null || isNaN(value)) {
				newErrors.timeout = 'Timeout is required.';
			} else if (value < 1) {
				newErrors.timeout = 'Timeout must be positive.';
			}
		}

		setErrors(newErrors);
	};

	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const target = e.target as HTMLInputElement;
		const { name, value, type, checked } = target;
		const val = type === 'checkbox' ? checked : type === 'number' ? parseFloat(value) : value;

		if (name === 'modelName') {
			setModelName(val as ModelName);
			validateField(name, val);
		} else {
			setFormData(prev => ({ ...prev, [name]: val }));
			if (['displayName', 'temperature', 'promptLength', 'outputLength', 'timeout'].includes(name)) {
				validateField(name as 'displayName' | 'temperature' | 'promptLength' | 'outputLength' | 'timeout', val);
			}
		}
	};

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		// Validate all fields before submission
		validateField('modelName', modelName);
		validateField('displayName', formData.displayName);
		validateField('temperature', formData.temperature);
		validateField('promptLength', formData.promptLength);
		validateField('outputLength', formData.outputLength);
		validateField('timeout', formData.timeout);

		// Check if there are any errors
		if (Object.keys(errors).length === 0 && modelName && formData.displayName) {
			onSubmit(modelName, formData);
		}
	};

	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl rounded-lg">
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{isEditMode ? 'Edit Model' : 'Add New Model'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close">
						<FiX size={16} />
					</button>
				</div>

				<form onSubmit={handleSubmit} className="space-y-4">
					<div className="grid grid-cols-1 md:grid-cols-2 gap-4">
						{/* Model Name (internal key) */}
						<div className="form-control">
							<label className="label">
								<span className="label-text">Model Name (Internal Key)*</span>
								<span className="label-text-alt tooltip tooltip-left" data-tip="Unique identifier for this model">
									<FiHelpCircle size={16} />
								</span>
							</label>
							<input
								type="text"
								name="modelName"
								value={modelName}
								onChange={handleChange}
								className={`input input-bordered w-full ${errors.modelName ? 'input-error' : ''}`}
								required
								disabled={isEditMode}
								placeholder="e.g., gpt-4, claude-opus"
							/>
							{errors.modelName && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={14} /> {errors.modelName}
									</span>
								</div>
							)}
						</div>

						{/* Display Name */}
						<div className="form-control">
							<label className="label">
								<span className="label-text">Display Name*</span>
								<span className="label-text-alt tooltip tooltip-left" data-tip="Name displayed to users">
									<FiHelpCircle size={16} />
								</span>
							</label>
							<input
								type="text"
								name="displayName"
								value={formData.displayName}
								onChange={handleChange}
								className={`input input-bordered w-full ${errors.displayName ? 'input-error' : ''}`}
								required
								placeholder="e.g., GPT-4, Claude Opus"
							/>
							{errors.displayName && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={14} /> {errors.displayName}
									</span>
								</div>
							)}
						</div>

						{/* Temperature */}
						<div className="form-control">
							<label className="label">
								<span className="label-text">Temperature (0.0 - 1.0)*</span>
								<span
									className="label-text-alt tooltip tooltip-left"
									data-tip="Controls randomness: lower values are more deterministic"
								>
									<FiHelpCircle size={16} />
								</span>
							</label>
							<input
								type="number"
								name="temperature"
								step="0.01"
								min="0"
								max="1"
								value={formData.temperature}
								onChange={handleChange}
								className={`input input-bordered w-full ${errors.temperature ? 'input-error' : ''}`}
								required
							/>
							{errors.temperature && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={14} /> {errors.temperature}
									</span>
								</div>
							)}
						</div>

						{/* Timeout */}
						<div className="form-control">
							<label className="label">
								<span className="label-text">Timeout (seconds)*</span>
								<span className="label-text-alt tooltip tooltip-left" data-tip="Maximum time to wait for response">
									<FiHelpCircle size={16} />
								</span>
							</label>
							<input
								type="number"
								name="timeout"
								step="1"
								min="1"
								value={formData.timeout}
								onChange={handleChange}
								className={`input input-bordered w-full ${errors.timeout ? 'input-error' : ''}`}
								required
							/>
							{errors.timeout && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={14} /> {errors.timeout}
									</span>
								</div>
							)}
						</div>

						{/* Prompt Length */}
						<div className="form-control">
							<label className="label">
								<span className="label-text">Prompt Length (tokens)*</span>
								<span className="label-text-alt tooltip tooltip-left" data-tip="Maximum tokens for input">
									<FiHelpCircle size={16} />
								</span>
							</label>
							<input
								type="number"
								name="promptLength"
								step="1"
								min="1"
								value={formData.promptLength}
								onChange={handleChange}
								className={`input input-bordered w-full ${errors.promptLength ? 'input-error' : ''}`}
								required
							/>
							{errors.promptLength && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={14} /> {errors.promptLength}
									</span>
								</div>
							)}
						</div>

						{/* Output Length */}
						<div className="form-control">
							<label className="label">
								<span className="label-text">Output Length (tokens)*</span>
								<span className="label-text-alt tooltip tooltip-left" data-tip="Maximum tokens for output">
									<FiHelpCircle size={16} />
								</span>
							</label>
							<input
								type="number"
								name="outputLength"
								step="1"
								min="1"
								value={formData.outputLength}
								onChange={handleChange}
								className={`input input-bordered w-full ${errors.outputLength ? 'input-error' : ''}`}
								required
							/>
							{errors.outputLength && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={14} /> {errors.outputLength}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Toggle Options */}
					<div className="divider"></div>
					<div className="grid grid-cols-1 md:grid-cols-3 gap-4">
						<div className="form-control">
							<label className="label cursor-pointer justify-start gap-3">
								<input
									type="checkbox"
									name="isEnabled"
									checked={formData.isEnabled}
									onChange={handleChange}
									className="checkbox checkbox-primary"
								/>
								<span className="label-text">Enabled</span>
							</label>
						</div>

						<div className="form-control">
							<label className="label cursor-pointer justify-start gap-3">
								<input
									type="checkbox"
									name="stream"
									checked={formData.stream}
									onChange={handleChange}
									className="checkbox checkbox-primary"
								/>
								<span className="label-text">Streaming</span>
							</label>
						</div>

						<div className="form-control">
							<label className="label cursor-pointer justify-start gap-3">
								<input
									type="checkbox"
									name="reasoningSupport"
									checked={formData.reasoningSupport}
									onChange={handleChange}
									className="checkbox checkbox-primary"
								/>
								<span className="label-text">Reasoning Support</span>
							</label>
						</div>
					</div>

					{/* System Prompt */}
					<div className="form-control mt-2">
						<label className="label">
							<span className="label-text">System Prompt</span>
							<span
								className="label-text-alt tooltip tooltip-left"
								data-tip="Instructions that define the model's behavior"
							>
								<FiHelpCircle size={16} />
							</span>
						</label>
						<textarea
							name="systemPrompt"
							value={formData.systemPrompt}
							onChange={handleChange}
							className="textarea textarea-bordered h-24"
							placeholder="Enter system prompt instructions here..."
						/>
					</div>

					{/* Actions */}
					<div className="modal-action">
						<button type="button" className="btn" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-primary">
							{isEditMode ? 'Save Changes' : 'Add Model'}
						</button>
					</div>
				</form>
			</div>
			<div className="modal-backdrop" onClick={onClose}></div>
		</dialog>
	);
};

export default ModifyModelModal;
