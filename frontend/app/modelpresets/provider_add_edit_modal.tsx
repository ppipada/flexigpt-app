import type { ChangeEvent, FormEvent } from 'react';
import { useEffect, useMemo, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

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

import { Dropdown } from '@/components/dropdown';

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
	sdkType: ProviderSDKType.ProviderSDKTypeOpenAIChatCompletions,
	isEnabled: true,
	origin: '',
	chatCompletionPathPrefix: '/v1/chat/completions',
	apiKeyHeaderKey: 'Authorization',
	defaultHeadersRawJSON: '',
	apiKey: '',
};

interface AddEditProviderPresetModalProps {
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

type ErrorState = Partial<Record<keyof FormData, string>>;

export function AddEditProviderPresetModal({
	isOpen,
	mode,
	onClose,
	onSubmit,
	existingProviderNames,
	allProviderPresets,
	initialPreset,
	apiKeyAlreadySet = false,
}: AddEditProviderPresetModalProps) {
	const [formData, setFormData] = useState<FormData>(DEFAULT_FORM);
	const [errors, setErrors] = useState<ErrorState>({});

	const dialogRef = useRef<HTMLDialogElement | null>(null);

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

			[ProviderSDKType.ProviderSDKTypeOpenAIChatCompletions]: {
				isEnabled: true,
				displayName: SDK_DISPLAY_NAME[ProviderSDKType.ProviderSDKTypeOpenAIChatCompletions],
			},

			[ProviderSDKType.ProviderSDKTypeOpenAIResponses]: {
				isEnabled: true,
				displayName: SDK_DISPLAY_NAME[ProviderSDKType.ProviderSDKTypeOpenAIResponses],
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

	// Open/close native dialog
	useEffect(() => {
		if (!isOpen) return;
		const dialog = dialogRef.current;
		if (!dialog) return;

		if (!dialog.open) dialog.showModal();

		return () => {
			if (dialog.open) dialog.close();
		};
	}, [isOpen]);

	const handleDialogClose = () => {
		onClose();
	};

	const validateField = (field: keyof FormData, val: string, currentErrors: ErrorState): ErrorState => {
		let newErrs: ErrorState = { ...currentErrors };
		const v = typeof val === 'string' ? val.trim() : val;

		if (field === 'providerName' && mode === 'add') {
			if (!v) newErrs.providerName = 'Provider name required.';
			else if (typeof v === 'string' && !/^[\w-]+$/.test(v))
				newErrs.providerName = 'Letters, numbers, dash & underscore only.';
			else if (typeof v === 'string' && existingProviderNames.includes(v))
				newErrs.providerName = 'Provider already exists.';
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

		return newErrs;
	};

	const validateForm = (state: FormData): ErrorState => {
		let next: ErrorState = {};
		next = validateField('providerName', state.providerName, next);
		next = validateField('displayName', state.displayName, next);
		next = validateField('origin', state.origin, next);
		next = validateField('defaultHeadersRawJSON', state.defaultHeadersRawJSON, next);
		if (mode === 'add' || state.apiKey.trim()) {
			next = validateField('apiKey', state.apiKey, next);
		}
		next = validateField('sdkType', state.sdkType, next);
		return next;
	};

	const handleInput = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;
		const newVal = type === 'checkbox' ? checked : value;

		setFormData(prev => ({ ...prev, [name]: newVal }));

		if (['providerName', 'displayName', 'origin', 'defaultHeadersRawJSON', 'apiKey'].includes(name)) {
			setErrors(prev => validateField(name as keyof FormData, String(newVal), prev));
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
		setErrors(prev => validateField('sdkType', key, prev));
	};

	const allValid = useMemo(() => {
		const validationErrors = validateForm(formData);
		const hasErr = Object.values(validationErrors).some(Boolean);
		const requiredFilled =
			formData.providerName.trim() &&
			formData.displayName.trim() &&
			formData.origin.trim() &&
			(mode === 'add' ? formData.apiKey.trim() : true);

		return !hasErr && requiredFilled;
	}, [formData, mode]);

	const handleSubmit = (e: FormEvent) => {
		e.preventDefault();

		const finalErrors = validateForm(formData);
		setErrors(finalErrors);
		if (Object.values(finalErrors).some(Boolean)) return;

		let defaultHeaders: Record<string, string> = {};
		if (formData.defaultHeadersRawJSON.trim()) {
			try {
				defaultHeaders = JSON.parse(formData.defaultHeadersRawJSON.trim());
			} catch {
				// Should already be reflected in errors via validateForm
				return;
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
		dialogRef.current?.close();
	};

	if (!isOpen) return null;

	return createPortal(
		<dialog ref={dialogRef} className="modal" onClose={handleDialogClose}>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-3xl overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					{/* Header */}
					<div className="mb-4 flex items-center justify-between">
						<h3 className="text-lg font-bold">{mode === 'add' ? 'Add Provider' : 'Edit Provider'}</h3>
						<button
							type="button"
							className="btn btn-sm btn-circle bg-base-300"
							onClick={() => dialogRef.current?.close()}
							aria-label="Close"
						>
							<FiX size={12} />
						</button>
					</div>

					<form noValidate onSubmit={handleSubmit} className="space-y-4">
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
											className="btn btn-sm btn-ghost flex items-center rounded-xl"
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
												className="btn btn-sm btn-ghost rounded-xl"
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
									className={`input input-bordered w-full rounded-xl ${errors.providerName ? 'input-error' : ''}`}
									disabled={mode === 'edit'}
									spellCheck="false"
									autoComplete="off"
									autoFocus
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
									className={`input input-bordered w-full rounded-xl ${errors.displayName ? 'input-error' : ''}`}
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
									className={`input input-bordered w-full rounded-xl ${errors.origin ? 'input-error' : ''}`}
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
									className="input input-bordered w-full rounded-xl"
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
									className="input input-bordered w-full rounded-xl"
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
									className={`textarea textarea-bordered h-24 w-full rounded-xl ${
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
									className={`input input-bordered w-full rounded-xl ${errors.apiKey ? 'input-error' : ''}`}
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
							<button type="button" className="btn bg-base-300 rounded-xl" onClick={() => dialogRef.current?.close()}>
								Cancel
							</button>
							<button type="submit" className="btn btn-primary rounded-xl" disabled={!allValid}>
								{mode === 'add' ? 'Add Provider' : 'Save'}
							</button>
						</div>
					</form>
				</div>
			</div>
			{/* NOTE: no modal-backdrop here: backdrop click should NOT close this modal */}
		</dialog>,
		document.body
	);
}
