import React, { type FC, useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import { type ModelName, type ProviderName, ReasoningLevel, ReasoningType } from '@/models/aiprovidermodel';
import { DefaultModelSetting, type ModelSetting } from '@/models/settingmodel';

import { PopulateModelSettingDefaults } from '@/apis/settingstore_helper';

import Dropdown from '@/components/dropdown';

// For ReasoningType: a record where every option is `isEnabled: true`.
// That ensures the dropdown can show them. We also set filterDisabled={false} to skip filtering.
const reasoningTypeItems: Record<ReasoningType, { isEnabled: boolean; displayName: string }> = {
	[ReasoningType.SingleWithLevels]: { isEnabled: true, displayName: 'Reasoning only, with Levels' },
	[ReasoningType.HybridWithTokens]: { isEnabled: true, displayName: 'Hybrid, with Reasoning Tokens' },
};

const reasoningLevelItems: Record<ReasoningLevel, { isEnabled: boolean; displayName: string }> = {
	[ReasoningLevel.Low]: { isEnabled: true, displayName: 'Low' },
	[ReasoningLevel.Medium]: { isEnabled: true, displayName: 'Medium' },
	[ReasoningLevel.High]: { isEnabled: true, displayName: 'High' },
};

interface ModifyModelModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (modelName: ModelName, modelData: ModelSetting) => void;
	providerName: ProviderName;
	// If provided, indicates edit mode.
	initialModelName?: ModelName;
	// When editing existing model data.
	initialData?: ModelSetting;
	existingModels: Record<ModelName, ModelSetting>;
}

interface ModelSettingFormData {
	displayName: string;
	isEnabled: boolean;
	stream: boolean;
	maxPromptLength: string;
	maxOutputLength: string;
	temperature: string;

	reasoningSupport: boolean;
	reasoningType?: ReasoningType;
	reasoningLevel?: ReasoningLevel;
	reasoningTokens?: string;

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

	// Default numeric values and placeholders.
	const [defaultValues, setDefaultValues] = useState<ModelSetting>(DefaultModelSetting);

	// If not edit mode, user must provide a new modelName.
	const [modelName, setModelName] = useState<ModelName>(initialModelName ?? ('' as ModelName));

	// Local form data (storing numeric as strings for easy clearing).
	const [formData, setFormData] = useState<ModelSettingFormData>({
		displayName: '',
		isEnabled: true,
		stream: false,
		maxPromptLength: '',
		maxOutputLength: '',
		temperature: '',
		reasoningSupport: false,
		reasoningType: ReasoningType.SingleWithLevels,
		reasoningLevel: ReasoningLevel.Medium,
		reasoningTokens: '',
		systemPrompt: '',
		timeout: '',
	});

	// Validation errors.
	const [errors, setErrors] = useState<{
		modelName?: string;
		displayName?: string;
		temperature?: string;
		maxPromptLength?: string;
		maxOutputLength?: string;
		timeout?: string;
		reasoningTokens?: string;
	}>({});

