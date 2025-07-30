import type { FC } from 'react';
import { useEffect, useState } from 'react';

import { FiCheckCircle, FiChevronDown, FiChevronUp, FiHelpCircle, FiTrash2, FiXCircle } from 'react-icons/fi';

import type { ProviderName } from '@/spec/modelpreset';
import { ProviderInfoDescription } from '@/spec/modelpreset';
import type { AISetting, AISettingAttrs } from '@/spec/setting';

import { SetAISettingAPIKey, SetAISettingAttrs } from '@/apis/settingstore_helper';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';

interface ProviderSettingsCardProps {
	provider: ProviderName;
	settings: AISetting;
	defaultProvider: ProviderName;
	inbuiltProvider?: boolean;
	enabledProviders: ProviderName[];
	onProviderSettingChange: (provider: ProviderName, settings: AISetting) => void;
	onProviderDelete: (provider: ProviderName) => Promise<void>;
}

const isValidUrl = (url: string) => {
	try {
		// Accept empty string as valid (optional: require non-empty)
		if (!url) return false;
		new URL(url);
		return true;
	} catch {
		return false;
	}
};

const ProviderSettingsCard: FC<ProviderSettingsCardProps> = ({
	provider,
	settings,
	defaultProvider,
	inbuiltProvider = false,
	enabledProviders,
	onProviderSettingChange,
	onProviderDelete,
}) => {
	const [isExpanded, setIsExpanded] = useState(false);
	const [isEnabled, setIsEnabled] = useState(!!settings.isEnabled);
	const [localSettings, setLocalSettings] = useState<AISetting>(settings);

	const [showActionDeniedAlert, setShowActionDeniedAlert] = useState(false);
	const [actionDeniedMessage, setActionDeniedMessage] = useState('');
	const [showDeleteModal, setShowDeleteModal] = useState(false);

	// Validation states
	const [apiKeyError, setApiKeyError] = useState('');
	const [originError, setOriginError] = useState('');
	const [chatPathError, setChatPathError] = useState('');

	const isLastEnabled = isEnabled && enabledProviders.length === 1;

	useEffect(() => {
		setIsEnabled(!!settings.isEnabled);
		setLocalSettings(settings);
	}, [settings]);

	const toggleExpand = () => {
		if (isEnabled) setIsExpanded(prev => !prev);
	};

	const updateLocalSettings = (key: keyof AISetting, value: any) => {
		const updated = { ...localSettings, [key]: value };
		setLocalSettings(updated);
		onProviderSettingChange(provider, updated);
	};

	const validateApiKey = (apiKey: string) => {
		if (!apiKey || apiKey.trim() === '') {
			setApiKeyError('API Key cannot be empty.');
			return false;
		}
		setApiKeyError('');
		return true;
	};

	const validateOrigin = (origin: string) => {
		if (!origin || !isValidUrl(origin)) {
			setOriginError('Origin must be a valid URL (e.g., https://api.example.com)');
			return false;
		}
		setOriginError('');
		return true;
	};

	const validateChatPath = (path: string) => {
		// Optional: require starts with /
		if (path && !path.startsWith('/')) {
			setChatPathError('Path should start with a "/"');
			return false;
		}
		setChatPathError('');
		return true;
	};

	const toggleProviderEnable = async () => {
		const newState = !isEnabled;

		if (!newState) {
			// guard: cannot disable default provider
			if (provider === defaultProvider) {
				setActionDeniedMessage('Cannot disable the default provider. Choose another default first.');
				setShowActionDeniedAlert(true);
				return;
			}
			// Optional: Prevent disabling last enabled provider
			if (isLastEnabled) {
				setActionDeniedMessage('Cannot disable the last enabled provider.');
				setShowActionDeniedAlert(true);
				return;
			}
		}

		setIsEnabled(newState);
		updateLocalSettings('isEnabled', newState);
		await SetAISettingAttrs(provider, { isEnabled: newState } as AISettingAttrs);
	};

	const handleProviderDelete = () => {
		if (provider === defaultProvider || inbuiltProvider) {
			setActionDeniedMessage('Cannot delete the default or in-built provider.');
			setShowActionDeniedAlert(true);
			return;
		}
		setShowDeleteModal(true);
	};

	const confirmDelete = async () => {
		await onProviderDelete(provider);
		setShowDeleteModal(false);
	};

	return (
		<div className="bg-base-100 rounded-2xl shadow-lg p-4 mb-8">
			{/* ── Header row ─────────────────────────────────────── */}
			<div className="grid grid-cols-12 gap-4 items-center">
				{/* Provider Title*/}
				<div className="col-span-3 flex items-center space-x-4">
					<h3 className="text-sm font-semibold capitalize">{provider}</h3>
				</div>
				{/* Enable/Disable Toggle */}
				<div className="col-span-3 flex items-center space-x-2 ml-1">
					<label className="text-sm">Enable</label>
					<input
						type="checkbox"
						checked={isEnabled}
						onChange={toggleProviderEnable}
						className="toggle toggle-accent rounded-full"
					/>
				</div>

				{/* Status & chevron */}
				<div className="col-span-6 cursor-pointer flex items-end justify-end gap-4" onClick={toggleExpand}>
					<div className="flex items-center">
						<span className="text-sm">API Key</span>
						{localSettings.apiKey ? (
							<FiCheckCircle className="text-success mx-2" />
						) : (
							<FiXCircle className="text-error mx-2" />
						)}
					</div>
					<div className="flex items-center">
						<span className="text-sm">Details</span>
						{isExpanded ? <FiChevronUp className="mx-2" /> : <FiChevronDown className="mx-2" />}
					</div>
				</div>
			</div>

			{/* ── Body ───────────────────────────────────────────── */}
			{isEnabled && isExpanded && (
				<div className="mt-8 space-y-4">
					{/* API-Key */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="label col-span-3">
							<span className="text-sm">API Key</span>
							<span className="tooltip tooltip-right" data-tip={ProviderInfoDescription.apiKey}>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="password"
								className={`input w-full input-bordered rounded-2xl ${apiKeyError ? 'input-error' : ''}`}
								value={localSettings.apiKey}
								onChange={e => {
									setLocalSettings({ ...localSettings, apiKey: e.target.value });
									if (apiKeyError) validateApiKey(e.target.value);
								}}
								onBlur={async e => {
									if (!validateApiKey(e.target.value)) return;
									updateLocalSettings('apiKey', e.target.value);
									await SetAISettingAPIKey(provider, e.target.value);
								}}
								spellCheck="false"
							/>
							{apiKeyError && <div className="text-error text-xs mt-1">{apiKeyError}</div>}
						</div>
					</div>

					{/* Origin */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="label col-span-3">
							<span className="text-sm">Origin</span>
							<span className="tooltip tooltip-right" data-tip={ProviderInfoDescription.origin}>
								<FiHelpCircle size={12} />
							</span>
						</label>

						<div className="col-span-9">
							<input
								type="text"
								className={`input w-full input-bordered rounded-2xl ${originError ? 'input-error' : ''}`}
								value={localSettings.origin}
								onChange={e => {
									setLocalSettings({ ...localSettings, origin: e.target.value });
									if (originError) validateOrigin(e.target.value);
								}}
								onBlur={async e => {
									if (!validateOrigin(e.target.value)) return;
									updateLocalSettings('origin', e.target.value);
									await SetAISettingAttrs(provider, { origin: e.target.value } as AISettingAttrs);
								}}
								spellCheck="false"
							/>
							{originError && <div className="text-error text-xs mt-1">{originError}</div>}
						</div>
					</div>

					{/* Chat-completion prefix */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="label col-span-3">
							<span className="text-sm">Chat Path</span>
							<span className="tooltip tooltip-right" data-tip={ProviderInfoDescription.chatCompletionPathPrefix}>
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								className={`input w-full input-bordered rounded-2xl ${chatPathError ? 'input-error' : ''}`}
								value={localSettings.chatCompletionPathPrefix}
								onChange={e => {
									setLocalSettings({ ...localSettings, chatCompletionPathPrefix: e.target.value });
									if (chatPathError) validateChatPath(e.target.value);
								}}
								onBlur={async e => {
									if (!validateChatPath(e.target.value)) return;
									updateLocalSettings('chatCompletionPathPrefix', e.target.value);
									await SetAISettingAttrs(provider, {
										chatCompletionPathPrefix: e.target.value,
									} as AISettingAttrs);
								}}
								spellCheck="false"
							/>
							{chatPathError && <div className="text-error text-xs mt-1">{chatPathError}</div>}
						</div>
					</div>

					{/* Delete provider */}
					<div className="flex justify-end">
						<button
							className="btn btn-md btn-ghost rounded-2xl flex items-center"
							onClick={handleProviderDelete}
							disabled={provider === defaultProvider || inbuiltProvider}
							title={
								provider === defaultProvider || inbuiltProvider
									? 'Cannot delete default / in-built provider'
									: 'Delete Provider'
							}
						>
							<FiTrash2 size={16} /> Delete Provider
						</button>
					</div>
				</div>
			)}

			{/* ── Modals / Alerts ───────────────────────────────── */}
			{showActionDeniedAlert && (
				<ActionDeniedAlert
					isOpen={showActionDeniedAlert}
					onClose={() => {
						setShowActionDeniedAlert(false);
						setActionDeniedMessage('');
					}}
					message={actionDeniedMessage}
				/>
			)}

			{showDeleteModal && (
				<DeleteConfirmationModal
					isOpen={showDeleteModal}
					title="Delete Provider"
					message={`Really delete provider "${provider}"?  This cannot be undone.`}
					confirmButtonText="Delete"
					onConfirm={confirmDelete}
					onClose={() => {
						setShowDeleteModal(false);
					}}
				/>
			)}
		</div>
	);
};

export default ProviderSettingsCard;
