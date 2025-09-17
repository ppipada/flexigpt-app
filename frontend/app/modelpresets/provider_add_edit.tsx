import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiUpload, FiX } from 'react-icons/fi';

import {
	type ProviderName,
	type ProviderPreset,
	ProviderSDKType,
	SDK_DEFAULTS,
	SDK_DISPLAY_NAME,
} from '@/spec/modelpreset';

import { GenerateRandomNumberString } from '@/lib/encode_decode';
import { omitManyKeys } from '@/lib/obj_utils';
import { isValidUrl } from '@/lib/text_utils';

import Dropdown from '@/components/dropdown';

type FormData = {
	providerName: string;
	displayName: string;
	sdkType: ProviderSDKType;
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
	sdkType: ProviderSDKType.ProviderSDKTypeOpenAI,
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
	allProviderPresets: Record<ProviderName, ProviderPreset>;

	/* edit-mode helpers */
	initialPreset?: ProviderPreset;
	apiKeyAlreadySet?: boolean;
}

const AddEditProviderPresetModal: FC<Props> = ({
	isOpen,
	mode,
	onClose,
	onSubmit,
	existingProviderNames,
	allProviderPresets,
	initialPreset,
	apiKeyAlreadySet = false,
}) => {
	const [formData, setFormData] = useState<FormData>(DEFAULT_FORM);
	const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});

	const prefillDropdownItems: Record<ProviderName, { isEnabled: boolean; displayName: string }> = useMemo(() => {
		const o: Record<ProviderName, { isEnabled: boolean; displayName: string }> = {} as any;
		for (const [name, p] of Object.entries(allProviderPresets)) {
			o[name] = { isEnabled: true, displayName: p.displayName || name };
		}
		return o;
	}, [allProviderPresets]);

	const sdkDropdownItems: Record<ProviderSDKType, { isEnabled: boolean; displayName: string }> = useMemo(
		() => ({
			[ProviderSDKType.ProviderSDKTypeAnthropic]: {
				isEnabled: true,
				displayName: SDK_DISPLAY_NAME[ProviderSDKType.ProviderSDKTypeAnthropic],
			},

			[ProviderSDKType.ProviderSDKTypeOpenAI]: {
				isEnabled: true,
				displayName: SDK_DISPLAY_NAME[ProviderSDKType.ProviderSDKTypeOpenAI],
			},
		}),
		[]
	);

	const [prefillMode, setPrefillMode] = useState(false);
	const [selectedPrefillKey, setSelectedPrefillKey] = useState<ProviderName | null>(null);

	const applyPrefill = (key: ProviderName) => {
		const src = allProviderPresets[key];

		setFormData(prev => ({
			...prev,

			/* IDs & secrets are intentionally NOT copied */
			displayName: src.displayName + '-' + GenerateRandomNumberString(3),
			sdkType: src.sdkType,
			isEnabled: true, // always enable newly added presets
			origin: src.origin,
			chatCompletionPathPrefix: src.chatCompletionPathPrefix,
			apiKeyHeaderKey: src.apiKeyHeaderKey,
			defaultHeadersRawJSON: JSON.stringify(src.defaultHeaders, null, 2),
			apiKey: '',
		}));
	};

	useEffect(() => {
		if (!isOpen) return;

		if (mode === 'edit' && initialPreset) {
			setFormData({
				providerName: initialPreset.name,
				displayName: initialPreset.displayName,
				sdkType: initialPreset.sdkType,
				isEnabled: initialPreset.isEnabled,
				origin: initialPreset.origin,
				chatCompletionPathPrefix: initialPreset.chatCompletionPathPrefix,
				apiKeyHeaderKey: initialPreset.apiKeyHeaderKey,
				defaultHeadersRawJSON: JSON.stringify(initialPreset.defaultHeaders, null, 2),
				apiKey: '',
			});
		} else {
			setFormData(DEFAULT_FORM);
			/* reset prefill helpers */
			setPrefillMode(false);
			setSelectedPrefillKey(null);
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

		if (field === 'sdkType') {
			if (!Object.values(ProviderSDKType).includes(v as ProviderSDKType)) newErrs.sdkType = 'Invalid SDK type.';
			else newErrs = omitManyKeys(newErrs, ['sdkType']);
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

	const onSdkTypeChange = (key: ProviderSDKType) => {
		setFormData(prev => {
			const next = { ...prev, sdkType: key };
			const defaults = SDK_DEFAULTS[key];

			// Fill associated fields only if empty
			if (!prev.chatCompletionPathPrefix.trim()) next.chatCompletionPathPrefix = defaults.chatPath;
			if (!prev.apiKeyHeaderKey.trim()) next.apiKeyHeaderKey = defaults.apiKeyHeaderKey;
			if (!prev.defaultHeadersRawJSON.trim())
				next.defaultHeadersRawJSON = JSON.stringify(defaults.defaultHeaders, null, 2);

			return next;
		});
		validateField('sdkType', key);
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

		/* final validation pass */
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
			sdkType: formData.sdkType,
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
			<div className="modal-box max-h-[80vh] max-w-3xl overflow-auto rounded-2xl">
				{/* Header */}
				<div className="mb-4 flex items-center justify-between">
					<h3 className="text-lg font-bold">{mode === 'add' ? 'Add Provider' : 'Edit Provider'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close" title="Close">
						<FiX size={12} />
					</button>
				</div>

				<form onSubmit={handleSubmit} className="space-y-4">
					{/* PREFILL (ADD mode only) */}
					{mode === 'add' && (
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Prefill from Existing</span>
							</label>

							<div className="col-span-9 flex items-center gap-2">
								{!prefillMode && (
									<button
										type="button"
										className="btn btn-sm btn-ghost flex items-center rounded-2xl"
										onClick={() => {
											setPrefillMode(true);
										}}
									>
										<FiUpload size={14} />
										<span className="ml-1">Copy Existing Provider</span>
									</button>
								)}

								{prefillMode && (
									<>
										<Dropdown<ProviderName>
											dropdownItems={prefillDropdownItems}
											selectedKey={selectedPrefillKey ?? ('' as ProviderName)}
											onChange={key => {
												setSelectedPrefillKey(key);
												applyPrefill(key);
												setPrefillMode(false); // auto-close
											}}
											filterDisabled={false}
											title="Select provider to copy"
											getDisplayName={k => prefillDropdownItems[k].displayName}
										/>
										<button
											type="button"
											className="btn btn-sm btn-ghost rounded-2xl"
											onClick={() => {
												setPrefillMode(false);
												setSelectedPrefillKey(null);
											}}
											title="Cancel prefill"
										>
											<FiX size={12} />
										</button>
									</>
								)}
							</div>
						</div>
					)}

					{/* SDK Type */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">SDK Type*</span>
							<span
								className="label-text-alt tooltip tooltip-right"
								data-tip="Select the backend SDK/API compatibility for this provider."
							>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<Dropdown<ProviderSDKType>
								dropdownItems={sdkDropdownItems}
								selectedKey={formData.sdkType}
								onChange={onSdkTypeChange}
								filterDisabled={false}
								title="Select SDK Type"
								getDisplayName={k => sdkDropdownItems[k].displayName}
							/>
							{errors.sdkType && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.sdkType}
									</span>
								</div>
							)}
						</div>
					</div>

					{/* Provider ID */}
					<div className="grid grid-cols-12 items-center gap-2">
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
					<div className="grid grid-cols-12 items-center gap-2">
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
					<div className="grid grid-cols-12 items-center gap-2">
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
					<div className="grid grid-cols-12 items-center gap-2">
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

					{/* API-key header key */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">API-Key Header Key</span>
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
					<div className="grid grid-cols-12 items-start gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Default Headers (JSON)</span>
						</label>
						<div className="col-span-9">
							<textarea
								name="defaultHeadersRawJSON"
								value={formData.defaultHeadersRawJSON}
								onChange={handleInput}
								className={`textarea textarea-bordered h-24 w-full rounded-2xl ${
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

					{/* API-Key */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3 flex flex-col items-start gap-0.5">
							<span className="label-text text-sm">API-Key*</span>
							{mode === 'edit' && apiKeyAlreadySet && (
								<span className="label-text-alt text-xs">(leave blank to keep current)</span>
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
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3 cursor-pointer">
							<span className="label-text text-sm">Enabled</span>
						</label>
						<div className="col-span-9">
							<input
								type="checkbox"
								name="isEnabled"
								checked={formData.isEnabled}
								onChange={handleInput}
								className="toggle toggle-accent"
							/>
						</div>
					</div>

					{/* Actions */}
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
