import { settingstoreAPI } from '@/backendapibase';
import { updateProviderAISettings } from '@/backendapihelper/settings_helper';
import DeleteConfirmationModal from '@/components/delete_confirmation';
import type { ProviderName } from '@/models/aiprovidermodel';
import type { AISetting, ModelSetting } from '@/models/settingmodel';
import ModelDropdown from '@/settings/model_dropdown';
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
import ModifyModelModal from './model_modal';

interface AISettingsCardProps {
	provider: ProviderName;
	settings: AISetting;
	aiSettings: Record<string, AISetting>;
}

const AISettingsCard: FC<AISettingsCardProps> = ({ provider, settings, aiSettings }) => {
	const [isExpanded, setIsExpanded] = useState(false);
	const [localSettings, setLocalSettings] = useState<AISetting>(settings);
	const [showModal, setShowModal] = useState(false);
	const [isModifyModelModalOpen, setIsModifyModelModalOpen] = useState(false);
	const [selectedModel, setSelectedModel] = useState<ModelSetting | null>(null);
	const [isDeleteModelModalOpen, setIsDeleteModelModalOpen] = useState(false);

	useEffect(() => {
		setLocalSettings(settings);
	}, [settings]);

	const toggleExpand = () => {
		if (localSettings.isEnabled) {
			setIsExpanded(!isExpanded);
		}
	};

	const saveSetting = async (key: string, value: any) => {
		const updatedSettings = { ...localSettings, [key]: value };
		setLocalSettings(updatedSettings);
		await settingstoreAPI.setSetting(`aiSettings.${provider}.${key}`, value);
		updateProviderAISettings(provider, updatedSettings);
	};

	const toggleEnable = () => {
		const newIsEnabled = !localSettings.isEnabled;
		if (!newIsEnabled) {
			const enabledProviders = Object.keys(aiSettings).filter(k => aiSettings[k].isEnabled && k !== provider);
			if (enabledProviders.length === 0) {
				setShowModal(true);
				return;
			}
		}
		saveSetting('isEnabled', newIsEnabled);
	};

	const handleModifyModelSubmit = async (modelData: ModelSetting) => {
		const existing = localSettings.modelSettings.find(m => m.name === modelData.name);
		const updatedModels = existing
			? localSettings.modelSettings.map(m => (m.name === modelData.name ? modelData : m))
			: [...localSettings.modelSettings, modelData];

		saveSetting('modelSettings', updatedModels);
		setIsModifyModelModalOpen(false);
		setSelectedModel(null);
	};

	const handleDeleteModelConfirm = async () => {
		const updatedModels = localSettings.modelSettings.filter(m => m.name !== selectedModel?.name);
		saveSetting('modelSettings', updatedModels);
		setIsDeleteModelModalOpen(false);
		setSelectedModel(null);
	};

	return (
		<div className="bg-base-100 rounded-lg shadow-lg p-4 mb-4">
			<div className="grid grid-cols-12 gap-4 items-center">
				<div className="col-span-3 flex items-center space-x-4">
					<h3 className="text-sm font-medium capitalize">{provider}</h3>
				</div>
				<div className="col-span-3 flex items-center space-x-4 ml-1">
					<label className="text-sm font-medium">Enable</label>
					<input
						type="checkbox"
						checked={localSettings.isEnabled}
						onChange={toggleEnable}
						className="toggle toggle-primary rounded-full"
					/>
				</div>
				<div className="col-span-6 cursor-pointer space-x-4 flex items-end justify-end" onClick={toggleExpand}>
					<div className="flex items-center">
						<span className="text-sm font-medium">API Key</span>
						{localSettings.apiKey ? (
							<FiCheckCircle className="text-green-500 mx-1" />
						) : (
							<FiXCircle className="text-red-500 mx-1" />
						)}
					</div>
					<div className="flex items-center">
						<span className="text-sm font-medium">Full Settings</span>
						{isExpanded ? <FiChevronUp size={16} className="mx-1" /> : <FiChevronDown size={16} className="mx-1" />}
					</div>
				</div>
			</div>

			{localSettings.isEnabled && isExpanded && (
				<div className="m-1 mt-8 space-y-4">
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm">API Key</label>
						<input
							type="password"
							className="input col-span-9"
							value={localSettings.apiKey}
							onChange={e => setLocalSettings({ ...localSettings, apiKey: e.target.value })}
							onBlur={e => saveSetting('apiKey', e.target.value)}
						/>
					</div>

					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm">Origin</label>
						<input
							type="text"
							className="input col-span-9"
							value={localSettings.origin}
							onChange={e => setLocalSettings({ ...localSettings, origin: e.target.value })}
							onBlur={e => saveSetting('origin', e.target.value)}
						/>
					</div>

					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm">Default Model</label>
						<div className="col-span-6">
							<ModelDropdown
								modelSettings={localSettings.modelSettings}
								defaultModel={localSettings.defaultModel}
								onModelChange={modelName => saveSetting('defaultModel', modelName)}
							/>
						</div>
						<div className="col-span-3 flex justify-end">
							<button
								className="btn btn-md btn-ghost"
								onClick={() => {
									setSelectedModel(null);
									setIsModifyModelModalOpen(true);
								}}
							>
								<FiPlus size={16} /> Add Model
							</button>
						</div>
					</div>

					<div className="overflow-x-auto">
						<table className="table table-zebra w-full">
							<thead>
								<tr className="font-semibold text-sm bg-base-300">
									<th>Model Name</th>
									<th className="text-center">Enabled</th>
									<th className="text-center">Reasoning</th>
									<th className="text-right pr-8">Actions</th>
								</tr>
							</thead>
							<tbody>
								{localSettings.modelSettings.map(model => (
									<tr key={model.name}>
										<td>{model.name}</td>
										<td className="text-center">
											<input
												type="checkbox"
												checked={model.isEnabled}
												onChange={() => {
													const updatedModels = localSettings.modelSettings.map(m =>
														m.name === model.name ? { ...m, isEnabled: !m.isEnabled } : m
													);
													saveSetting('modelSettings', updatedModels);
												}}
												className="toggle toggle-primary rounded-full"
											/>
										</td>
										<td className="text-center">{model.reasoningSupport ? <FiCheck /> : <FiX />}</td>
										<td className="text-right">
											<button
												className="btn btn-sm btn-ghost"
												onClick={() => {
													setSelectedModel(model);
													setIsModifyModelModalOpen(true);
												}}
											>
												<FiEdit />
											</button>
											<button
												className="btn btn-sm btn-ghost"
												onClick={() => {
													setSelectedModel(model);
													setIsDeleteModelModalOpen(true);
												}}
											>
												<FiTrash2 />
											</button>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				</div>
			)}

			{showModal && (
				<dialog className="modal modal-open">
					<div className="modal-box">
						<FiAlertTriangle size={24} />
						<p>Cannot disable the last provider!</p>
					</div>
					<form method="dialog" className="modal-backdrop">
						<button onClick={() => setShowModal(false)}>OK</button>
					</form>
				</dialog>
			)}

			{isModifyModelModalOpen && (
				<ModifyModelModal
					isOpen={isModifyModelModalOpen}
					onClose={() => setIsModifyModelModalOpen(false)}
					onSubmit={handleModifyModelSubmit}
					initialData={selectedModel || undefined}
					existingModels={localSettings.modelSettings}
				/>
			)}

			{isDeleteModelModalOpen && (
				<DeleteConfirmationModal
					isOpen={isDeleteModelModalOpen}
					onClose={() => setIsDeleteModelModalOpen(false)}
					onConfirm={handleDeleteModelConfirm}
					title="Delete Model"
					message={`Delete "${selectedModel?.name}"?`}
					confirmButtonText="Delete"
				/>
			)}
		</div>
	);
};

export default AISettingsCard;
