import type { FC } from 'react';
import { useEffect, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp, FiEdit, FiPlus, FiTrash2, FiX } from 'react-icons/fi';

import type { ProviderPreset } from '@/models/modelpresetsmodel';
import { type ModelPreset, type ModelPresetID, type ProviderName } from '@/models/modelpresetsmodel';

import { modelPresetStoreAPI } from '@/apis/baseapi';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';
import Dropdown from '@/components/dropdown';

import ModifyModelModal from '@/modelconfig/model_modify_modal';

interface ProviderPresetCardProps {
	provider: ProviderName;
	isEnabled: boolean; // comes from aiSettings[provider].isEnabled
	preset: ProviderPreset;
	inbuiltProviderPresets?: Record<ModelPresetID, ModelPreset>;

	onPresetChange: (provider: ProviderName, newPreset: ProviderPreset) => void;
	onProviderDelete: (provider: ProviderName) => Promise<void>;
}

const ProviderPresetCard: FC<ProviderPresetCardProps> = ({
	provider,
	isEnabled,
	preset,
	inbuiltProviderPresets,
	onPresetChange,
	onProviderDelete,
}) => {
	/* ── local state ──────────────────────────────────────────── */
	const [isExpanded, setIsExpanded] = useState(false);
	const [modelPresets, setModelPresets] = useState<Record<ModelPresetID, ModelPreset>>(preset.modelPresets);
	const [defaultModelID, setDefaultModelID] = useState<ModelPresetID>(preset.defaultModelPresetID);

	const [selectedModelID, setSelectedModelID] = useState<ModelPresetID | null>(null);
	const [isModifyModalOpen, setIsModifyModalOpen] = useState(false);
	const [isDeleteModelModalOpen, setIsDeleteModelModalOpen] = useState(false);

	const [showActionDenied, setShowActionDenied] = useState(false);
	const [actionDeniedMsg, setActionDeniedMsg] = useState('');

	const [isDeleteProviderModalOpen, setIsDeleteProviderModalOpen] = useState(false);

	/* sync props → state */
	useEffect(() => {
		setModelPresets(preset.modelPresets);
		setDefaultModelID(preset.defaultModelPresetID);
	}, [preset]);

	/* ── helpers ─────────────────────────────────────────────── */
	const isModelRemovable = (id: ModelPresetID) => {
		if (id === defaultModelID) return false;
		if (inbuiltProviderPresets && id in inbuiltProviderPresets) return false;
		return true;
	};

	const isModelReasoning = (id: ModelPresetID) => {
		if (id in modelPresets && 'reasoning' in modelPresets[id] && modelPresets[id].reasoning !== undefined) {
			return true;
		}
		if (
			inbuiltProviderPresets &&
			id in inbuiltProviderPresets &&
			'reasoning' in inbuiltProviderPresets[id] &&
			inbuiltProviderPresets[id].reasoning !== undefined
		) {
			return true;
		}
		return false;
	};

	/* ── UI handlers ─────────────────────────────────────────── */
	const toggleExpand = () => {
		if (isEnabled) setIsExpanded(prev => !prev);
	};

	const handleAddModel = () => {
		setSelectedModelID(null);
		setIsModifyModalOpen(true);
	};

	const handleEditModel = (id: ModelPresetID) => {
		setSelectedModelID(id);
		setIsModifyModalOpen(true);
	};

	const handleModifyModelSubmit = async (id: ModelPresetID, data: ModelPreset) => {
		try {
			const newMap = { ...modelPresets, [id]: data };
			setModelPresets(newMap);
			onPresetChange(provider, { ...preset, modelPresets: newMap });
			await modelPresetStoreAPI.addModelPreset(provider, id, data);
			setIsModifyModalOpen(false);
			setSelectedModelID(null);
		} catch (err) {
			console.error('Failed to save model preset:', err, (err as Error).stack || '');
			setActionDeniedMsg('Failed to save model preset. Please try again.');
			setShowActionDenied(true);
		}
	};

	/* enable / disable model */
	const toggleModelEnable = async (id: ModelPresetID) => {
		if (id === defaultModelID && modelPresets[id].isEnabled) {
			setActionDeniedMsg('Cannot disable the default model. Choose another default first.');
			setShowActionDenied(true);
			return;
		}

		try {
			const updated = { ...modelPresets[id], isEnabled: !modelPresets[id].isEnabled };
			const newMap = { ...modelPresets, [id]: updated };
			setModelPresets(newMap);
			onPresetChange(provider, { ...preset, modelPresets: newMap });
			await modelPresetStoreAPI.addModelPreset(provider, id, updated);
		} catch (err) {
			console.error('Failed to update model enable state:', err, (err as Error).stack || '');
			setActionDeniedMsg('Failed to update model enable state. Please try again.');
			setShowActionDenied(true);
		}
	};

	/* deletion */
	const handleDeleteModelRequest = (id: ModelPresetID) => {
		if (!isModelRemovable(id)) {
			setActionDeniedMsg('Cannot delete the default or in-built model.');
			setShowActionDenied(true);
			return;
		}
		setSelectedModelID(id);
		setIsDeleteModelModalOpen(true);
	};

	const confirmDeleteModel = async () => {
		if (!selectedModelID) return;
		if (!isModelRemovable(selectedModelID)) return;

		try {
			const newMap = Object.fromEntries(Object.entries(modelPresets).filter(([k]) => k !== selectedModelID));
			setModelPresets(newMap);
			onPresetChange(provider, { ...preset, modelPresets: newMap });

			await modelPresetStoreAPI.deleteModelPreset(provider, selectedModelID);

			setIsDeleteModelModalOpen(false);
			setSelectedModelID(null);
		} catch (err) {
			console.error('Failed to delete model:', err, (err as Error).stack || '');
			setActionDeniedMsg('Failed to delete model. Please try again.');
			setShowActionDenied(true);
		}
	};

	/* default model change */
	const handleDefaultModelChange = async (id: ModelPresetID) => {
		try {
			setDefaultModelID(id);
			onPresetChange(provider, { ...preset, defaultModelPresetID: id });
			await modelPresetStoreAPI.setDefaultModelPreset(provider, id);
		} catch (err) {
			console.error('Failed to set default model:', err, (err as Error).stack || '');
			setActionDeniedMsg('Failed to set default model. Please try again.');
			setShowActionDenied(true);
		}
	};

	const confirmDeleteProvider = async () => {
		try {
			await onProviderDelete(provider);
			setIsDeleteProviderModalOpen(false);
		} catch (err) {
			console.error('Failed to delete provider:', err, (err as Error).stack || '');
			setActionDeniedMsg('Failed to delete provider. Please try again.');
			setShowActionDenied(true);
		}
	};

	/* ── RENDER ──────────────────────────────────────────────── */
	return (
		<div className="bg-base-100 rounded-xl shadow-lg px-4 py-2 mb-8">
			{/* Header */}
			<div className="grid grid-cols-12 gap-2 items-center">
				{/* Provider Title*/}
				<div className="col-span-2 flex items-center space-x-4">
					<h3 className="text-sm font-semibold capitalize">{provider}</h3>
				</div>

				{/* Provider status */}
				<div className="col-span-2 text-sm font-medium">
					<span className={isEnabled ? 'text-success' : 'text-error'}>{isEnabled ? 'Enabled' : 'Disabled'}</span>
				</div>

				{/* Default model dropdown */}
				<div className="flex items-center gap-x-2 col-span-6">
					<label className="text-sm whitespace-nowrap">Default</label>
					<Dropdown<ModelPresetID>
						dropdownItems={modelPresets}
						selectedKey={defaultModelID}
						onChange={handleDefaultModelChange}
						filterDisabled={true}
						title="Select Default Model"
						getDisplayName={k => modelPresets[k].displayName}
					/>
				</div>

				{/* Chevron */}
				<div className="col-span-2 flex justify-end items-center cursor-pointer gap-1" onClick={toggleExpand}>
					<label className="text-sm whitespace-nowrap">All Models</label>
					{isExpanded ? <FiChevronUp /> : <FiChevronDown />}
				</div>
			</div>

			{/* Body – models table */}
			{isExpanded && (
				<div className="mt-8 space-y-4">
					{/* Models table */}
					<div className="overflow-x-auto border border-base-content/10 rounded-2xl">
						<table className="table table-zebra w-full">
							<thead>
								<tr className="text-sm font-semibold bg-base-300">
									<th>Model Name</th>
									<th className="text-center">Enabled</th>
									<th className="text-center">Reasoning</th>
									<th className="text-right pr-8">Actions</th>
								</tr>
							</thead>
							<tbody>
								{Object.entries(modelPresets).map(([id, m]) => (
									<tr key={id} className="hover:bg-base-300">
										<td>{m.displayName || id}</td>
										<td className="flex justify-center">
											<input
												type="checkbox"
												checked={m.isEnabled}
												onChange={() => toggleModelEnable(id)}
												className="toggle toggle-accent rounded-full"
											/>
										</td>
										<td className="text-center">{isModelReasoning(id) ? <FiCheck /> : <FiX />}</td>
										<td className="text-right">
											{/* edit */}
											<button
												className="btn btn-sm btn-ghost rounded-2xl"
												onClick={() => {
													handleEditModel(id);
												}}
												title="Edit Model"
											>
												<FiEdit size={16} />
											</button>

											{/* delete */}
											<button
												className="btn btn-sm btn-ghost rounded-2xl"
												onClick={() => {
													handleDeleteModelRequest(id);
												}}
												disabled={!isModelRemovable(id)}
												title={isModelRemovable(id) ? 'Delete Model' : 'Cannot delete default or in-built model'}
											>
												<FiTrash2 size={16} />
											</button>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>

					{/* Add-model button */}
					<div className="flex justify-end">
						<button className="btn btn-ghost rounded-xl flex items-center" onClick={handleAddModel}>
							<FiPlus /> <span className="ml-1">Add Model</span>
						</button>
					</div>
				</div>
			)}

			{/* ── Dialogs / Alerts ─────────────────────────────── */}
			{isModifyModalOpen && (
				<ModifyModelModal
					isOpen={isModifyModalOpen}
					onClose={() => {
						setIsModifyModalOpen(false);
					}}
					onSubmit={handleModifyModelSubmit}
					providerName={provider}
					initialModelID={selectedModelID || undefined}
					initialData={selectedModelID ? modelPresets[selectedModelID] : undefined}
					existingModels={modelPresets}
				/>
			)}

			{isDeleteModelModalOpen && (
				<DeleteConfirmationModal
					isOpen={isDeleteModelModalOpen}
					onClose={() => {
						setIsDeleteModelModalOpen(false);
					}}
					onConfirm={confirmDeleteModel}
					title="Delete Model"
					message={`Delete model "${selectedModelID ?? ''}"? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>
			)}

			{showActionDenied && (
				<ActionDeniedAlert
					isOpen={showActionDenied}
					onClose={() => {
						setShowActionDenied(false);
						setActionDeniedMsg('');
					}}
					message={actionDeniedMsg}
				/>
			)}

			{isDeleteProviderModalOpen && (
				<DeleteConfirmationModal
					isOpen={isDeleteProviderModalOpen}
					onClose={() => {
						setIsDeleteProviderModalOpen(false);
					}}
					onConfirm={confirmDeleteProvider}
					title="Delete Provider"
					message={`Really delete provider "${provider}" (and all its model presets)?`}
					confirmButtonText="Delete"
				/>
			)}
		</div>
	);
};

export default ProviderPresetCard;