	useEffect(() => {
		async function loadData() {
			let mName: ModelName = '';
			if (initialModelName) {
				mName = initialModelName;
			}
			const merged = await PopulateModelSettingDefaults(providerName, mName, initialData);
			setDefaultValues(merged);

			// Convert numbers to strings for the local form usage.
			setFormData({
				displayName: merged.displayName,
				isEnabled: merged.isEnabled,
				stream: merged.stream ?? false,
				maxPromptLength: String(merged.maxPromptLength ?? ''),
				maxOutputLength: String(merged.maxOutputLength ?? ''),
				temperature: String(merged.temperature ?? ''),
				reasoningSupport: !!merged.reasoning,
				reasoningType: merged.reasoning?.type ?? ReasoningType.SingleWithLevels,
				reasoningLevel: merged.reasoning?.level ?? ReasoningLevel.Medium,
				reasoningTokens: merged.reasoning?.tokens ? String(merged.reasoning.tokens) : '',
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
		| 'timeout'
		| 'reasoningTokens';

	type ValidationErrors = Partial<Record<ValidationField, string>>;

	const validateField = (field: ValidationField, value: unknown) => {
		// Remove old error for this field.
		const newErrors: ValidationErrors = Object.fromEntries(Object.entries(errors).filter(([key]) => key !== field));

		// Model Name is required only if we're adding a new model (not edit mode).
		if (field === 'modelName' && !isEditMode) {
			if (typeof value === 'string' && !value.trim()) {
				newErrors.modelName = 'Model name is required.';
			} else if (typeof value === 'string' && Object.prototype.hasOwnProperty.call(existingModels, value)) {
				newErrors.modelName = 'Model name must be unique.';
			}
		}

		// Display Name is always required.
		if (field === 'displayName') {
			if (typeof value === 'string' && !value.trim()) {
				newErrors.displayName = 'Display name is required.';
			}
		}

		// Validate numeric fields if a non-empty string is provided.
		if (['temperature', 'maxPromptLength', 'maxOutputLength', 'timeout'].includes(field)) {
			let strVal = '';
			if (typeof value === 'number' || typeof value === 'boolean') {
				strVal = String(value);
			} else if (typeof value === 'string') {
				strVal = value.trim();
			}

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

		// Validate reasoningTokens if the reasoningType is HybridWithTokens.
		if (field === 'reasoningTokens' && formData.reasoningType === ReasoningType.HybridWithTokens) {
			if (typeof value === 'string') {
				const strVal = value.trim();
				if (strVal.length > 0) {
					const numValue = Number(strVal);
					if (Number.isNaN(numValue) || numValue < 1024) {
						newErrors.reasoningTokens = 'Reasoning tokens must be a positive number > 1024.';
					}
				}
			}
		}

		setErrors(newErrors);
	};

	// Handle changes for all fields.
	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;

		if (type === 'checkbox') {
			setFormData(prev => ({ ...prev, [name]: checked }));
			return;
		}

		// If changing "modelName" (only relevant if adding a new model).
		if (name === 'modelName') {
			setModelName(value);
			validateField('modelName', value);
			return;
		}

		setFormData(prev => ({ ...prev, [name]: value }));

		// Trigger validation for relevant fields.
		if (name === 'displayName') {
			validateField('displayName', value);
		}
		if (
			name === 'temperature' ||
			name === 'maxPromptLength' ||
			name === 'maxOutputLength' ||
			name === 'timeout' ||
			name === 'reasoningTokens'
		) {
			validateField(name as ValidationField, value);
		}
	};

	// Compute if form is valid for "Add Model" / "Save Changes" button.
	const isAllValid = useMemo(() => {
		// No errors in the errors object.
		const noErrors = !Object.values(errors).some(Boolean);

		// If adding a model, modelName must be non-empty.
		const modelNameValid = isEditMode || modelName.trim().length > 0;

		// Always require a non-empty displayName.
		const displayNameValid = formData.displayName.trim().length > 0;

		return noErrors && modelNameValid && displayNameValid;
	}, [errors, isEditMode, modelName, formData.displayName]);

	// On form submit, parse final values and re-validate.
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		// Re-validate fields.
		validateField('modelName', modelName);
		validateField('displayName', formData.displayName);
		validateField('temperature', formData.temperature);
		validateField('maxPromptLength', formData.maxPromptLength);
		validateField('maxOutputLength', formData.maxOutputLength);
		validateField('timeout', formData.timeout);
		validateField('reasoningTokens', formData.reasoningTokens);

		// If any errors remain, or the required fields are empty, do not proceed.
		const hasErrors = Object.keys(errors).length > 0;
		if (hasErrors || !isAllValid) {
			return;
		}

		// Convert strings to numbers (fallback to defaults) for numeric fields.
		const parseOrDefault = (val: string, def: number) => (val.trim() === '' ? def : Number(val));

		const finalData: ModelSetting = {
			displayName: formData.displayName.trim(),
			isEnabled: formData.isEnabled,
			stream: formData.stream,
			maxPromptLength: parseOrDefault(formData.maxPromptLength, defaultValues.maxPromptLength ?? 2048),
			maxOutputLength: parseOrDefault(formData.maxOutputLength, defaultValues.maxOutputLength ?? 1024),
			temperature: parseOrDefault(formData.temperature, defaultValues.temperature ?? 0.1),
			systemPrompt: formData.systemPrompt,
			timeout: parseOrDefault(formData.timeout, defaultValues.timeout ?? 60),
		};

		// Build reasoning object if reasoningSupport is true.
		if (formData.reasoningSupport) {
			finalData.reasoning = {
				type: formData.reasoningType ?? ReasoningType.SingleWithLevels,
				level: formData.reasoningLevel ?? ReasoningLevel.Medium,
				tokens: parseOrDefault(formData.reasoningTokens ?? '', defaultValues.reasoning?.tokens ?? 0),
			};
		} else {
			finalData.reasoning = undefined;
		}

		onSubmit(modelName, finalData);
	};

	// If the modal isn’t open, return null.
	if (!isOpen) return null;

	// Helper for numeric placeholders showing "Default: X".
	const numPlaceholder = (field: keyof ModelSetting) => {
		const value = defaultValues[field];
		if (value === undefined || typeof value === 'object') {
			return 'Default: N/A'; // or any other placeholder for null/undefined
		}
		return `Default: ${String(value)}`;
	};

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* Header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{isEditMode ? 'Edit Model' : 'Add New Model'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close" title="Close">
						<FiX size={12} />
					</button>
				</div>

				{/* Form */}
				<form onSubmit={handleSubmit} className="space-y-4">
					{/* Model Name (Internal Key) */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Model Name*</span>
							<span className="label-text-alt tooltip" data-tip="Unique identifier for this model.">
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
								disabled={isEditMode} // Cannot change modelName if editing an existing model
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
								className="toggle toggle-accent rounded-full"
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
								className="toggle toggle-accent rounded-full"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Toggle: reasoningSupport */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3 cursor-pointer">
							<span className="label-text text-sm">Supports Reasoning</span>
							<span
								className="label-text-alt tooltip"
								data-tip="If enabled, configure below how the model handles reasoning"
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="checkbox"
								name="reasoningSupport"
								checked={formData.reasoningSupport}
								onChange={handleChange}
								className="toggle toggle-accent rounded-full"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Reasoning Type (use Dropdown instead of <select>) */}
					{formData.reasoningSupport && (
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Reasoning Type</span>
								<span className="label-text-alt tooltip" data-tip="singleWithLevels or hybridWithTokens">
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<Dropdown<ReasoningType>
									dropdownItems={reasoningTypeItems}
									selectedKey={formData.reasoningType ?? ReasoningType.SingleWithLevels}
									onChange={newType => {
										setFormData(prev => ({ ...prev, reasoningType: newType }));
									}}
									filterDisabled={false}
									title="Select Reasoning Type"
									getDisplayName={key => reasoningTypeItems[key].displayName}
								/>
							</div>
						</div>
					)}

					{/* Reasoning Level (use Dropdown if reasoningType === SingleWithLevels) */}
					{formData.reasoningSupport && formData.reasoningType === ReasoningType.SingleWithLevels && (
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Reasoning Level</span>
								<span className="label-text-alt tooltip" data-tip="Pick how deep the chain-of-thought is">
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<Dropdown<ReasoningLevel>
									dropdownItems={reasoningLevelItems}
									selectedKey={formData.reasoningLevel ?? ReasoningLevel.Medium}
									onChange={newLevel => {
										setFormData(prev => ({ ...prev, reasoningLevel: newLevel }));
									}}
									filterDisabled={false}
									title="Select Reasoning Level"
									getDisplayName={key => reasoningLevelItems[key].displayName}
								/>
							</div>
						</div>
					)}

					{/* Reasoning Tokens (show if reasoningType === hybridWithTokens) */}
					{formData.reasoningSupport && formData.reasoningType === ReasoningType.HybridWithTokens && (
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Reasoning Tokens</span>
								<span className="label-text-alt tooltip" data-tip="Number of tokens dedicated to reasoning">
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<input
									type="text"
									name="reasoningTokens"
									value={formData.reasoningTokens}
									onChange={handleChange}
									className={`input input-bordered w-full rounded-xl ${errors.reasoningTokens ? 'input-error' : ''}`}
									placeholder="e.g., 1024"
									spellCheck="false"
								/>
								{errors.reasoningTokens && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.reasoningTokens}
										</span>
									</div>
								)}
							</div>
						</div>
					)}

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

					{/* Max Prompt Length */}
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

					{/* Max Output Length */}
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
						<button type="submit" className="btn btn-primary rounded-xl" disabled={!isAllValid}>
							{isEditMode ? 'Save Changes' : 'Add Model'}
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default ModifyModelModal;
