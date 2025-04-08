import { providerSetAPI, settingstoreAPI } from '@/apis/baseapi';
import { UpdateProviderAISettings } from '@/apis/settingstore_helper';
import DeleteConfirmationModal from '@/components/delete_confirmation';
import type { ConfigurationResponse, ModelName, ProviderName } from '@/models/aiprovidermodel';
import { ProviderInfoDescription } from '@/models/aiprovidermodel';
import type { AISetting, ModelSetting } from '@/models/settingmodel';
import ModelDropdown from '@/settings/model_dropdown';

import ActionDeniedAlert from '@/components/action_denied';
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
import ModifyModelModal from './model_modify_modal';

interface AISettingsCardProps {
	provider: ProviderName;
	settings: AISetting;
	aiSettings: Record<string, AISetting>;
	defaultProvider: ProviderName;
	onProviderSettingChange: (provider: ProviderName, settings: AISetting) => void;
}

const AISettingsCard: FC<AISettingsCardProps> = ({
	provider,
	settings,
	aiSettings,
	defaultProvider,
	onProviderSettingChange,
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
	const [configurationInfo, setConfigurationInfo] = useState<ConfigurationResponse | null>(null);

	useEffect(() => {
		const fetchConfigurationInfo = async () => {
			try {
				const info = await providerSetAPI.getConfigurationInfo();
				if (!info || info.defaultProvider === '' || Object.keys(info.configuredProviders).length === 0) {
					return;
				}
				setConfigurationInfo(info);
			} catch (error) {
				console.error('Failed to fetch configuration info:', error);
			}
		};

		fetchConfigurationInfo();
	}, []);

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

	const toggleEnable = () => {
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
		handleSettingChange('isEnabled', newIsEnabled);
	};

	// Handle setting changes and save to backend
	const handleSettingChange = async (key: string, value: any) => {
		const updatedSettings = {
			...localSettings,
			[key]: value,
		};
		setLocalSettings(updatedSettings);
		await settingstoreAPI.setSetting(`aiSettings.${provider}.${key}`, value);
		UpdateProviderAISettings(provider, updatedSettings);
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
		if (modelName === localSettings.defaultModel) {
			return true;
		}
		if (configurationInfo && configurationInfo.configuredProviders[provider]?.models[modelName]) {
			return true;
		}

		return false;
	};

	const handleDeleteModel = (modelName: ModelName) => {
		if (isModelRemovable(modelName)) {
			setActionDeniedMessage('Cannot delete the default model. Please select a different default model first.');
			setShowActionDeniedAlert(true);
			return;
		}
		setSelectedModelName(modelName);
		setIsDeleteModelModalOpen(true);
	};

	const handleDeleteModelConfirm = () => {
		const modelName: ModelName = selectedModelName || '';
		if (isModelRemovable(modelName)) {
			setActionDeniedMessage('Cannot delete the default model. Please select a different default model first.');
			setShowActionDeniedAlert(true);
			return;
		}

		const updatedModels = selectedModelName
			? Object.fromEntries(Object.entries(modelSettings).filter(([key]) => key !== selectedModelName))
			: { ...modelSettings };

		setModelSettings(updatedModels);
		handleSettingChange('modelSettings', updatedModels);
		setIsDeleteModelModalOpen(false);
		setSelectedModelName(null);
	};

	const closeDeleteModelModal = () => {
		setIsDeleteModelModalOpen(false);
		setSelectedModelName(null);
	};

	const handleModifyModelSubmit = (modelName: ModelName, modelData: ModelSetting) => {
		const updatedModels = { ...modelSettings };
		updatedModels[modelName] = modelData;
		setModelSettings(updatedModels);
		handleSettingChange('modelSettings', updatedModels);
		setIsModifyModelModalOpen(false);
		setSelectedModelName(null);
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
						onChange={toggleEnable}
						className="toggle toggle-primary rounded-full"
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
							<FiChevronUp size={16} className="mx-1 text-gray-500" />
						) : (
							<FiChevronDown size={16} className="mx-1 text-gray-500" />
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
							className="input col-span-9 w-full h-10 rounded-xl border border-base-300 px-4 py-2"
							style={{ fontSize: '14px' }}
							value={localSettings.apiKey}
							onChange={e => {
								setLocalSettings({
									...localSettings,
									apiKey: e.target.value,
								});
							}}
							onBlur={e => {
								handleSettingChange('apiKey', e.target.value);
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
							className="input col-span-9 w-full h-10 rounded-xl border border-base-300 px-4 py-2"
							style={{ fontSize: '14px' }}
							value={localSettings.origin}
							onChange={e => {
								setLocalSettings({
									...localSettings,
									origin: e.target.value,
								});
							}}
							onBlur={e => {
								handleSettingChange('origin', e.target.value);
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
							className="input col-span-9 w-full h-10 rounded-xl border border-base-300 px-4 py-2"
							style={{ fontSize: '14px' }}
							value={localSettings.chatCompletionPathPrefix}
							onChange={e => {
								setLocalSettings({
									...localSettings,
									origin: e.target.value,
								});
							}}
							onBlur={e => {
								handleSettingChange('chatCompletionPathPrefix', e.target.value);
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
							<ModelDropdown
								modelSettings={modelSettings}
								defaultModel={localSettings.defaultModel}
								onModelChange={modelName => {
									handleSettingChange('defaultModel', modelName);
								}}
							/>
						</div>
						<div className="col-span-3 flex justify-end">
							<button className="btn btn-md btn-ghost rounded-xl flex items-center" onClick={handleAddModel}>
								<FiPlus size={16} /> Add Model
							</button>
						</div>
					</div>

					{/* Models Table */}
					<div className="overflow-x-auto">
						<table className="table table-zebra w-full">
							<thead>
								<tr className="font-semibold text-sm px-4 py-0 m-0 bg-base-300">
									<th className="rounded-tl-2xl">Model Name</th>
									<th className="text-center">Enabled</th>
									<th className="text-center">Reasoning</th>
									<th className="text-right rounded-tr-2xl pr-8">Actions</th>
								</tr>
							</thead>
							<tbody>
								{Object.entries(modelSettings).map(([modelName, model], index, array) => (
									<tr key={modelName} className="hover:bg-base-300 border-none shadow-none">
										<td className={index === array.length - 1 ? 'rounded-bl-2xl' : ''}>
											{model.displayName || modelName}
										</td>
										<td className="flex items-center justify-center">
											<input
												type="checkbox"
												checked={model.isEnabled}
												onChange={() => {
													const updatedModels = { ...modelSettings };
													updatedModels[modelName] = { ...model, isEnabled: !model.isEnabled };
													setModelSettings(updatedModels);
													handleSettingChange('modelSettings', updatedModels);
												}}
												className="toggle toggle-primary rounded-full"
											/>
										</td>
										<td>
											<div className="flex items-center justify-center">
												{model.reasoningSupport ? <FiCheck size={16} /> : <FiX size={16} />}
											</div>
										</td>
										<td className={index === array.length - 1 ? 'rounded-br-2xl text-right' : 'text-right'}>
											<button
												className="btn btn-sm btn-ghost rounded-2xl"
												aria-label="Edit Model"
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
												disabled={isModelRemovable(modelName)}
												title={isModelRemovable(modelName) ? 'Cannot delete default model' : 'Delete model'}
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
		</div>
	);
};

export default AISettingsCard;
