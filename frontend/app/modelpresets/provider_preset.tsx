import type { FC } from 'react';
import { useEffect, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp, FiEdit, FiPlus, FiTrash2, FiX } from 'react-icons/fi';

import type { ProviderPreset } from '@/models/aimodelmodel';
import { type ModelPreset, type ModelPresetID, type ProviderName } from '@/models/aimodelmodel';

import { modelPresetStoreAPI } from '@/apis/baseapi';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';
import Dropdown from '@/components/dropdown';

import ModifyModelModal from '@/modelpresets/model_modify_modal';

interface ProviderPresetCardProps {
	provider: ProviderName;
	isEnabled: boolean; // comes from aiSettings[provider].isEnabled
	preset: ProviderPreset;
	inbuiltProviderModels?: Record<ModelPresetID, ModelPreset>;

	defaultProvider: ProviderName; // only for “is deletable?” logic
	onPresetChange: (provider: ProviderName, newPreset: ProviderPreset) => void;
	onProviderDelete: (provider: ProviderName) => Promise<void>;
}

const ProviderPresetCard: FC<ProviderPresetCardProps> = ({
	provider,
	isEnabled,
	preset,
	inbuiltProviderModels,
	defaultProvider,
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
		if (inbuiltProviderModels && id in inbuiltProviderModels) return false;
		return true;
	};

	const isModelReasoning = (id: ModelPresetID) => {
		if (id in modelPresets && 'reasoning' in modelPresets[id] && modelPresets[id].reasoning !== undefined) {
			return true;
		}
		if (
			inbuiltProviderModels &&
			id in inbuiltProviderModels &&
			'reasoning' in inbuiltProviderModels[id] &&
			inbuiltProviderModels[id].reasoning !== undefined
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

	/* provider deletion */
	const isProviderRemovable = !inbuiltProviderModels && provider !== defaultProvider;
	const requestDeleteProvider = () => {
		if (!isProviderRemovable) {
			setActionDeniedMsg('Cannot delete default or in-built provider.');
			setShowActionDenied(true);
			return;
		}
		setIsDeleteProviderModalOpen(true);
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
		<div className="bg-base-100 rounded-xl shadow-lg p-4 mb-4">
			{/* Header */}
			<div className="grid grid-cols-12 gap-4 items-center">
				{/* Provider name */}
				<div className="col-span-3 capitalize font-medium">{provider}</div>

				{/* Provider status */}
				<div className="col-span-3 text-sm">
					Status:{' '}
					<span className={isEnabled ? 'text-success' : 'text-error'}>{isEnabled ? 'Enabled' : 'Disabled'}</span>
				</div>

				{/* Default model dropdown */}
				<div className="col-span-4">
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
				<div className="col-span-2 flex justify-end cursor-pointer" onClick={toggleExpand}>
					{isExpanded ? <FiChevronUp /> : <FiChevronDown />}
				</div>
			</div>

			{/* Body – models table */}
			{isExpanded && (
				<div className="mt-8 space-y-4">
					{/* Add-model button */}
					<div className="flex justify-end">
						<button className="btn btn-ghost rounded-xl flex items-center" onClick={handleAddModel}>
							<FiPlus /> <span className="ml-1">Add Model</span>
						</button>
					</div>

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

					{/* Delete provider */}
					<div className="flex justify-end">
						<button
							className="btn btn-ghost rounded-2xl flex items-center"
							onClick={requestDeleteProvider}
							disabled={!isProviderRemovable}
							title={!isProviderRemovable ? 'Cannot delete default or in-built provider' : 'Delete Provider'}
						>
							<FiTrash2 size={16} /> Delete Provider
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
