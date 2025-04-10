import React, { type FC, useEffect, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import type { ModelName, ProviderName } from '@/models/aiprovidermodel';
import { DefaultModelSetting, type ModelSetting } from '@/models/settingmodel';

import { PopulateModelSettingDefaults } from '@/apis/settingstore_helper';

interface ModifyModelModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (modelName: ModelName, modelData: ModelSetting) => void;
	providerName: ProviderName;
	initialModelName?: ModelName; // if provided, indicates edit mode
	initialData?: ModelSetting; // when editing existing model data
	existingModels: Record<ModelName, ModelSetting>;
}

// We'll use a separate interface for the local form data,
// especially for numeric fields, so we can store them as strings (allowing blank).
interface ModelSettingFormData {
	displayName: string;
	isEnabled: boolean;
	stream: boolean;
	maxPromptLength: string;
	maxOutputLength: string;
	temperature: string;
	reasoningSupport: boolean;
	systemPrompt: string;
	timeout: string;
}

const ModifyModelModal: FC<ModifyModelModalProps> = ({
	isOpen,
	onClose,
	onSubmit,
	providerName,
	initialModelName,
	initialData,
	existingModels,
}) => {
	const isEditMode = Boolean(initialModelName);

	// Default numeric values that we also show in placeholders
	const [defaultValues, setDefaultValues] = useState<ModelSetting>(DefaultModelSetting);

	const [modelName, setModelName] = useState<ModelName>(initialModelName ?? ('' as ModelName));

	// Local form data (storing numeric as strings for easy clearing)
	const [formData, setFormData] = useState<ModelSettingFormData>({
		displayName: '',
		isEnabled: true,
		stream: false,
		maxPromptLength: '',
		maxOutputLength: '',
		temperature: '',
		reasoningSupport: false,
		systemPrompt: '',
		timeout: '',
	});

	const [errors, setErrors] = useState<{
		modelName?: string;
		displayName?: string;
		temperature?: string;
		maxPromptLength?: string;
		maxOutputLength?: string;
		timeout?: string;
	}>({});

	useEffect(() => {
		async function loadData() {
			let mName: ModelName = '';
			if (initialModelName) {
				mName = initialModelName;
			}
			const merged = await PopulateModelSettingDefaults(providerName, mName, initialData);
			setDefaultValues(merged);

			// Convert numbers to strings for local form usage
			setFormData({
				displayName: merged.displayName,
				isEnabled: merged.isEnabled,
				stream: merged.stream ?? false,
				maxPromptLength: String(merged.maxPromptLength ?? ''),
				maxOutputLength: String(merged.maxOutputLength ?? ''),
				temperature: String(merged.temperature ?? ''),
				reasoningSupport: merged.reasoningSupport ?? false,
				systemPrompt: merged.systemPrompt ?? '',
				timeout: String(merged.timeout ?? ''),
			});

			setModelName(mName);
			setErrors({});
		}

		if (isOpen) {
			void loadData();
		}
	}, [isOpen, providerName, initialModelName, initialData]);

	type ValidationField =
		| 'modelName'
		| 'displayName'
		| 'temperature'
		| 'maxPromptLength'
		| 'maxOutputLength'
		| 'timeout';
	type ValidationErrors = Partial<Record<ValidationField, string>>;

	const validateField = (field: ValidationField, value: unknown) => {
		const newErrors: ValidationErrors = Object.fromEntries(Object.entries(errors).filter(([key]) => key !== field));

		// Model Name is required only if we're adding a new model (not edit mode)
		if (field === 'modelName' && !isEditMode) {
			if (typeof value === 'string' && !value.trim()) {
				newErrors.modelName = 'Model name is required.';
			} else if (typeof value === 'string' && Object.prototype.hasOwnProperty.call(existingModels, value)) {
				newErrors.modelName = 'Model name must be unique.';
			}
		}

		// Display Name is always required
		if (field === 'displayName') {
			if (typeof value === 'string' && !value.trim()) {
				newErrors.displayName = 'Display name is required.';
			}
		}

		// For numeric fields, only validate if there's a non-empty string
		if (['temperature', 'maxPromptLength', 'maxOutputLength', 'timeout'].includes(field)) {
			let strVal = '';
			if (typeof value === 'number' || typeof value === 'boolean') {
				strVal = String(value);
			}
			strVal = strVal.trim();

			if (strVal.length > 0) {
				const numValue = Number(strVal);
				if (Number.isNaN(numValue)) {
					newErrors[field] = `${field} must be a valid number.`;
				} else {
					if (field === 'temperature' && (numValue < 0 || numValue > 1)) {
						newErrors.temperature = 'Temperature must be between 0 and 1.';
					}
					if ((field === 'maxPromptLength' || field === 'maxOutputLength' || field === 'timeout') && numValue < 1) {
						newErrors[field] = `${field} must be positive.`;
					}
				}
			}
		}

		setErrors(newErrors);
	};

	// ----------------------------------------------------
	// Handle changes for all fields
	// ----------------------------------------------------
	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const target = e.target as HTMLInputElement;
		const { name, value, type, checked } = target;

		if (type === 'checkbox') {
			setFormData(prev => ({ ...prev, [name]: checked }));
			return;
		}

		// Model Name
		if (name === 'modelName') {
			setModelName(value);
			validateField('modelName', value);
			return;
		}

		// For other text/numeric inputs, just store the raw string
		setFormData(prev => ({ ...prev, [name]: value }));
		// Trigger validation for required fields
		if (name === 'displayName') {
			validateField('displayName', value);
		}
		if (name === 'temperature' || name === 'maxPromptLength' || name === 'maxOutputLength' || name === 'timeout') {
			validateField(name as ValidationField, value);
		}
	};

	// ----------------------------------------------------
	// On form submit, parse final values and re-validate
	// ----------------------------------------------------
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		// Re-validate fields
		validateField('modelName', modelName);
		validateField('displayName', formData.displayName);
		validateField('temperature', formData.temperature);
		validateField('maxPromptLength', formData.maxPromptLength);
		validateField('maxOutputLength', formData.maxOutputLength);
		validateField('timeout', formData.timeout);

		// Check for errors
		const hasErrors = Object.keys(errors).length > 0;
		if (hasErrors) {
			// If errors exist, don’t proceed
			return;
		}

		// If required fields are missing, don't proceed
		if ((!isEditMode && !modelName.trim()) || !formData.displayName.trim()) {
			return;
		}

		// Convert strings to numbers (fallback to defaults) for numeric fields
		const parseOrDefault = (val: string, def: number) => (val.trim() === '' ? def : Number(val));

		const finalData: ModelSetting = {
			displayName: formData.displayName.trim(),
			isEnabled: formData.isEnabled,
			stream: formData.stream,
			maxPromptLength: parseOrDefault(formData.maxPromptLength, defaultValues.maxPromptLength ?? 2048),
			maxOutputLength: parseOrDefault(formData.maxOutputLength, defaultValues.maxOutputLength ?? 1024),
			temperature: parseOrDefault(formData.temperature, defaultValues.temperature ?? 0.1),
			reasoningSupport: formData.reasoningSupport,
			systemPrompt: formData.systemPrompt,
			timeout: parseOrDefault(formData.timeout, defaultValues.timeout ?? 60),
		};

		onSubmit(modelName, finalData);
	};

	// If we’re not open, return null (don’t render)
	if (!isOpen) return null;

	const numPlaceholder = (field: keyof ModelSetting) => {
		const value = defaultValues[field];

		if (value === undefined || typeof value === 'object') {
			return 'Default: N/A'; // Or any other placeholder for null/undefined
		}
		return `Default: ${String(value)}`; // Convert other types to string
	};

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* Header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{isEditMode ? 'Edit Model' : 'Add New Model'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close">
						<FiX size={12} />
					</button>
				</div>

				{/* Form */}
				<form onSubmit={handleSubmit} className="space-y-4">
					{/* Model Name (Internal Key) */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Model Name*</span>
							<span className="label-text-alt tooltip" data-tip="Unique identifier for this model">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="modelName"
								value={modelName}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-xl ${errors.modelName ? 'input-error' : ''}`}
								placeholder="e.g., gpt-4, claude-opus"
								disabled={isEditMode}
								spellCheck="false"
							/>
							{errors.modelName && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.modelName}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Display Name */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Display Name*</span>
							<span className="label-text-alt tooltip" data-tip="Name displayed to users">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="displayName"
								value={formData.displayName}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-xl ${errors.displayName ? 'input-error' : ''}`}
								placeholder="e.g., GPT-4, Claude Opus"
								spellCheck="false"
							/>
							{errors.displayName && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.displayName}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Toggle: isEnabled */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3 cursor-pointer">
							<span className="label-text text-sm">Enabled</span>
						</label>
						<div className="col-span-9">
							<input
								type="checkbox"
								name="isEnabled"
								checked={formData.isEnabled}
								onChange={handleChange}
								className="toggle toggle-primary rounded-full"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Toggle: stream */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3 cursor-pointer">
							<span className="label-text text-sm">Streaming</span>
						</label>
						<div className="col-span-9">
							<input
								type="checkbox"
								name="stream"
								checked={formData.stream}
								onChange={handleChange}
								className="toggle toggle-primary rounded-full"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Toggle: reasoningSupport */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3 cursor-pointer">
							<span className="label-text text-sm">Reasoning Support</span>
						</label>
						<div className="col-span-9">
							<input
								type="checkbox"
								name="reasoningSupport"
								checked={formData.reasoningSupport}
								onChange={handleChange}
								className="toggle toggle-primary rounded-full"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Temperature */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Temperature (0.0-1.0)</span>
							<span
								className="label-text-alt tooltip"
								data-tip="Controls randomness: lower values are more deterministic"
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="temperature"
								value={formData.temperature}
								onChange={handleChange}
								placeholder={numPlaceholder('temperature')}
								className={`input input-bordered w-full rounded-xl ${errors.temperature ? 'input-error' : ''}`}
								spellCheck="false"
							/>
							{errors.temperature && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.temperature}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Timeout */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Timeout (seconds)</span>
							<span className="label-text-alt tooltip" data-tip="Maximum time to wait for response">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="timeout"
								value={formData.timeout}
								onChange={handleChange}
								placeholder={numPlaceholder('timeout')}
								className={`input input-bordered w-full rounded-xl ${errors.timeout ? 'input-error' : ''}`}
								spellCheck="false"
							/>
							{errors.timeout && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.timeout}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Prompt Length */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Max Prompt Tokens</span>
							<span className="label-text-alt tooltip" data-tip="Maximum tokens for input">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="maxPromptLength"
								value={formData.maxPromptLength}
								onChange={handleChange}
								placeholder={numPlaceholder('maxPromptLength')}
								className={`input input-bordered w-full rounded-xl ${errors.maxPromptLength ? 'input-error' : ''}`}
								spellCheck="false"
							/>
							{errors.maxPromptLength && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.maxPromptLength}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Output Length */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Max Output Tokens</span>
							<span className="label-text-alt tooltip" data-tip="Maximum tokens for output">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="maxOutputLength"
								value={formData.maxOutputLength}
								onChange={handleChange}
								placeholder={numPlaceholder('maxOutputLength')}
								className={`input input-bordered w-full rounded-xl ${errors.maxOutputLength ? 'input-error' : ''}`}
								spellCheck="false"
							/>
							{errors.maxOutputLength && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.maxOutputLength}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* System Prompt */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">System Prompt</span>
							<span className="label-text-alt tooltip" data-tip="Instructions that define the model's behavior">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<textarea
								name="systemPrompt"
								value={formData.systemPrompt}
								onChange={handleChange}
								className="textarea textarea-bordered w-full rounded-xl h-24"
								placeholder="Enter system prompt instructions here..."
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Actions */}
					<div className="modal-action">
						<button type="button" className="btn rounded-xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-primary rounded-xl">
							{isEditMode ? 'Save Changes' : 'Add Model'}
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default ModifyModelModal;
