import { type FC, useEffect, useState } from 'react';
import { FiPlus, FiX } from 'react-icons/fi';

import { type ProviderName } from '@/models/aiprovidermodel';
import type { AISetting, ModelSetting } from '@/models/settingmodel';

import ModifyModelModal from '@/settings/model_modify_modal';

interface AddProviderModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (providerName: ProviderName, newSettings: AISetting) => void;
}

const AddProviderModal: FC<AddProviderModalProps> = ({ isOpen, onClose, onSubmit }) => {
	const [providerName, setProviderName] = useState('');
	const [apiKey, setApiKey] = useState('');
	const [origin, setOrigin] = useState('');
	const [chatCompletionPathPrefix, setChatCompletionPathPrefix] = useState('');

	// For the "default" model
	const [defaultModelName, setDefaultModelName] = useState('');
	const [modelSettings, setModelSettings] = useState<Record<string, ModelSetting>>({});

	// State for controlling the internal ModifyModelModal
	const [isModifyModelModalOpen, setIsModifyModelModalOpen] = useState(false);

	useEffect(() => {
		if (!isOpen) {
			// Reset states when closed
			setProviderName('');
			setApiKey('');
			setOrigin('');
			setChatCompletionPathPrefix('');
			setDefaultModelName('');
			setModelSettings({});
			setIsModifyModelModalOpen(false);
		}
	}, [isOpen]);

	const handleAddDefaultModel = () => {
		setIsModifyModelModalOpen(true);
	};

	const handleModifyModelSubmit = (modelName: string, modelData: ModelSetting) => {
		setDefaultModelName(modelName);
		setModelSettings({ [modelName]: modelData });
		setIsModifyModelModalOpen(false);
	};

	const handleConfirm = () => {
		// Basic validation
		if (!providerName.trim()) {
			alert('Provider name is required.');
			return;
		}
		if (!defaultModelName.trim()) {
			alert('Please configure at least one default model.');
			return;
		}

		const newProviderSettings: AISetting = {
			isEnabled: true, // new providers are enabled by default
			apiKey: apiKey,
			origin: origin,
			chatCompletionPathPrefix: chatCompletionPathPrefix,
			defaultModel: defaultModelName,
			modelSettings: modelSettings,
		};

		onSubmit(providerName, newProviderSettings);
		onClose();
	};

	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* Header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">Add Provider</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close">
						<FiX size={12} />
					</button>
				</div>

				{/* Body (grid-based layout) */}
				<div className="space-y-4">
					{/* Provider Name */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Provider Name</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								className="input input-bordered w-full rounded-xl"
								value={providerName}
								onChange={e => {
									setProviderName(e.target.value.toLocaleLowerCase());
								}}
								placeholder="e.g. openai2"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* API Key */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">API Key</span>
						</label>
						<div className="col-span-9">
							<input
								type="password"
								className="input input-bordered w-full rounded-xl"
								value={apiKey}
								onChange={e => {
									setApiKey(e.target.value);
								}}
								placeholder="Your provider's API key"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Origin */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Origin/FQDN</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								className="input input-bordered w-full rounded-xl"
								value={origin}
								onChange={e => {
									setOrigin(e.target.value);
								}}
								placeholder="e.g. https://api.provider.com"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Chat completion path prefix */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Chat Path Prefix</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								className="input input-bordered w-full rounded-xl"
								value={chatCompletionPathPrefix}
								onChange={e => {
									setChatCompletionPathPrefix(e.target.value);
								}}
								placeholder="e.g. /v1/chat/completions"
								spellCheck="false"
							/>
						</div>
					</div>

					{/* Default Model */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Default Model</span>
						</label>
						<div className="col-span-9 flex flex-col space-y-2">
							<div className="flex items-center justify-between">
								{defaultModelName ? (
									<p className="text-sm">
										Configured: <strong>{defaultModelName}</strong>
									</p>
								) : (
									<p className="text-sm italic">No default model configured</p>
								)}
								<button className="btn btn-md btn-ghost flex items-center rounded-xl" onClick={handleAddDefaultModel}>
									<FiPlus size={16} />
									<span className="ml-1">Configure Default Model</span>
								</button>
							</div>
						</div>
					</div>
				</div>

				{/* Actions */}
				<div className="modal-action mt-6">
					<button type="button" className="btn rounded-xl" onClick={onClose}>
						Cancel
					</button>
					<button type="button" className="btn btn-primary rounded-xl" onClick={handleConfirm}>
						Add Provider
					</button>
				</div>
			</div>

			{/* Reuse of ModifyModelModal to define the default model */}
			{isModifyModelModalOpen && (
				<ModifyModelModal
					isOpen={isModifyModelModalOpen}
					onClose={() => {
						setIsModifyModelModalOpen(false);
					}}
					onSubmit={handleModifyModelSubmit}
					providerName={providerName}
					initialModelName={defaultModelName || undefined}
					initialData={defaultModelName ? modelSettings[defaultModelName] : undefined}
					existingModels={modelSettings}
				/>
			)}
		</dialog>
	);
};

export default AddProviderModal;
