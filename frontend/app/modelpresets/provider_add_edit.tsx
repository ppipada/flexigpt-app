import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import { ProviderAPIType, type ProviderName, type ProviderPreset } from '@/spec/modelpreset';

import { omitManyKeys } from '@/lib/obj_utils';
import { isValidUrl } from '@/lib/text_utils';

type FormData = {
	providerName: string;
	displayName: string;
	apiType: ProviderAPIType;
	isEnabled: boolean;
	origin: string;
	chatCompletionPathPrefix: string;
	apiKeyHeaderKey: string;
	defaultHeadersRawJSON: string;
	apiKey: string;
};

const DEFAULT_FORM: FormData = {
	providerName: '',
	displayName: '',
	apiType: ProviderAPIType.OpenAICompatible,
	isEnabled: true,
	origin: '',
	chatCompletionPathPrefix: '/v1/chat/completions',
	apiKeyHeaderKey: 'Authorization',
	defaultHeadersRawJSON: '',
	apiKey: '',
};

interface Props {
	isOpen: boolean;
	mode: 'add' | 'edit';
	onClose: () => void;
	onSubmit: (
		providerName: ProviderName,
		payload: Omit<ProviderPreset, 'isBuiltIn' | 'defaultModelPresetID' | 'modelPresets'>,
		apiKey: string | null
	) => void;
	existingProviderNames: ProviderName[];
	initialPreset?: ProviderPreset; // for edit
	apiKeyAlreadySet?: boolean; // for edit
}

