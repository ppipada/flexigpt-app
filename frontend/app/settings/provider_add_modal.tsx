import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import type { ProviderName } from '@/spec/modelpreset';
import type { AISetting } from '@/spec/setting';

import { omitManyKeys } from '@/lib/obj_utils';

interface AddProviderModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (providerName: ProviderName, newSettings: AISetting) => void;
	existingProviderNames: string[];
}

interface ProviderFormData {
	providerName: string;
	apiKey: string;
	origin: string;
	chatCompletionPathPrefix: string;
}

const AddProviderModal: FC<AddProviderModalProps> = ({ isOpen, onClose, onSubmit, existingProviderNames }) => {
	const [formData, setFormData] = useState<ProviderFormData>({
		providerName: '',
		apiKey: '',
		origin: '',
		chatCompletionPathPrefix: '',
	});

	const [errors, setErrors] = useState<Partial<Record<keyof ProviderFormData, string>>>({});

	// Reset form when closed
	useEffect(() => {
		if (!isOpen) {
			setFormData({
				providerName: '',
				apiKey: '',
				origin: '',
				chatCompletionPathPrefix: '',
			});
			setErrors({});
		}
	}, [isOpen]);

	// Validation
	const validateField = (field: keyof ProviderFormData, value: string) => {
		const newErrors = omitManyKeys(errors, [field]) as Record<keyof ProviderFormData, string>;

		if (field === 'providerName') {
			if (!value.trim()) newErrors.providerName = 'Provider name is required.';
			else if (existingProviderNames.includes(value)) newErrors.providerName = 'This provider name already exists.';
		}
		if (field === 'apiKey' && !value.trim()) newErrors.apiKey = 'API key is required.';
		if (field === 'origin') {
			if (!value.trim()) newErrors.origin = 'Origin is required.';
			else {
				try {
					new URL(value);
				} catch {
					newErrors.origin = 'Origin must be a valid URL (e.g. https://api.provider.com)';
				}
			}
		}
		if (field === 'chatCompletionPathPrefix') {
			if (value && !value.startsWith('/')) {
				newErrors.chatCompletionPathPrefix = 'Path should start with a "/"';
			}
		}
		setErrors(newErrors);
	};

	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const { name, value } = e.target;
		setFormData(prev => ({ ...prev, [name]: value }));
		validateField(name as keyof ProviderFormData, value);
	};

	// Overall validity
	const isAllValid = useMemo(() => {
		const noErrors = !Object.values(errors).some(Boolean);
		const hasRequired = formData.providerName.trim() && formData.apiKey.trim() && formData.origin.trim();
		return noErrors && hasRequired;
	}, [errors, formData]);

	// Submit
	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		Object.entries(formData).forEach(([k, v]) => {
			validateField(k as keyof ProviderFormData, v);
		});

		if (!isAllValid) return;

		const settings: AISetting = {
			isEnabled: true,
			apiKey: formData.apiKey,
			origin: formData.origin,
			chatCompletionPathPrefix: formData.chatCompletionPathPrefix,
		};

		onSubmit(formData.providerName.trim(), settings);
		onClose();
	};

	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* Header */}
				<div className="flex justify-between items-center">
					<h3 className="font-bold text-lg">Add New Provider</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close" title="Close">
						<FiX size={12} />
					</button>
				</div>
				<h4 className="flex items-center gap-2 text-xs text-neutral/60 mt-2 mb-8">
					<FiAlertCircle size={16} />
					<span>Only OpenAI API-compatible custom providers are supported.</span>
				</h4>

				<form onSubmit={handleSubmit} className="space-y-4">
					{/* Provider name */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Provider Name*</span>
							<span className="label-text-alt tooltip" data-tip="Unique identifier for this provider.">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								name="providerName"
								value={formData.providerName}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-2xl ${errors.providerName ? 'input-error' : ''}`}
								placeholder="e.g. openai-alt"
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

					{/* API key */}
					<div className="grid grid-cols-12 gap-2 items-center">
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
								value={formData.apiKey}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-2xl ${errors.apiKey ? 'input-error' : ''}`}
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
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Origin / FQDN*</span>
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
								value={formData.origin}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-2xl ${errors.origin ? 'input-error' : ''}`}
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

					{/* Chat-completion prefix */}
					<div className="grid grid-cols-12 gap-2 items-center">
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
								value={formData.chatCompletionPathPrefix}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-2xl ${errors.chatCompletionPathPrefix ? 'input-error' : ''}`}
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

					{/* actions */}
					<div className="modal-action mt-6">
						<button type="button" className="btn btn-md rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-md btn-primary rounded-2xl" disabled={!isAllValid}>
							Add Provider
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default AddProviderModal;
