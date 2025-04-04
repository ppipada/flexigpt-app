import DeleteConfirmationModal from '@/components/delete_confirmation';
import type { ProviderName } from '@/models/aiprovidermodel';
import { ProviderInfoDescription } from '@/models/aiprovidermodel';
import type { AISetting, ModelSetting } from '@/models/settingmodel';

import type { FC } from 'react';
import { useState } from 'react';
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
import ModifyModelModal from './modify_model_modal';

interface AISettingsCardProps {
	provider: ProviderName;
	settings: AISetting;
	onChange: (key: string, value: any) => void;
	onSave: (key: string, value: any) => Promise<void>;
	aiSettings: Record<string, AISetting>;
}

const AISettingsCard: FC<AISettingsCardProps> = ({ provider, settings, onChange, onSave, aiSettings }) => {
	const [isExpanded, setIsExpanded] = useState(false);
	const [isEnabled, setIsEnabled] = useState(!!settings.isEnabled);
	const [showModal, setShowModal] = useState(false);
	const initModelSettings = settings.modelSettings;

	const [modelSettings, setModelSettings] = useState<ModelSetting[]>(initModelSettings);
	const [isModifyModelModalOpen, setIsModifyModelModalOpen] = useState(false);
	const [selectedModel, setSelectedModel] = useState<ModelSetting | null>(null);
	const [isDeleteModelModalOpen, setIsDeleteModelModalOpen] = useState(false);

	console.log(JSON.stringify(modelSettings, null, 2));

	const toggleExpand = () => {
		if (isEnabled) {
			setIsExpanded(!isExpanded);
		}
	};

	const toggleEnable = () => {
		const newIsEnabled = !isEnabled;
		if (!newIsEnabled) {
			const enabledProviders = Object.keys(aiSettings).filter(k => aiSettings[k].isEnabled && k !== provider);
			if (enabledProviders.length === 0) {
				setShowModal(true);
				return;
			}
		}
		setIsEnabled(newIsEnabled);
		onChange('isEnabled', newIsEnabled);
		onSave('isEnabled', newIsEnabled);
	};

	// Handlers for models
	const handleAddModel = () => {
		setSelectedModel(null);
		setIsModifyModelModalOpen(true);
	};

	const handleEditModel = (model: ModelSetting) => {
		setSelectedModel(model);
		setIsModifyModelModalOpen(true);
	};

	const handleDeleteModel = (model: ModelSetting) => {
		setSelectedModel(model);
		setIsDeleteModelModalOpen(true);
	};

	const handleDeleteModelConfirm = () => {
		const updatedModels = modelSettings.filter(m => m.name !== selectedModel?.name);
		setModelSettings(updatedModels);
		onChange('modelSettings', updatedModels);
		onSave('modelSettings', updatedModels);
		setIsDeleteModelModalOpen(false);
		setSelectedModel(null);
	};

	const closeDeleteModelModal = () => {
		setIsDeleteModelModalOpen(false);
		setSelectedModel(null);
	};

	const handleModifyModelSubmit = (modelData: ModelSetting) => {
		let updatedModels;
		if (selectedModel) {
			// Edit existing model
			updatedModels = modelSettings.map(m => (m.name === selectedModel.name ? modelData : m));
		} else {
			// Add new model
			updatedModels = [...modelSettings, modelData];
		}
		setModelSettings(updatedModels);
		onChange('modelSettings', updatedModels);
		onSave('modelSettings', updatedModels);
		setIsModifyModelModalOpen(false);
		setSelectedModel(null);
	};

	return (
		<div className="bg-base-100 rounded-lg shadow-lg p-4 mb-4">
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
						{settings.apiKey ? (
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
							className="input col-span-9 w-full h-10 rounded-lg px-4 py-2"
							style={{ fontSize: '14px' }}
							value={settings.apiKey}
							onChange={e => {
								onChange('apiKey', e.target.value);
							}}
							onBlur={e => {
								onSave('apiKey', e.target.value);
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
							className="input col-span-9 w-full h-10 rounded-lg px-4 py-2"
							style={{ fontSize: '14px' }}
							value={settings.origin}
							onChange={e => {
								onChange('origin', e.target.value);
							}}
							onBlur={e => {
								onSave('origin', e.target.value);
							}}
							spellCheck="false"
						/>
					</div>

					{/* Models : Default and add */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm text-left tooltip" data-tip={ProviderInfoDescription['defaultModel']}>
							Default Model
						</label>
						<select
							className="col-span-6 select select-bordered w-full h-10 rounded-lg px-4 py-2 text-sm"
							value={settings.defaultModel}
							onChange={e => {
								onChange('defaultModel', e.target.value);
								onSave('defaultModel', e.target.value);
							}}
						>
							{modelSettings.map(model => (
								<option key={model.name} value={model.name}>
									{model.name}
								</option>
							))}
						</select>
						<div className="col-span-3 flex justify-end">
							<button className="btn btn-md btn-ghost rounded-2xl flex items-center" onClick={handleAddModel}>
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
								{modelSettings.map((model, index) => (
									<tr key={model.name} className="hover:bg-base-300 border-none shadow-none">
										<td className={index === modelSettings.length - 1 ? 'rounded-bl-2xl' : ''}>{model.name}</td>
										<td className="flex items-center justify-center">
											<input
												type="checkbox"
												checked={model.isEnabled}
												onChange={() => {
													const updatedModels = modelSettings.map(m =>
														m.name === model.name ? { ...m, isEnabled: !m.isEnabled } : m
													);
													setModelSettings(updatedModels);
													onChange('modelSettings', updatedModels);
													onSave('modelSettings', updatedModels);
												}}
												className="toggle toggle-primary rounded-full"
											/>
										</td>
										<td>
											<div className="flex items-center justify-center">
												{model.reasoningSupport ? <FiCheck size={16} /> : <FiX size={16} />}
											</div>
										</td>
										<td className={index === modelSettings.length - 1 ? 'rounded-br-2xl text-right' : 'text-right'}>
											<button
												className="btn btn-sm btn-ghost rounded-2xl"
												aria-label="Edit Model"
												onClick={() => {
													handleEditModel(model);
												}}
											>
												<FiEdit size={16} />
											</button>
											<button
												className="btn btn-sm btn-ghost rounded-2xl"
												aria-label="Delete Model"
												onClick={() => {
													handleDeleteModel(model);
												}}
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
					initialData={selectedModel || undefined}
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
					message={`Are you sure you want to delete the model "${selectedModel?.name || ''}"? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>
			)}
		</div>
	);
};

export default AISettingsCard;
