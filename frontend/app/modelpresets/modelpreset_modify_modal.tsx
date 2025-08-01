import React, { type FC, useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import {
	type ModelPreset,
	type ModelPresetID,
	type ProviderName,
	ReasoningLevel,
	ReasoningType,
} from '@/spec/modelpreset';

import Dropdown from '@/components/dropdown';

const reasoningTypeItems: Record<ReasoningType, { isEnabled: boolean; displayName: string }> = {
	[ReasoningType.SingleWithLevels]: {
		isEnabled: true,
		displayName: 'Reasoning only, with Levels',
	},
	[ReasoningType.HybridWithTokens]: {
		isEnabled: true,
		displayName: 'Hybrid, with Reasoning Tokens',
	},
};

const reasoningLevelItems: Record<ReasoningLevel, { isEnabled: boolean; displayName: string }> = {
	[ReasoningLevel.Low]: { isEnabled: true, displayName: 'Low' },
	[ReasoningLevel.Medium]: { isEnabled: true, displayName: 'Medium' },
	[ReasoningLevel.High]: { isEnabled: true, displayName: 'High' },
};

/** Defaults we want to *apply* in **Add** mode (placeholders use it too). */
const AddModeDefaults: ModelPreset = {
	id: '',
	name: '',
	slug: '',
	displayName: '',
	isEnabled: true,
	isBuiltIn: false,
	stream: true,
	maxPromptLength: 2048,
	maxOutputLength: 1024,
	temperature: 0.1,
	reasoning: undefined,
	systemPrompt: '',
	timeout: 60,
	additionalParametersRawJSON: undefined,
};

interface ModifyModelModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (modelPresetID: ModelPresetID, modelData: ModelPreset) => void;
	providerName: ProviderName;

	initialModelID?: ModelPresetID; // ⇒ edit mode when truthy
	initialData?: ModelPreset;
	existingModels: Record<ModelPresetID, ModelPreset>;
}

