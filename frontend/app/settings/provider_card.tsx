import type { FC } from 'react';
import { useEffect, useState } from 'react';

import {
	FiAlertTriangle,
	FiCheck,
	FiCheckCircle,
	FiChevronDown,
	FiChevronUp,
	FiEdit,
	FiPlus,
	FiTrash2,
	FiX,
	FiXCircle,
} from 'react-icons/fi';

import type { ModelName, ModelPreset, ProviderName } from '@/models/aiprovidermodel';
import { ProviderInfoDescription } from '@/models/aiprovidermodel';
import type { AISetting, AISettingAttrs, ModelSetting } from '@/models/settingmodel';

import { settingstoreAPI } from '@/apis/baseapi';
import { SetAISettingAPIKey, SetAISettingAttrs } from '@/apis/settingstore_helper';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';
import Dropdown from '@/components/dropdown';

import ModifyModelModal from '@/settings/model_modify_modal';

interface AISettingsCardProps {
	provider: ProviderName;
	settings: AISetting;
	aiSettings: Record<string, AISetting>;
	defaultProvider: ProviderName;
	inbuiltProviderModels: Record<ModelName, ModelPreset> | undefined;
	onProviderSettingChange: (provider: ProviderName, settings: AISetting) => void;
	onProviderDelete: (provider: ProviderName) => Promise<void>;
}

