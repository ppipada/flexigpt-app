import React, { type FC, useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import type { ChatOptions } from '@/models/aimodelmodel';

interface AdvancedParamsModalProps {
	isOpen: boolean;
	onClose: () => void;
	currentModel: ChatOptions;
	onSave: (updatedModel: ChatOptions) => void;
}

// A modal that allows editing advanced parameters:
// streaming, maxPromptLength, maxOutputLength, systemPrompt.
// It does simple numeric validations and then calls `onSave` with the updated model.
const AdvancedParamsModal: FC<AdvancedParamsModalProps> = ({ isOpen, onClose, currentModel, onSave }) => {
	// We store these fields locally as strings (for numeric fields) to allow easy blank entry.
	const [stream, setStream] = useState<boolean>(false);
	const [maxPromptLength, setMaxPromptLength] = useState<string>('');
	const [maxOutputLength, setMaxOutputLength] = useState<string>('');
	const [systemPrompt, setSystemPrompt] = useState<string>('');

	// For tracking/validating errors:
	const [errors, setErrors] = useState<{
		maxPromptLength?: string;
		maxOutputLength?: string;
	}>({});

	// On modal open, initialize local state with the currentModel fields.
	useEffect(() => {
		if (isOpen) {
			setStream(currentModel.stream);
			setMaxPromptLength(String(currentModel.maxPromptLength));
			setMaxOutputLength(String(currentModel.maxOutputLength));
			setSystemPrompt(currentModel.systemPrompt);
			setErrors({});
		}
	}, [isOpen]);

	type ValidationField = 'maxPromptLength' | 'maxOutputLength';

	type ValidationErrors = Partial<Record<ValidationField, string>>;

	// Validate a single field name, storing the result in `errors` state.
	const validateField = (field: ValidationField, value: string) => {
		const newErrors: ValidationErrors = Object.fromEntries(Object.entries(errors).filter(([key]) => key !== field));

		if (value.trim() !== '') {
			const numValue = Number(value.trim());
			if (Number.isNaN(numValue) || numValue <= 0) {
				newErrors[field] = `${field} must be a positive number.`;
			}
		}

		setErrors(newErrors);
	};

	const isFormValid = useMemo(() => {
		return !Object.values(errors).some(Boolean);
	}, [errors]);

	const handleSave = (e: React.FormEvent) => {
		e.preventDefault();

		// Re-validate numeric fields.
		validateField('maxPromptLength', maxPromptLength);
		validateField('maxOutputLength', maxOutputLength);

		if (!isFormValid) {
			return; // there's an error, do not proceed.
		}

		// Parse numeric fields; if empty, let them remain as the existing model's values.
		const parsedMaxPromptLength =
			maxPromptLength.trim() === '' ? currentModel.maxPromptLength : Number(maxPromptLength.trim());
		const parsedMaxOutputLength =
			maxOutputLength.trim() === '' ? currentModel.maxOutputLength : Number(maxOutputLength.trim());

		// Build the updated model.
		const updatedModel: ChatOptions = {
			...currentModel,
			stream,
			maxPromptLength: parsedMaxPromptLength,
			maxOutputLength: parsedMaxOutputLength,
			systemPrompt,
		};

		onSave(updatedModel);
	};

	if (!isOpen) {
		return null;
	}

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* Header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">Advanced Model Parameters</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close" title="Close">
						<FiX size={12} />
					</button>
				</div>

				{/* Form for advanced fields */}
				<form onSubmit={handleSave} className="space-y-4">
					{/* Toggle: stream */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-4 cursor-pointer">
							<span className="label-text text-sm">Streaming</span>
							<span className="label-text-alt tooltip" data-tip="Stream data continuously.">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-8">
							<input
								type="checkbox"
								checked={stream}
								onChange={e => {
									setStream(e.target.checked);
								}}
								className="toggle toggle-accent rounded-full"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Max Prompt Length */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-4">
							<span className="label-text text-sm">Max Prompt Tokens</span>
							<span className="label-text-alt tooltip" data-tip="Maximum tokens for input prompt">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-8">
							<input
								type="text"
								value={maxPromptLength}
								onChange={e => {
									setMaxPromptLength(e.target.value);
									validateField('maxPromptLength', e.target.value);
								}}
								className={`input input-bordered w-full rounded-xl ${errors.maxPromptLength ? 'input-error' : ''}`}
								placeholder={`Default: ${currentModel.maxPromptLength}`}
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
						<label className="label col-span-4">
							<span className="label-text text-sm">Max Output Tokens</span>
							<span className="label-text-alt tooltip" data-tip="Maximum tokens for model output">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-8">
							<input
								type="text"
								value={maxOutputLength}
								onChange={e => {
									setMaxOutputLength(e.target.value);
									validateField('maxOutputLength', e.target.value);
								}}
								className={`input input-bordered w-full rounded-xl ${errors.maxOutputLength ? 'input-error' : ''}`}
								placeholder={`Default: ${currentModel.maxOutputLength}`}
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
						<label className="label col-span-4">
							<span className="label-text text-sm">System Prompt</span>
							<span className="label-text-alt tooltip" data-tip="Behavior instructions.">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-8">
							<textarea
								value={systemPrompt}
								onChange={e => {
									setSystemPrompt(e.target.value);
								}}
								className="textarea textarea-bordered w-full rounded-xl h-24"
								placeholder="Enter system prompt instructions here..."
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Action buttons */}
					<div className="modal-action">
						<button type="button" className="btn rounded-xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-primary rounded-xl" disabled={!isFormValid}>
							Save
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default AdvancedParamsModal;