interface ModelPresetFormData {
	presetLabel: string;
	name: string;
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
	initialModelID,
	initialData,
	existingModels,
}) => {
	const isEditMode = Boolean(initialModelID);

	const [defaultValues] = useState<ModelPreset>(AddModeDefaults);
	const [modelPresetID, setModelPresetID] = useState<ModelPresetID>(initialModelID ?? ('' as ModelPresetID));

	const blankForm: ModelPresetFormData = {
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
	};

	const [formData, setFormData] = useState<ModelPresetFormData>(blankForm);

	useEffect(() => {
		if (!isOpen) return;

		const loadData = () => {
			if (isEditMode && initialData) {
				/* ---------- EDIT mode: copy ONLY what exists ---------------- */
				setFormData({
					presetLabel: initialData.displayName,
					name: initialData.name,
					isEnabled: initialData.isEnabled,
					stream: initialData.stream ?? false,
					maxPromptLength: initialData.maxPromptLength !== undefined ? String(initialData.maxPromptLength) : '',
					maxOutputLength: initialData.maxOutputLength !== undefined ? String(initialData.maxOutputLength) : '',
					temperature: initialData.temperature !== undefined ? String(initialData.temperature) : '',
					reasoningSupport: !!initialData.reasoning,
					reasoningType: initialData.reasoning?.type ?? ReasoningType.SingleWithLevels,
					reasoningLevel: initialData.reasoning?.level ?? ReasoningLevel.Medium,
					reasoningTokens: initialData.reasoning?.tokens !== undefined ? String(initialData.reasoning.tokens) : '',
					systemPrompt: initialData.systemPrompt ?? '',
					timeout: initialData.timeout !== undefined ? String(initialData.timeout) : '',
				});
				setModelPresetID(initialData.id);
			} else {
				/* ---------- ADD mode: pre-fill with defaults ---------------- */
				setFormData({
					presetLabel: '',
					name: '',
					isEnabled: true,
					stream: AddModeDefaults.stream ?? false,
					maxPromptLength: String(AddModeDefaults.maxPromptLength ?? ''),
					maxOutputLength: String(AddModeDefaults.maxOutputLength ?? ''),
					temperature: String(AddModeDefaults.temperature ?? ''),
					reasoningSupport: false,
					reasoningType: ReasoningType.SingleWithLevels,
					reasoningLevel: ReasoningLevel.Medium,
					reasoningTokens: '',
					systemPrompt: '',
					timeout: String(AddModeDefaults.timeout ?? ''),
				});
				setModelPresetID('' as ModelPresetID);
			}

			setErrors({});
		};

		loadData();
	}, [isOpen]); // keep deps minimal to avoid needless reloads

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

	const [errors, setErrors] = useState<ValidationErrors>({});

	const calcNumericError = (
		field: ValidationField,
		strVal: string,
		minOrRange?: { min?: number; max?: number }
	): string | undefined => {
		if (strVal.trim() === '') return;
		const num = Number(strVal);
		if (Number.isNaN(num)) return `${field} must be a valid number.`;
		if (minOrRange?.min !== undefined && num < minOrRange.min) return `${field} must be ≥ ${minOrRange.min}.`;
		if (minOrRange?.max !== undefined && num > minOrRange.max) return `${field} must be ≤ ${minOrRange.max}.`;
		return;
	};

	const runValidation = (): ValidationErrors => {
		const v: ValidationErrors = {};

		/* id ----------------------------------------------------------- */
		if (!isEditMode) {
			const idTrim = modelPresetID.trim();
			if (!idTrim) v.modelPresetID = 'Model Preset ID is required.';
			else if (!/^[a-zA-Z0-9-]+$/.test(idTrim)) v.modelPresetID = 'Only letters, numbers, and hyphens allowed.';
			else if (Object.prototype.hasOwnProperty.call(existingModels, idTrim))
				v.modelPresetID = 'Model Preset ID must be unique.';
		}

		/* required strings -------------------------------------------- */
		if (!formData.name.trim()) v.name = 'Model Name is required.';
		if (!formData.presetLabel.trim()) v.presetLabel = 'Preset Label is required.';

		/* numeric fields ---------------------------------------------- */
		const maybe = (f: ValidationField, rng?: { min?: number; max?: number }) => {
			let value: string;
			if (f === 'modelPresetID') {
				value = modelPresetID;
			} else {
				value = formData[f as keyof ModelPresetFormData] as string;
			}
			v[f] = calcNumericError(f, value, rng);
		};
		maybe('temperature', { min: 0, max: 1 });
		maybe('maxPromptLength', { min: 1 });
		maybe('maxOutputLength', { min: 1 });
		maybe('timeout', { min: 1 });

		if (formData.reasoningSupport && formData.reasoningType === ReasoningType.HybridWithTokens) {
			maybe('reasoningTokens', { min: 1024 });
		}
		const newV: ValidationErrors = Object.fromEntries(
			Object.entries(v).filter(([key]) => v[key as ValidationField] !== undefined)
		);

		return newV;
	};

	/* helper - string ⇒ number (keep default if blank / NaN) ----------- */
	const parseOrDefault = (val: string, def: number) => {
		const n = Number(val);
		return val.trim() === '' || Number.isNaN(n) ? def : n;
	};

	const handleFieldChange = (field: keyof ModelPresetFormData, value: string | boolean) => {
		setFormData(prev => ({ ...prev, [field]: value }));
	};

	/* unified change handler for inputs ------------------------------- */
	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;

		if (name === 'modelPresetID') {
			setModelPresetID(value);
			return;
		}

		if (type === 'checkbox') {
			handleFieldChange(name as any, checked);
		} else {
			handleFieldChange(name as any, value);
		}
	};

	const isAllValid = useMemo(() => {
		const currentErrors = runValidation();
		return Object.keys(currentErrors).length === 0;
	}, [formData, modelPresetID]);

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		const newErrors = runValidation();
		setErrors(newErrors);

		if (Object.keys(newErrors).length !== 0) return;

		const finalModelPresetID = modelPresetID.trim();

		const finalData: ModelPreset = {
			id: finalModelPresetID,
			slug: finalModelPresetID,
			isBuiltIn: false,

			name: formData.name.trim(),
			displayName: formData.presetLabel.trim(),
			isEnabled: formData.isEnabled,
			stream: formData.stream,

			maxPromptLength: parseOrDefault(formData.maxPromptLength, defaultValues.maxPromptLength ?? 2048),
			maxOutputLength: parseOrDefault(formData.maxOutputLength, defaultValues.maxOutputLength ?? 1024),
			temperature: parseOrDefault(formData.temperature, defaultValues.temperature ?? 0.1),
			timeout: parseOrDefault(formData.timeout, defaultValues.timeout ?? 60),

			systemPrompt: formData.systemPrompt,
			// additionalParametersRawJSON intentionally omitted unless you add a field
		};

		if (formData.reasoningSupport) {
			finalData.reasoning = {
				type: formData.reasoningType ?? ReasoningType.SingleWithLevels,
				level: formData.reasoningLevel ?? ReasoningLevel.Medium,
				tokens: parseOrDefault(formData.reasoningTokens ?? '', defaultValues.reasoning?.tokens ?? 0),
			};
		}

		onSubmit(finalModelPresetID, finalData);
	};

	if (!isOpen) return null;

	const numPlaceholder = (field: keyof ModelPreset) => {
		const v = defaultValues[field];
		return v === undefined || typeof v === 'object' ? 'Default: N/A' : `Default: ${String(v)}`;
	};

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* Header -------------------------------------------------- */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{isEditMode ? 'Edit Model Preset' : 'Add Model Preset'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close" title="Close">
						<FiX size={12} />
					</button>
				</div>

				{/* FORM ---------------------------------------------------- */}
				<form onSubmit={handleSubmit} className="space-y-4">
					{/* ModelPresetID ------------------------------------- */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Model Preset ID*</span>
							<span
								className="label-text-alt tooltip tooltip-right"
								data-tip="Unique identifier. Letters, numbers, hyphen."
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								name="modelPresetID"
								type="text"
								className={`input input-bordered w-full rounded-2xl ${errors.modelPresetID ? 'input-error' : ''}`}
								value={modelPresetID}
								onChange={handleChange}
								disabled={isEditMode}
								placeholder="e.g. gpt4-preset"
								autoComplete="off"
								spellCheck="false"
							/>
							{errors.modelPresetID && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} />
										{errors.modelPresetID}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Model Name ---------------------------------------- */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Model Name*</span>
							<span
								className="label-text-alt tooltip tooltip-right"
								data-tip="The name you send to the completions API (e.g. gpt-4)"
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								name="name"
								type="text"
								className={`input input-bordered w-full rounded-2xl ${errors.name ? 'input-error' : ''}`}
								value={formData.name}
								onChange={handleChange}
								placeholder="e.g. gpt-4, claude-3-opus-20240229"
								autoComplete="off"
								spellCheck="false"
							/>
							{errors.name && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} />
										{errors.name}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Preset Label -------------------------------------- */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Preset Label*</span>
							<span className="label-text-alt tooltip tooltip-right" data-tip="Friendly name shown in the UI">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								name="presetLabel"
								type="text"
								className={`input input-bordered w-full rounded-2xl ${errors.presetLabel ? 'input-error' : ''}`}
								value={formData.presetLabel}
								onChange={handleChange}
								placeholder="e.g. GPT-4 (Creative)"
								autoComplete="off"
								spellCheck="false"
							/>
							{errors.presetLabel && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} />
										{errors.presetLabel}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Toggles: isEnabled / stream / reasoningSupport ---- */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3 cursor-pointer">
							<span className="label-text text-sm">Enabled</span>
						</label>
						<div className="col-span-9">
							<input
								type="checkbox"
								name="isEnabled"
								className="toggle toggle-accent rounded-full"
								checked={formData.isEnabled}
								onChange={handleChange}
							/>
						</div>
					</div>

					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3 cursor-pointer">
							<span className="label-text text-sm">Streaming</span>
						</label>
						<div className="col-span-9">
							<input
								type="checkbox"
								name="stream"
								className="toggle toggle-accent rounded-full"
								checked={formData.stream}
								onChange={handleChange}
							/>
						</div>
					</div>

					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3 cursor-pointer">
							<span className="label-text text-sm">Supports Reasoning</span>
							<span className="label-text-alt tooltip tooltip-right" data-tip="If enabled, configure below">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="checkbox"
								name="reasoningSupport"
								className="toggle toggle-accent rounded-full"
								checked={formData.reasoningSupport}
								onChange={handleChange}
							/>
						</div>
					</div>

					{/* Reasoning controls -------------------------------- */}
					{formData.reasoningSupport && (
						<>
							{/* type */}
							<div className="grid grid-cols-12 items-center gap-2">
								<label className="label col-span-3">
									<span className="label-text text-sm">Reasoning Type</span>
									<span className="label-text-alt tooltip tooltip-right">
										<FiHelpCircle size={12} />
									</span>
								</label>
								<div className="col-span-9">
									<Dropdown<ReasoningType>
										dropdownItems={reasoningTypeItems}
										selectedKey={formData.reasoningType ?? ReasoningType.SingleWithLevels}
										onChange={t => {
											setFormData(p => ({ ...p, reasoningType: t }));
										}}
										filterDisabled={false}
										title="Select Reasoning Type"
										getDisplayName={k => reasoningTypeItems[k].displayName}
									/>
								</div>
							</div>

							{/* level (for SingleWithLevels) */}
							{formData.reasoningType === ReasoningType.SingleWithLevels && (
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Reasoning Level</span>
										<span className="label-text-alt tooltip tooltip-right">
											<FiHelpCircle size={12} />
										</span>
									</label>
									<div className="col-span-9">
										<Dropdown<ReasoningLevel>
											dropdownItems={reasoningLevelItems}
											selectedKey={formData.reasoningLevel ?? ReasoningLevel.Medium}
											onChange={lvl => {
												setFormData(p => ({ ...p, reasoningLevel: lvl }));
											}}
											filterDisabled={false}
											title="Select Reasoning Level"
											getDisplayName={k => reasoningLevelItems[k].displayName}
										/>
									</div>
								</div>
							)}

							{/* tokens (for HybridWithTokens) */}
							{formData.reasoningType === ReasoningType.HybridWithTokens && (
								<div className="grid grid-cols-12 items-center gap-2">
									<label className="label col-span-3">
										<span className="label-text text-sm">Reasoning Tokens</span>
										<span className="label-text-alt tooltip tooltip-right">
											<FiHelpCircle size={12} />
										</span>
									</label>
									<div className="col-span-9">
										<input
											name="reasoningTokens"
											type="text"
											className={`input input-bordered w-full rounded-2xl ${
												errors.reasoningTokens ? 'input-error' : ''
											}`}
											value={formData.reasoningTokens}
											onChange={handleChange}
											placeholder="e.g. 1024"
											spellCheck="false"
										/>
										{errors.reasoningTokens && (
											<div className="label">
												<span className="label-text-alt text-error flex items-center gap-1">
													<FiAlertCircle size={12} />
													{errors.reasoningTokens}
												</span>
											</div>
										)}
									</div>
								</div>
							)}
						</>
					)}

					{/* Temperature -------------------------------------- */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Temperature (0-1)</span>
							<span className="label-text-alt tooltip tooltip-right">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								name="temperature"
								type="text"
								className={`input input-bordered w-full rounded-2xl ${errors.temperature ? 'input-error' : ''}`}
								value={formData.temperature}
								onChange={handleChange}
								placeholder={numPlaceholder('temperature')}
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

					{/* Timeout ------------------------------------------ */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Timeout (s)</span>
							<span className="label-text-alt tooltip tooltip-right">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								name="timeout"
								type="text"
								className={`input input-bordered w-full rounded-2xl ${errors.timeout ? 'input-error' : ''}`}
								value={formData.timeout}
								onChange={handleChange}
								placeholder={numPlaceholder('timeout')}
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

					{/* Prompt / Output lengths -------------------------- */}
					{(['maxPromptLength', 'maxOutputLength'] as const).map(f => (
						<div className="grid grid-cols-12 items-center gap-2" key={f}>
							<label className="label col-span-3">
								<span className="label-text text-sm">
									{f === 'maxPromptLength' ? 'Max Prompt Tokens' : 'Max Output Tokens'}
								</span>
								<span className="label-text-alt tooltip tooltip-right">
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<input
									name={f}
									type="text"
									className={`input input-bordered w-full rounded-2xl ${errors[f] ? 'input-error' : ''}`}
									value={formData[f]}
									onChange={handleChange}
									placeholder={numPlaceholder(f)}
									spellCheck="false"
								/>
								{errors[f] && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors[f]}
										</span>
									</div>
								)}
							</div>
						</div>
					))}

					{/* System Prompt ------------------------------------ */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">System Prompt</span>
							<span className="label-text-alt tooltip tooltip-right">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<textarea
								name="systemPrompt"
								className="textarea textarea-bordered w-full rounded-2xl h-24"
								value={formData.systemPrompt}
								onChange={handleChange}
								placeholder="Enter instructions here…"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Actions ------------------------------------------ */}
					<div className="modal-action">
						<button type="button" className="btn rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-primary rounded-2xl" disabled={!isAllValid}>
							{isEditMode ? 'Save Changes' : 'Add Preset'}
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default ModifyModelModal;