const AISettingsCard: FC<AISettingsCardProps> = ({
	provider,
	settings,
	aiSettings,
	defaultProvider,
	inbuiltProviderModels,
	onProviderSettingChange,
	onProviderDelete,
}) => {
	const [isExpanded, setIsExpanded] = useState(false);
	const [isEnabled, setIsEnabled] = useState(!!settings.isEnabled);
	const [showModal, setShowModal] = useState(false);
	const [localSettings, setLocalSettings] = useState<AISetting>(settings);

	const [modelSettings, setModelSettings] = useState<Record<ModelName, ModelSetting>>(settings.modelSettings);
	const [isModifyModelModalOpen, setIsModifyModelModalOpen] = useState(false);
	const [selectedModelName, setSelectedModelName] = useState<ModelName | null>(null);
	const [isDeleteModelModalOpen, setIsDeleteModelModalOpen] = useState(false);

	const [showActionDeniedAlert, setShowActionDeniedAlert] = useState(false);
	const [actionDeniedMessage, setActionDeniedMessage] = useState('');
	const [isDeleteProviderModalOpen, setIsDeleteProviderModalOpen] = useState(false);

	// Update local state when props change
	useEffect(() => {
		setIsEnabled(!!settings.isEnabled);
		setLocalSettings(settings);
		setModelSettings(settings.modelSettings);
	}, [settings]);

	const toggleExpand = () => {
		if (isEnabled) {
			setIsExpanded(!isExpanded);
		}
	};

	const toggleProviderEnable = async () => {
		const newIsEnabled = !isEnabled;
		if (!newIsEnabled) {
			// Prevent disabling the default provider
			if (provider === defaultProvider) {
				setActionDeniedMessage(
					'Cannot disable the default provider. Please select a different default provider first.'
				);
				setShowActionDeniedAlert(true);
				return;
			}
			const enabledProviders = Object.keys(aiSettings).filter(k => aiSettings[k].isEnabled && k !== provider);
			if (enabledProviders.length === 0) {
				setShowModal(true);
				return;
			}
		}
		setIsEnabled(newIsEnabled);
		updateLocalSettings('isEnabled', newIsEnabled);
		const aiSettingAttrs: AISettingAttrs = {
			isEnabled: newIsEnabled,
		};
		await SetAISettingAttrs(provider, aiSettingAttrs);
	};

	const toggleModelEnable = async (modelName: ModelName) => {
		const model = modelSettings[modelName];

		// If we are about to disable the default model, deny the action.
		if (model.isEnabled && modelName === localSettings.defaultModel) {
			setActionDeniedMessage('Cannot disable the default model. Please select a different default model first.');
			setShowActionDeniedAlert(true);
			return;
		}

		// Otherwise proceed with toggling
		const updatedModel = { ...model, isEnabled: !model.isEnabled };
		const updatedModels = { ...modelSettings, [modelName]: updatedModel };

		// Update local state
		setModelSettings(updatedModels);
		updateLocalSettings('modelSettings', updatedModels);

		// Persist to a backend
		await settingstoreAPI.addModelSetting(provider, modelName, updatedModel);
	};

	// Handle setting changes and save to backend
	const updateLocalSettings = async (key: string, value: any) => {
		const updatedSettings = {
			...localSettings,
			[key]: value,
		};
		setLocalSettings(updatedSettings);
		onProviderSettingChange(provider, updatedSettings);
	};

	// Handlers for models
	const handleAddModel = () => {
		setSelectedModelName(null);
		setIsModifyModelModalOpen(true);
	};

	const handleEditModel = (modelName: ModelName) => {
		setSelectedModelName(modelName);
		setIsModifyModelModalOpen(true);
	};

	const isModelRemovable = (modelName: ModelName) => {
		// Cannot remove if it is the default model
		if (modelName === localSettings.defaultModel) {
			return false;
		}
		// Cannot remove if it is an inbuilt model
		if (inbuiltProviderModels && modelName in inbuiltProviderModels) {
			return false;
		}

		return true;
	};

	const isModelReasoningSupport = (modelName: ModelName) => {
		if (modelSettings[modelName].reasoning) {
			return true;
		}

		if (inbuiltProviderModels && modelName in inbuiltProviderModels && inbuiltProviderModels[modelName].reasoning) {
			return true;
		}

		return false;
	};

	const handleDeleteModel = (modelName: ModelName) => {
		if (!isModelRemovable(modelName)) {
			setActionDeniedMessage('Cannot delete the default model or an inbuilt model.');
			setShowActionDeniedAlert(true);
			return;
		}
		setSelectedModelName(modelName);
		setIsDeleteModelModalOpen(true);
	};

	const handleDeleteModelConfirm = async () => {
		const modelName: ModelName = selectedModelName || '';
		if (!isModelRemovable(modelName)) {
			setActionDeniedMessage('Cannot delete the default model or an inbuilt model.');
			setShowActionDeniedAlert(true);
			return;
		}

		const updatedModels = Object.fromEntries(Object.entries(modelSettings).filter(([key]) => key !== modelName));

		setModelSettings(updatedModels);
		updateLocalSettings('modelSettings', updatedModels);
		await settingstoreAPI.deleteModelSetting(provider, modelName);

		setIsDeleteModelModalOpen(false);
		setSelectedModelName(null);
	};

	const closeDeleteModelModal = () => {
		setIsDeleteModelModalOpen(false);
		setSelectedModelName(null);
	};

	const handleModifyModelSubmit = async (modelName: ModelName, modelData: ModelSetting) => {
		const updatedModels = { ...modelSettings };
		updatedModels[modelName] = modelData;
		setModelSettings(updatedModels);
		updateLocalSettings('modelSettings', updatedModels);
		await settingstoreAPI.addModelSetting(provider, modelName, modelData);
		setIsModifyModelModalOpen(false);
		setSelectedModelName(null);
	};

	const isProviderRemovable = (provider: ProviderName) => {
		if (provider === defaultProvider) return false;
		if (inbuiltProviderModels) return false;
		return true;
	};

	const handleDeleteProvider = (provider: ProviderName) => {
		if (!isProviderRemovable(provider)) {
			setActionDeniedMessage('Cannot delete the default provider or an inbuilt provider.');
			setShowActionDeniedAlert(true);
			return;
		}
		setIsDeleteProviderModalOpen(true);
	};

	const handleDeleteProviderConfirm = async () => {
		await onProviderDelete(provider); // call parent to remove from state & store
		setIsDeleteProviderModalOpen(false);
	};

	const closeDeleteProviderModal = () => {
		setIsDeleteProviderModalOpen(false);
	};

	return (
		<div className="bg-base-100 rounded-xl shadow-lg p-4 mb-4">
			<div className="grid grid-cols-12 gap-4 items-center">
				{/* Provider Title*/}
				<div className="col-span-3 flex items-center space-x-4">
					<h3 className="text-sm font-medium capitalize">{provider}</h3>
				</div>
				{/* Enable/Disable Toggle */}
				<div className="col-span-3 flex items-center space-x-4 ml-1">
					<label className="text-sm font-medium">Enable</label>
					<input
						type="checkbox"
						checked={isEnabled}
						onChange={toggleProviderEnable}
						className="toggle toggle-accent rounded-full"
						spellCheck="false"
					/>
				</div>
				{/* Full Settings with Chevron */}
				<div className="col-span-6 cursor-pointer space-x-4 flex items-end justify-end" onClick={toggleExpand}>
					<div className="flex items-center">
						<span className="text-sm font-medium">API Key</span>
						{localSettings.apiKey ? (
							<FiCheckCircle className="text-green-500 mx-1" title="API Key Configured" />
						) : (
							<FiXCircle className="text-red-500 mx-1" title="API Key Not Configured" />
						)}
					</div>
					<div className="flex items-center">
						<span className="text-sm font-medium">Full Settings</span>
						{isExpanded ? (
							<FiChevronUp size={16} className="mx-1 text-neutral/60" />
						) : (
							<FiChevronDown size={16} className="mx-1 text-neutral/60" />
						)}
					</div>
				</div>
			</div>

			{isEnabled && isExpanded && (
				<div className="m-1 mt-8 space-y-4">
					{/* API Key */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm text-left tooltip" data-tip={ProviderInfoDescription['apiKey']}>
							API Key
						</label>
						<input
							type="password"
							className="input col-span-9 w-full h-10 rounded-xl border border-neutral/20 px-4 py-2"
							style={{ fontSize: '14px' }}
							value={localSettings.apiKey}
							onChange={e => {
								setLocalSettings({
									...localSettings,
									apiKey: e.target.value,
								});
							}}
							onBlur={async e => {
								updateLocalSettings('apiKey', e.target.value);
								await SetAISettingAPIKey(provider, e.target.value);
							}}
							spellCheck="false"
						/>
					</div>

					{/* Origin */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm text-left tooltip" data-tip={ProviderInfoDescription['origin']}>
							Origin
						</label>
						<input
							type="text"
							className="input col-span-9 w-full h-10 rounded-xl border border-neutral/20 px-4 py-2"
							style={{ fontSize: '14px' }}
							value={localSettings.origin}
							onChange={e => {
								setLocalSettings({
									...localSettings,
									origin: e.target.value,
								});
							}}
							onBlur={async e => {
								updateLocalSettings('origin', e.target.value);
								await SetAISettingAttrs(provider, { origin: e.target.value } as AISettingAttrs);
							}}
							spellCheck="false"
						/>
					</div>

					{/* Chat completion path prefix */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label
							className="col-span-3 text-sm text-left tooltip"
							data-tip={ProviderInfoDescription['chatCompletionPathPrefix']}
						>
							Chat Path Prefix
						</label>
						<input
							type="text"
							className="input col-span-9 w-full h-10 rounded-xl border border-neutral/20 px-4 py-2"
							style={{ fontSize: '14px' }}
							value={localSettings.chatCompletionPathPrefix}
							onChange={e => {
								setLocalSettings({
									...localSettings,
									chatCompletionPathPrefix: e.target.value,
								});
							}}
							onBlur={async e => {
								updateLocalSettings('chatCompletionPathPrefix', e.target.value);
								await SetAISettingAttrs(provider, { chatCompletionPathPrefix: e.target.value } as AISettingAttrs);
							}}
							spellCheck="false"
						/>
					</div>

					{/* Models : Default and add */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm text-left tooltip" data-tip={ProviderInfoDescription['defaultModel']}>
							Default Model
						</label>
						<div className="col-span-6">
							<Dropdown<ModelName>
								dropdownItems={modelSettings}
								selectedKey={localSettings.defaultModel}
								onChange={async (modelName: ModelName) => {
									updateLocalSettings('defaultModel', modelName);
									await SetAISettingAttrs(provider, { defaultModel: modelName } as AISettingAttrs);
								}}
								filterDisabled={true}
								title="Select Default Model"
								getDisplayName={key => modelSettings[key].displayName}
							/>
						</div>
						<div className="col-span-3 flex justify-end">
							<button className="btn btn-md btn-ghost rounded-xl flex items-center" onClick={handleAddModel}>
								<FiPlus size={16} /> Add Model
							</button>
						</div>
					</div>

					{/* Models Table */}
					<div className="overflow-x-auto border border-base-content/10 rounded-2xl">
						<table className="table table-zebra w-full">
							<thead>
								<tr className="font-semibold text-sm px-4 py-0 m-0 bg-base-300">
									<th className="text-base-content">Model Name</th>
									<th className="text-base-content text-center">Enabled</th>
									<th className="text-base-content text-center">Reasoning</th>
									<th className="text-base-content text-right pr-8">Actions</th>
								</tr>
							</thead>
							<tbody>
								{Object.entries(modelSettings).map(([modelName, model]) => (
									<tr key={modelName} className="hover:bg-base-300 border-none shadow-none">
										<td>{model.displayName || modelName}</td>
										<td className="flex items-center justify-center">
											<input
												type="checkbox"
												checked={model.isEnabled}
												onChange={async () => {
													await toggleModelEnable(modelName);
												}}
												className="toggle toggle-accent rounded-full"
											/>
										</td>
										<td>
											<div className="flex items-center justify-center">
												{isModelReasoningSupport(modelName) ? <FiCheck size={16} /> : <FiX size={16} />}
											</div>
										</td>
										<td className="text-right">
											<button
												className="btn btn-sm btn-ghost rounded-2xl"
												aria-label="Edit Model"
												title="Edit Model"
												onClick={() => {
													handleEditModel(modelName);
												}}
											>
												<FiEdit size={16} />
											</button>
											<button
												className="btn btn-sm btn-ghost rounded-2xl"
												aria-label="Delete Model"
												onClick={() => {
													handleDeleteModel(modelName);
												}}
												disabled={!isModelRemovable(modelName)}
												title={!isModelRemovable(modelName) ? 'Cannot Delete Default Or Inbuilt Model' : 'Delete Model'}
											>
												<FiTrash2 size={16} />
											</button>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
					{/* End of Models */}

					{/* Delete Provider */}
					<div className="flex justify-end">
						<button
							className="btn btn-md btn-ghost rounded-2xl flex items-center"
							aria-label="Delete Provider"
							onClick={() => {
								handleDeleteProvider(provider);
							}}
							disabled={!isProviderRemovable(provider)}
							title={!isProviderRemovable(provider) ? 'Cannot Delete Default Or Inbuilt Provider' : 'Delete Provider'}
						>
							<FiTrash2 size={16} /> Delete Provider
						</button>
					</div>
				</div>
			)}

			{showModal && (
				<dialog className="modal modal-open">
					<div className="modal-box w-5/6 max-w-4xl">
						<div className="flex flex-row items-center">
							<FiAlertTriangle size={24} />
							<p className="text-lg px-4">Cannot disable the last provider !!!</p>
						</div>
					</div>
					<form method="dialog" className="modal-backdrop w-full">
						<button
							onClick={() => {
								setShowModal(false);
							}}
						>
							OK
						</button>
					</form>
				</dialog>
			)}

			{/* Modify Model Modal */}
			{isModifyModelModalOpen && (
				<ModifyModelModal
					isOpen={isModifyModelModalOpen}
					onClose={() => {
						setIsModifyModelModalOpen(false);
					}}
					onSubmit={handleModifyModelSubmit}
					providerName={provider}
					initialModelName={selectedModelName || undefined}
					initialData={selectedModelName ? modelSettings[selectedModelName] : undefined}
					existingModels={modelSettings}
				/>
			)}

			{/* Delete Model Confirmation Modal */}
			{isDeleteModelModalOpen && (
				<DeleteConfirmationModal
					isOpen={isDeleteModelModalOpen}
					onClose={closeDeleteModelModal}
					onConfirm={handleDeleteModelConfirm}
					title="Delete Model"
					message={`Are you sure you want to delete the model "${selectedModelName || ''}"? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>
			)}

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

			{/* ADDED: Delete Provider Modal */}
			{isDeleteProviderModalOpen && (
				<DeleteConfirmationModal
					isOpen={isDeleteProviderModalOpen}
					onClose={closeDeleteProviderModal}
					onConfirm={handleDeleteProviderConfirm}
					title="Delete Provider"
					message={`Are you sure you want to delete the entire provider "${provider}"? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>
			)}
		</div>
	);
};

export default AISettingsCard;