const AddEditProviderPresetModal: FC<Props> = ({
	isOpen,
	mode,
	onClose,
	onSubmit,
	existingProviderNames,
	initialPreset,
	apiKeyAlreadySet = false,
}) => {
	const [formData, setFormData] = useState<FormData>(DEFAULT_FORM);
	const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});

	useEffect(() => {
		if (!isOpen) return;

		if (mode === 'edit' && initialPreset) {
			setFormData({
				providerName: initialPreset.name,
				displayName: initialPreset.displayName,
				apiType: initialPreset.apiType,
				isEnabled: initialPreset.isEnabled,
				origin: initialPreset.origin,
				chatCompletionPathPrefix: initialPreset.chatCompletionPathPrefix,
				apiKeyHeaderKey: initialPreset.apiKeyHeaderKey,
				defaultHeadersRawJSON: JSON.stringify(initialPreset.defaultHeaders, null, 2),
				apiKey: '',
			});
		} else {
			setFormData(DEFAULT_FORM);
		}
		setErrors({});
	}, [isOpen, mode, initialPreset]);

	const validateField = (field: keyof FormData, val: string) => {
		let newErrs = { ...errors };

		const v = val.trim();

		if (field === 'providerName' && mode === 'add') {
			if (!v) newErrs.providerName = 'Provider name required.';
			else if (!/^[\w-]+$/.test(v)) newErrs.providerName = 'Letters, numbers, dash & underscore only.';
			else if (existingProviderNames.includes(v)) newErrs.providerName = 'Provider already exists.';
			else newErrs = omitManyKeys(newErrs, ['providerName']);
		}

		if (field === 'displayName') {
			if (!v) newErrs.displayName = 'Display name required.';
			else newErrs = omitManyKeys(newErrs, ['displayName']);
		}

		if (field === 'origin') {
			if (!isValidUrl(v)) newErrs.origin = 'Must be a valid URL.';
			else newErrs = omitManyKeys(newErrs, ['origin']);
		}

		if (field === 'defaultHeadersRawJSON') {
			if (v) {
				try {
					JSON.parse(v);
					newErrs = omitManyKeys(newErrs, ['defaultHeadersRawJSON']);
				} catch {
					newErrs.defaultHeadersRawJSON = 'Invalid JSON.';
				}
			} else newErrs = omitManyKeys(newErrs, ['defaultHeadersRawJSON']);
		}

		if (field === 'apiKey' && mode === 'add') {
			if (!v) newErrs.apiKey = 'API key required.';
			else newErrs = omitManyKeys(newErrs, ['apiKey']);
		}

		setErrors(newErrs);
	};

	const handleInput = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;
		const newVal = type === 'checkbox' ? checked : value;

		setFormData(prev => ({ ...prev, [name]: newVal }));

		if (['providerName', 'displayName', 'origin', 'defaultHeadersRawJSON', 'apiKey'].includes(name)) {
			validateField(name as keyof FormData, String(newVal));
		}
	};

	const allValid = useMemo(() => {
		const hasErr = Object.values(errors).some(Boolean);
		const requiredFilled =
			formData.providerName.trim() &&
			formData.displayName.trim() &&
			formData.origin.trim() &&
			(mode === 'add' ? formData.apiKey.trim() : true);

		return !hasErr && requiredFilled;
	}, [errors, formData, mode]);

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		(Object.entries(formData) as [keyof FormData, string][]).forEach(([k, v]) => {
			if (typeof v === 'string') validateField(k, v);
		});
		if (!allValid) return;

		let defaultHeaders: Record<string, string> = {};
		if (formData.defaultHeadersRawJSON.trim()) {
			try {
				defaultHeaders = JSON.parse(formData.defaultHeadersRawJSON.trim());
			} catch {
				/* already flagged */
			}
		}

		const payload = {
			name: formData.providerName.trim(),
			displayName: formData.displayName.trim(),
			apiType: formData.apiType,
			isEnabled: formData.isEnabled,
			origin: formData.origin.trim(),
			chatCompletionPathPrefix: formData.chatCompletionPathPrefix.trim(),
			apiKeyHeaderKey: formData.apiKeyHeaderKey.trim(),
			defaultHeaders,
		};

		onSubmit(formData.providerName.trim(), payload, formData.apiKey.trim() || null);
		onClose();
	};

	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{mode === 'add' ? 'Add Provider' : 'Edit Provider'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close" title="Close">
						<FiX size={12} />
					</button>
				</div>

				<form onSubmit={handleSubmit} className="space-y-4">
					{/* Provider ID */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Provider ID*</span>
							{mode === 'add' && (
								<span
									className="label-text-alt tooltip tooltip-right"
									data-tip="Unique identifier (letters, numbers, dash, underscore)."
								>
									<FiHelpCircle size={12} />
								</span>
							)}
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="providerName"
								value={formData.providerName}
								onChange={handleInput}
								className={`input input-bordered w-full rounded-2xl ${errors.providerName ? 'input-error' : ''}`}
								disabled={mode === 'edit'}
								spellCheck="false"
								autoComplete="off"
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

					{/* Display Name */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Display Name*</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="displayName"
								value={formData.displayName}
								onChange={handleInput}
								className={`input input-bordered w-full rounded-2xl ${errors.displayName ? 'input-error' : ''}`}
								spellCheck="false"
								autoComplete="off"
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

					{/* Origin */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Origin*</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="origin"
								value={formData.origin}
								onChange={handleInput}
								className={`input input-bordered w-full rounded-2xl ${errors.origin ? 'input-error' : ''}`}
								spellCheck="false"
								autoComplete="off"
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

					{/* Chat-completion Path */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Chat Path*</span>
							<span className="label-text-alt tooltip tooltip-right" data-tip="Endpoint path for chat completions.">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="chatCompletionPathPrefix"
								value={formData.chatCompletionPathPrefix}
								onChange={handleInput}
								className="input input-bordered w-full rounded-2xl"
								spellCheck="false"
								autoComplete="off"
							/>
						</div>
					</div>

					{/* API-secret header key */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">API-Auth-Key Header</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="apiKeyHeaderKey"
								value={formData.apiKeyHeaderKey}
								onChange={handleInput}
								className="input input-bordered w-full rounded-2xl"
								spellCheck="false"
								autoComplete="off"
							/>
						</div>
					</div>

					{/* Default Headers */}
					<div className="grid grid-cols-12 gap-2 items-start">
						<label className="label col-span-3">
							<span className="label-text text-sm">Default Headers (JSON)</span>
						</label>
						<div className="col-span-9">
							<textarea
								name="defaultHeadersRawJSON"
								value={formData.defaultHeadersRawJSON}
								onChange={handleInput}
								className={`textarea textarea-bordered w-full rounded-2xl h-24 ${
									errors.defaultHeadersRawJSON ? 'textarea-error' : ''
								}`}
								spellCheck="false"
							/>
							{errors.defaultHeadersRawJSON && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.defaultHeadersRawJSON}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* API-secret */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3 flex flex-col items-start gap-0.5">
							{/* main label */}
							<span className="label-text text-sm">API-Auth-Key</span>

							{/* optional note shown only when editing and a key is already set */}
							{mode === 'edit' && apiKeyAlreadySet ? (
								<span className="label-text-alt text-xs text-gray-500">(leave blank to keep current)</span>
							) : (
								/* the asterisk for required field in “add” mode */
								<span className="label-text-alt text-error text-xs">*</span>
							)}
						</label>

						<div className="col-span-9">
							<input
								type="password"
								name="apiKey"
								value={formData.apiKey}
								onChange={handleInput}
								className={`input input-bordered w-full rounded-2xl ${errors.apiKey ? 'input-error' : ''}`}
								placeholder={mode === 'edit' && apiKeyAlreadySet ? '********' : ''}
								spellCheck="false"
								autoComplete="off"
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
					{/* Enabled toggle */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3 cursor-pointer">
							<span className="label-text text-sm">Enabled</span>
						</label>
						<div className="col-span-9">
							<input
								type="checkbox"
								name="isEnabled"
								checked={formData.isEnabled}
								onChange={handleInput}
								className="toggle toggle-accent rounded-full"
							/>
						</div>
					</div>

					{/* actions */}
					<div className="modal-action">
						<button type="button" className="btn rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-primary rounded-2xl" disabled={!allValid}>
							{mode === 'add' ? 'Add Provider' : 'Save'}
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default AddEditProviderPresetModal;
