import { type FC, useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiPlus, FiX } from 'react-icons/fi';

import { type ProviderName } from '@/models/aiprovidermodel';
import type { AISetting, ModelSetting } from '@/models/settingmodel';

import ModifyModelModal from '@/settings/model_modify_modal';

interface AddProviderModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (providerName: ProviderName, newSettings: AISetting) => void;
	/**
	 * List of existing provider names in your system.
	 * Used to prevent duplicate providerName.
	 */
	existingProviderNames: string[];
}

/** Local interface for our form data, to match the style of ModifyModelModal. */
interface ProviderFormData {
	providerName: string;
	apiKey: string;
	origin: string;
	chatCompletionPathPrefix: string;
	/** Must not be blank—once set by the child modal, user can proceed. */
	defaultModelName: string;
}

const AddProviderModal: FC<AddProviderModalProps> = ({ isOpen, onClose, onSubmit, existingProviderNames }) => {
	const [formData, setFormData] = useState<ProviderFormData>({
		providerName: '',
		apiKey: '',
		origin: '',
		chatCompletionPathPrefix: '',
		defaultModelName: '',
	});

	// Model settings for the single default model
	const [modelSettings, setModelSettings] = useState<Record<string, ModelSetting>>({});
	const [isModifyModelModalOpen, setIsModifyModelModalOpen] = useState(false);

	// Mirroring the error handling style in ModifyModelModal
	const [errors, setErrors] = useState<{
		providerName?: string;
		apiKey?: string;
		origin?: string;
		chatCompletionPathPrefix?: string;
		defaultModelName?: string;
	}>({});

	// Reset all states when the modal closes
	useEffect(() => {
		if (!isOpen) {
			setFormData({
				providerName: '',
				apiKey: '',
				origin: '',
				chatCompletionPathPrefix: '',
				defaultModelName: '',
			});
			setModelSettings({});
			setIsModifyModelModalOpen(false);
			setErrors({});
		}
	}, [isOpen]);

	// We define the list of fields to validate
	type ValidationField = 'providerName' | 'apiKey' | 'origin' | 'chatCompletionPathPrefix' | 'defaultModelName';

	type ValidationErrors = Partial<Record<ValidationField, string>>;

	/**
	 * Field-level validation: run these checks whenever a user changes a field,
	 * populating/clearing error messages as needed.
	 */
	const validateField = (field: keyof ProviderFormData, value: string) => {
		// Copy existing errors minus any error tied to this field
		const newErrors: ValidationErrors = Object.fromEntries(Object.entries(errors).filter(([key]) => key !== field));

		if (field === 'providerName') {
			if (!value.trim()) {
				newErrors.providerName = 'Provider name is required.';
			} else if (value.includes('.')) {
				newErrors.providerName = 'Provider name cannot contain a dot (".").';
			} else if (value.includes(' ')) {
				newErrors.providerName = 'Provider name cannot contain spaces.';
			} else if (existingProviderNames.includes(value)) {
				newErrors.providerName = 'This provider name already exists.';
			}
		}

		if (field === 'apiKey') {
			if (!value.trim()) {
				newErrors.apiKey = 'API key is required.';
			}
		}

		if (field === 'origin') {
			if (!value.trim()) {
				newErrors.origin = 'Origin is required.';
			} else {
				try {
					// Attempt parsing as URL
					new URL(value);
				} catch {
					newErrors.origin = 'Origin must be a valid URL.';
				}
			}
		}

		// If you want chatCompletionPathPrefix to be mandatory, uncomment:
		// if (field === 'chatCompletionPathPrefix') {
		//   if (!value.trim()) {
		//     newErrors.chatCompletionPathPrefix = 'Chat path prefix is required.';
		//   }
		// }

		if (field === 'defaultModelName') {
			if (!value.trim()) {
				newErrors.defaultModelName = 'Please configure at least one default model.';
			}
		}

		setErrors(newErrors);
	};

	/**
	 * Keep track of field changes and run validation as the user types.
	 */
	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const { name, value } = e.target;
		setFormData(prev => ({ ...prev, [name]: value }));
		validateField(name as keyof ProviderFormData, value);
	};

	/**
	 * Show the ModifyModelModal to allow user to define the default model.
	 */
	const handleAddDefaultModel = () => {
		setIsModifyModelModalOpen(true);
	};

	/**
	 * Called when user finishes configuring the default model in ModifyModelModal.
	 */
	const handleModifyModelSubmit = (modelName: string, modelData: ModelSetting) => {
		setFormData(prev => ({ ...prev, defaultModelName: modelName }));
		setModelSettings({ [modelName]: modelData });
		setIsModifyModelModalOpen(false);
	};

	/**
	 * Disable/enable the "Add Provider" button based on whether the form is valid.
	 *
	 * - No validation errors.
	 * - All required fields are non-empty.
	 */
	const isAllValid = useMemo(() => {
		const hasNoErrors = !Object.values(errors).some(Boolean);
		const hasRequiredFields =
			!!formData.providerName.trim() &&
			!!formData.apiKey.trim() &&
			!!formData.origin.trim() &&
			!!formData.defaultModelName.trim();
		return hasNoErrors && hasRequiredFields;
	}, [errors, formData]);

	/**
	 * Final form submission. We do a last validation pass to ensure all fields are valid.
	 */
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		// Validate each field one last time:
		validateField('providerName', formData.providerName);
		validateField('apiKey', formData.apiKey);
		validateField('origin', formData.origin);
		validateField('chatCompletionPathPrefix', formData.chatCompletionPathPrefix);
		validateField('defaultModelName', formData.defaultModelName);

		// After final validation, check for any errors
		const hasAnyErrors = Object.values(errors).some(Boolean);
		if (hasAnyErrors) return;

		// Also confirm that required fields are filled (defensive check)
		if (!isAllValid) return;

		// Construct the new provider settings
		const newProviderSettings: AISetting = {
			isEnabled: true,
			apiKey: formData.apiKey,
			origin: formData.origin,
			chatCompletionPathPrefix: formData.chatCompletionPathPrefix,
			defaultModel: formData.defaultModelName,
			modelSettings,
		};

		onSubmit(formData.providerName, newProviderSettings);
		onClose();
	};

	// If not open, simply return null (no rendering).
	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* Header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">Add New Provider</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close">
						<FiX size={12} />
					</button>
				</div>

				{/* Form Body */}
				<form onSubmit={handleSubmit} className="space-y-4">
					{/* Provider Name */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Provider Name*</span>
							<span
								className="label-text-alt tooltip"
								data-tip="Unique identifier, no dots or spaces, not already in use"
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="providerName"
								className={`input input-bordered w-full rounded-xl ${errors.providerName ? 'input-error' : ''}`}
								value={formData.providerName}
								onChange={handleChange}
								placeholder="e.g. openai2"
								spellCheck="false"
							/>
							{errors.providerName && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.providerName}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* API Key */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">API Key*</span>
							<span className="label-text-alt tooltip" data-tip="Your provider's API key">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="password"
								name="apiKey"
								className={`input input-bordered w-full rounded-xl ${errors.apiKey ? 'input-error' : ''}`}
								value={formData.apiKey}
								onChange={handleChange}
								placeholder="Your provider's API key"
								spellCheck="false"
							/>
							{errors.apiKey && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.apiKey}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Origin */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Origin/FQDN*</span>
							<span
								className="label-text-alt tooltip"
								data-tip="Base URL for API requests (e.g. https://api.provider.com)"
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="origin"
								className={`input input-bordered w-full rounded-xl ${errors.origin ? 'input-error' : ''}`}
								value={formData.origin}
								onChange={handleChange}
								placeholder="e.g. https://api.openai.com"
								spellCheck="false"
							/>
							{errors.origin && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.origin}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Chat completion path prefix */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Chat Path Prefix</span>
							<span className="label-text-alt tooltip" data-tip="Path for chat completions (e.g. /v1/chat/completions)">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="chatCompletionPathPrefix"
								className={`input input-bordered w-full rounded-xl ${
									errors.chatCompletionPathPrefix ? 'input-error' : ''
								}`}
								value={formData.chatCompletionPathPrefix}
								onChange={handleChange}
								placeholder="e.g. /v1/chat/completions"
								spellCheck="false"
							/>
							{errors.chatCompletionPathPrefix && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.chatCompletionPathPrefix}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Default Model */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Default Model*</span>
							<span className="label-text-alt tooltip" data-tip="Configure or edit a default model for this provider">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9 flex flex-col space-y-2">
							<div className="flex items-center justify-between">
								{formData.defaultModelName ? (
									<p className="text-sm">
										Configured: <strong>{formData.defaultModelName}</strong>
									</p>
								) : (
									<p className="text-sm italic">No default model configured</p>
								)}
								<button
									type="button"
									className="btn btn-md btn-ghost flex items-center rounded-xl"
									onClick={handleAddDefaultModel}
								>
									<FiPlus size={16} />
									<span className="ml-1">Configure Default Model</span>
								</button>
							</div>
							{errors.defaultModelName && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.defaultModelName}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Actions */}
					<div className="modal-action mt-6">
						<button type="button" className="btn rounded-xl" onClick={onClose}>
							Cancel
						</button>
						<button
							type="submit"
							className="btn btn-primary rounded-xl"
							disabled={!isAllValid}
							/* <--- Disable if form not valid */
						>
							Add Provider
						</button>
					</div>
				</form>
			</div>

			{/* Nested ModifyModelModal to define the default model */}
			{isModifyModelModalOpen && (
				<ModifyModelModal
					isOpen={isModifyModelModalOpen}
					onClose={() => {
						setIsModifyModelModalOpen(false);
					}}
					onSubmit={handleModifyModelSubmit}
					providerName={formData.providerName}
					initialModelName={formData.defaultModelName || undefined}
					initialData={formData.defaultModelName ? modelSettings[formData.defaultModelName] : undefined}
					existingModels={modelSettings}
				/>
			)}
		</dialog>
	);
};

export default AddProviderModal;
