/* --------------------------------------------------------------------------
 * Add / Edit Provider-Preset Modal
 * ------------------------------------------------------------------------*/
import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiX } from 'react-icons/fi';

import { ProviderAPIType, type ProviderName, type ProviderPreset } from '@/spec/modelpreset';

import { omitManyKeys } from '@/lib/obj_utils';
import { isValidUrl } from '@/lib/text_utils';

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

const defaultForm: FormData = {
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

const AddEditProviderPresetModal: FC<Props> = ({
	isOpen,
	mode,
	onClose,
	onSubmit,
	existingProviderNames,
	initialPreset,
	apiKeyAlreadySet = false,
}) => {
	/* form-state ───────────────────────────────────────── */
	const [formData, setFormData] = useState<FormData>(defaultForm);
	const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});

	/* init on open / mode */
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
				defaultHeadersRawJSON: JSON.stringify(initialPreset.defaultHeaders),
				apiKey: '',
			});
		} else {
			setFormData(defaultForm);
		}
		setErrors({});
	}, [isOpen, mode, initialPreset]);

	const validate = (field: keyof FormData, val: string) => {
		const newErrs = omitManyKeys(errors, [field]) as Partial<Record<keyof FormData, string>>;

		if (field === 'providerName' && mode === 'add') {
			if (!val.trim()) newErrs.providerName = 'Provider name required.';
			else if (!/^[\w-]+$/.test(val.trim())) newErrs.providerName = 'Letters, numbers, dash and underscore only.';
			else if (existingProviderNames.includes(val.trim())) newErrs.providerName = 'Provider already exists.';
		}

		if (field === 'displayName' && !val.trim()) newErrs.displayName = 'Display name required.';

		if (field === 'origin') {
			if (!isValidUrl(val)) {
				newErrs.origin = 'Must be a valid URL.';
			}
		}

		if (field === 'defaultHeadersRawJSON' && val.trim()) {
			try {
				JSON.parse(val);
			} catch {
				newErrs.defaultHeadersRawJSON = 'Invalid JSON.';
			}
		}

		if (field === 'apiKey' && mode === 'add' && !val.trim()) newErrs.apiKey = 'API key required.';

		setErrors(newErrs);
	};

	/* on-change handler */
	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;

		const v = type === 'checkbox' ? checked : value;
		setFormData(prev => ({ ...prev, [name]: v }));
		validate(name as keyof FormData, v as string);
	};

	const allValid = useMemo(() => {
		const noErr = !Object.values(errors).some(Boolean);
		const required =
			formData.providerName.trim() &&
			formData.displayName.trim() &&
			formData.origin.trim() &&
			(mode === 'add' ? formData.apiKey.trim() : true);
		return noErr && required;
	}, [errors, formData, mode]);

	/* submit ───────────────────────────────────────────── */
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		(Object.entries(formData) as [keyof FormData, string][]).forEach(([k, v]) => {
			if (typeof v === 'string') validate(k, v);
		});
		if (!allValid) return;

		let headersObj: Record<string, string> = {};
		if (formData.defaultHeadersRawJSON.trim()) {
			try {
				headersObj = JSON.parse(formData.defaultHeadersRawJSON.trim());
			} catch {
				/* already validated */
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
			defaultHeaders: headersObj,
		};

		onSubmit(formData.providerName.trim(), payload, formData.apiKey.trim() ? formData.apiKey.trim() : null);
		onClose();
	};

	if (!isOpen) return null;

	/* render ───────────────────────────────────────────── */
	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* header */}
				<div className="flex justify-between items-center">
					<h3 className="font-bold text-lg">{mode === 'add' ? 'Add New Provider' : 'Edit Provider'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} title="Close">
						<FiX size={12} />
					</button>
				</div>

				<form onSubmit={handleSubmit} className="space-y-4 mt-4">
					{/* Provider ID */}
					<div className="form-control">
						<label className="label">
							<span className="label-text text-sm">Provider ID*</span>
						</label>
						<input
							type="text"
							name="providerName"
							className={`input input-bordered rounded-2xl ${errors.providerName ? 'input-error' : ''}`}
							placeholder="e.g. openai"
							value={formData.providerName}
							onChange={handleChange}
							disabled={mode === 'edit'}
							spellCheck="false"
						/>
						{errors.providerName && (
							<span className="text-error text-xs mt-1 flex items-center gap-1">
								<FiAlertCircle size={12} /> {errors.providerName}
							</span>
						)}
					</div>

					{/* Display name */}
					<div className="form-control">
						<label className="label">
							<span className="label-text text-sm">Display Name*</span>
						</label>
						<input
							type="text"
							name="displayName"
							className={`input input-bordered rounded-2xl ${errors.displayName ? 'input-error' : ''}`}
							value={formData.displayName}
							onChange={handleChange}
							spellCheck="false"
						/>
						{errors.displayName && (
							<span className="text-error text-xs mt-1 flex items-center gap-1">
								<FiAlertCircle size={12} /> {errors.displayName}
							</span>
						)}
					</div>

					{/* Origin */}
					<div className="form-control">
						<label className="label">
							<span className="label-text text-sm">Origin*</span>
						</label>
						<input
							type="text"
							name="origin"
							className={`input input-bordered rounded-2xl ${errors.origin ? 'input-error' : ''}`}
							value={formData.origin}
							onChange={handleChange}
							spellCheck="false"
						/>
						{errors.origin && (
							<span className="text-error text-xs mt-1 flex items-center gap-1">
								<FiAlertCircle size={12} /> {errors.origin}
							</span>
						)}
					</div>

					{/* Chat path */}
					<div className="form-control">
						<label className="label">
							<span className="label-text text-sm">Chat-completion Path*</span>
						</label>
						<input
							type="text"
							name="chatCompletionPathPrefix"
							className="input input-bordered rounded-2xl"
							value={formData.chatCompletionPathPrefix}
							onChange={handleChange}
							spellCheck="false"
						/>
					</div>

					{/* API-key header key */}
					<div className="form-control">
						<label className="label">
							<span className="label-text text-sm">API-Key Header</span>
						</label>
						<input
							type="text"
							name="apiKeyHeaderKey"
							className="input input-bordered rounded-2xl"
							value={formData.apiKeyHeaderKey}
							onChange={handleChange}
							spellCheck="false"
						/>
					</div>

					{/* Default headers */}
					<div className="form-control">
						<label className="label">
							<span className="label-text text-sm">Default Headers (JSON)</span>
						</label>
						<textarea
							name="defaultHeadersRawJSON"
							className={`textarea textarea-bordered rounded-2xl ${
								errors.defaultHeadersRawJSON ? 'textarea-error' : ''
							}`}
							rows={3}
							value={formData.defaultHeadersRawJSON}
							onChange={handleChange}
							spellCheck="false"
						/>
						{errors.defaultHeadersRawJSON && (
							<span className="text-error text-xs mt-1 flex items-center gap-1">
								<FiAlertCircle size={12} /> {errors.defaultHeadersRawJSON}
							</span>
						)}
					</div>

					{/* API-key */}
					<div className="form-control">
						<label className="label">
							<span className="label-text text-sm">
								{mode === 'add' ? 'API-Key*' : `API-Key ${apiKeyAlreadySet ? '(leave blank = keep current)' : '*'}`}
							</span>
						</label>
						<input
							type="password"
							name="apiKey"
							className={`input input-bordered rounded-2xl ${errors.apiKey ? 'input-error' : ''}`}
							placeholder={mode === 'edit' && apiKeyAlreadySet ? '********' : ''}
							value={formData.apiKey}
							onChange={handleChange}
							spellCheck="false"
						/>
						{errors.apiKey && (
							<span className="text-error text-xs mt-1 flex items-center gap-1">
								<FiAlertCircle size={12} /> {errors.apiKey}
							</span>
						)}
					</div>

					{/* Enabled toggle */}
					<div className="form-control">
						<label className="label cursor-pointer">
							<span className="label-text text-sm">Enabled</span>
							<input
								type="checkbox"
								name="isEnabled"
								className="toggle toggle-accent rounded-full"
								checked={formData.isEnabled}
								onChange={handleChange}
							/>
						</label>
					</div>

					{/* actions */}
					<div className="modal-action">
						<button type="button" className="btn rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" disabled={!allValid} className="btn btn-primary rounded-2xl">
							{mode === 'add' ? 'Add Provider' : 'Save'}
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default AddEditProviderPresetModal;
