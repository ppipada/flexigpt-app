import React, { type FC, useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import {
	DefaultModelPreset,
	type ModelPreset,
	type ModelPresetID,
	type ProviderName,
	ReasoningLevel,
	ReasoningType,
} from '@/models/modelpresetsmodel';

import { PopulateModelPresetDefaults } from '@/apis/modelpresetstore_helper';

import Dropdown from '@/components/dropdown';

// For ReasoningType: a record where every option is `isEnabled: true`.
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
	onSubmit: (modelPresetID: ModelPresetID, modelData: ModelPreset) => void;
	providerName: ProviderName;
	// If provided, indicates edit mode.
	initialModelID?: ModelPresetID;
	// When editing existing model data.
	initialData?: ModelPreset;
	existingModels: Record<ModelPresetID, ModelPreset>;
}

interface ModelPresetFormData {
	presetLabel: string; // replaces displayName
	name: string; // Model Name (for completions API)
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
	initialModelID,
	initialData,
	existingModels,
}) => {
	const isEditMode = Boolean(initialModelID);

	// Default numeric values and placeholders.
	const [defaultValues, setDefaultValues] = useState<ModelPreset>(DefaultModelPreset);

	// If not edit mode, user must provide a new modelPresetID.
	const [modelPresetID, setModelPresetID] = useState<ModelPresetID>(initialModelID ?? ('' as ModelPresetID));

	// Local form data (storing numeric as strings for easy clearing).
	const [formData, setFormData] = useState<ModelPresetFormData>({
		presetLabel: '',
		name: '',
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

	const [errors, setErrors] = useState<{
		modelPresetID?: string;
		name?: string;
		presetLabel?: string;
		temperature?: string;
		maxPromptLength?: string;
		maxOutputLength?: string;
		timeout?: string;
		reasoningTokens?: string;
	}>({});

	useEffect(() => {
		async function loadData() {
			let mID: ModelPresetID = '';
			if (initialModelID) {
				mID = initialModelID;
			}
			const merged = await PopulateModelPresetDefaults(providerName, mID, initialData);
			setDefaultValues(merged);

			// Convert numbers to strings for the local form usage.
			setFormData({
				presetLabel: merged.displayName,
				name: merged.name,
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

			setModelPresetID(mID);
			setErrors({});
		}

		if (isOpen) {
			void loadData();
		}
	}, [isOpen, providerName, initialModelID, initialData]);

	type ValidationField =
		| 'modelPresetID'
		| 'name'
		| 'presetLabel'
		| 'temperature'
		| 'maxPromptLength'
		| 'maxOutputLength'
		| 'timeout'
		| 'reasoningTokens';

	type ValidationErrors = Partial<Record<ValidationField, string>>;

	const validateField = (field: ValidationField, value: unknown) => {
		// Remove old error for this field.
		const newErrors: ValidationErrors = Object.fromEntries(Object.entries(errors).filter(([key]) => key !== field));

		// ModelPresetID: required, unique, alphanumeric only (no spaces or special chars)
		if (field === 'modelPresetID' && !isEditMode) {
			if (typeof value === 'string') {
				const trimmed = value.trim();
				if (!trimmed) {
					newErrors.modelPresetID = 'Model Preset ID is required.';
				} else if (!/^[a-zA-Z0-9-]+$/.test(trimmed)) {
					newErrors.modelPresetID = 'Only letters, numbers, and hyphens allowed.';
				} else if (Object.prototype.hasOwnProperty.call(existingModels, trimmed)) {
					newErrors.modelPresetID = 'Model Preset ID must be unique.';
				}
			}
		}

		// Model Name is always required.
		if (field === 'name') {
			if (typeof value === 'string' && !value.trim()) {
				newErrors.name = 'Model Name is required.';
			}
		}

		// Preset Label is always required.
		if (field === 'presetLabel') {
			if (typeof value === 'string' && !value.trim()) {
				newErrors.presetLabel = 'Preset Label is required.';
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

	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;

		if (type === 'checkbox') {
			setFormData(prev => ({ ...prev, [name]: checked }));
			return;
		}

		// If changing "modelPresetID" (only relevant if adding a new model).
		if (name === 'modelPresetID') {
			setModelPresetID(value);
			validateField('modelPresetID', value);
			return;
		}

		if (name === 'name') {
			setFormData(prev => ({ ...prev, name: value }));
			validateField('name', value);
			return;
		}

		if (name === 'presetLabel') {
			setFormData(prev => ({ ...prev, presetLabel: value }));
			validateField('presetLabel', value);
			return;
		}

		setFormData(prev => ({ ...prev, [name]: value }));

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

	const isAllValid = useMemo(() => {
		const noErrors = !Object.values(errors).some(Boolean);
		const modelPresetIDValid = isEditMode || (modelPresetID.trim().length > 0 && /^[a-zA-Z0-9-]+$/.test(modelPresetID));
		const presetLabelValid = formData.presetLabel.trim().length > 0;
		const nameValid = formData.name.trim().length > 0;
		return noErrors && modelPresetIDValid && presetLabelValid && nameValid;
	}, [errors, isEditMode, modelPresetID, formData.presetLabel, formData.name]);

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		validateField('modelPresetID', modelPresetID);
		validateField('name', formData.name);
		validateField('presetLabel', formData.presetLabel);
		validateField('temperature', formData.temperature);
		validateField('maxPromptLength', formData.maxPromptLength);
		validateField('maxOutputLength', formData.maxOutputLength);
		validateField('timeout', formData.timeout);
		validateField('reasoningTokens', formData.reasoningTokens);

		const hasErrors = Object.keys(errors).length > 0;
		if (hasErrors || !isAllValid) {
			return;
		}

		const parseOrDefault = (val: string, def: number) => (val.trim() === '' ? def : Number(val));
		const finalModelPresetID = modelPresetID.trim();

		const finalData: ModelPreset = {
			id: finalModelPresetID,
			name: formData.name.trim(),
			displayName: formData.presetLabel.trim(),
			isEnabled: formData.isEnabled,
			shortCommand: finalModelPresetID,
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

		onSubmit(finalModelPresetID, finalData);
	};

	// If the modal isn’t open, return null.
	if (!isOpen) return null;

	// Helper for numeric placeholders showing "Default: X".
	const numPlaceholder = (field: keyof ModelPreset) => {
		const value = defaultValues[field];
		if (value === undefined || typeof value === 'object') {
			return 'Default: N/A';
		}
		return `Default: ${String(value)}`;
	};

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* Header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{isEditMode ? 'Edit Model Preset' : 'Add Model Preset'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close" title="Close">
						<FiX size={12} />
					</button>
				</div>

				{/* Form */}
				<form onSubmit={handleSubmit} className="space-y-4">
					{/* ModelPresetID (Internal Key) */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Model Preset ID*</span>
							<span
								className="label-text-alt tooltip tooltip-right"
								data-tip="Unique identifier for this preset. Can use alphanumeric chars and hyphen."
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="modelPresetID"
								value={modelPresetID}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-xl ${errors.modelPresetID ? 'input-error' : ''}`}
								placeholder="e.g. gpt4preset, claudeOpus1"
								disabled={isEditMode}
								spellCheck="false"
								autoComplete="off"
							/>
							{errors.modelPresetID && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.modelPresetID}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Model Name (for completions API) */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Model Name*</span>
							<span
								className="label-text-alt tooltip tooltip-right"
								data-tip="The model name to send to the completions API (e.g. gpt-4, anthropic/claude-3-opus-20240229)"
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="name"
								value={formData.name}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-xl ${errors.name ? 'input-error' : ''}`}
								placeholder="e.g. gpt-4, anthropic/claude-3-opus-20240229"
								spellCheck="false"
								autoComplete="off"
							/>
							{errors.name && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.name}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Preset Label */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Preset Label*</span>
							<span
								className="label-text-alt tooltip tooltip-right"
								data-tip="A friendly name for this preset, shown in UI."
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="presetLabel"
								value={formData.presetLabel}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-xl ${errors.presetLabel ? 'input-error' : ''}`}
								placeholder="e.g. GPT-4 (Creative), Claude Opus Fast"
								spellCheck="false"
								autoComplete="off"
							/>
							{errors.presetLabel && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.presetLabel}
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
								className="label-text-alt tooltip tooltip-right"
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

					{/* Reasoning Type */}
					{formData.reasoningSupport && (
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Reasoning Type</span>
								<span className="label-text-alt tooltip tooltip-right" data-tip="singleWithLevels or hybridWithTokens">
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

					{/* Reasoning Level */}
					{formData.reasoningSupport && formData.reasoningType === ReasoningType.SingleWithLevels && (
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Reasoning Level</span>
								<span className="label-text-alt tooltip tooltip-right" data-tip="Pick how deep the chain-of-thought is">
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

					{/* Reasoning Tokens */}
					{formData.reasoningSupport && formData.reasoningType === ReasoningType.HybridWithTokens && (
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Reasoning Tokens</span>
								<span
									className="label-text-alt tooltip tooltip-right"
									data-tip="Number of tokens dedicated to reasoning"
								>
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
									placeholder="e.g. 1024"
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
								className="label-text-alt tooltip tooltip-right"
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
							<span className="text-sm">Timeout (seconds)</span>
							<span className="tooltip tooltip-right" data-tip="Maximum time to wait for response">
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
							<span className="label-text-alt tooltip tooltip-right" data-tip="Maximum tokens for input">
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
							<span className="label-text-alt tooltip tooltip-right" data-tip="Maximum tokens for output">
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
							<span
								className="label-text-alt tooltip tooltip-right"
								data-tip="Instructions that define the model's behavior"
							>
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
							{isEditMode ? 'Save Changes' : 'Add Preset'}
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default ModifyModelModal;
