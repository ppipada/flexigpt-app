import type { FC } from 'react';
import { useEffect, useState } from 'react';

import {
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

import { type ModelPreset, type ModelPresetID, type ProviderName, type ProviderPreset } from '@/spec/modelpreset';

import { modelPresetStoreAPI } from '@/apis/baseapi';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';
import Dropdown from '@/components/dropdown';

import ModifyModelModal from '@/modelpresets/modelpreset_modify_modal';

interface Props {
	provider: ProviderName;
	preset: ProviderPreset;
	defaultProvider: ProviderName;
	authKeySet: boolean;
	enabledProviders: ProviderName[];

	onProviderPresetChange: (provider: ProviderName, newPreset: ProviderPreset) => void;
	onProviderDelete: (provider: ProviderName) => Promise<void>;
	onRequestEdit: (provider: ProviderName) => void; // open edit-modal
}

const ProviderPresetCard: FC<Props> = ({
	provider,
	preset,
	defaultProvider,
	authKeySet,
	enabledProviders,
	onProviderPresetChange,
	onProviderDelete,
	onRequestEdit,
}) => {
	/* local-state sync ─────────────────────────────────── */
	const [localPreset, setLocalPreset] = useState<ProviderPreset>(preset);
	useEffect(() => {
		setLocalPreset(preset);
	}, [preset]);

	const [expanded, setExpanded] = useState(false);

	const [selectedModelID, setSelectedModelID] = useState<ModelPresetID | null>(null);
	const [showModifyModal, setShowModifyModal] = useState(false);
	const [showDelProviderModal, setShowDelProviderModal] = useState(false);
	const [showDelModelModal, setShowDelModelModal] = useState(false);

	const [deniedMsg, setDeniedMsg] = useState('');
	const [showDenied, setShowDenied] = useState(false);

	const isLastEnabled = localPreset.isEnabled && enabledProviders.length === 1;
	const providerIsBuiltIn = localPreset.isBuiltIn;

	/* helper to update parent + local copy */
	const updateLocal = (updater: (p: ProviderPreset) => ProviderPreset) => {
		setLocalPreset(prev => {
			const upd = updater(prev);
			onProviderPresetChange(provider, upd);
			return upd;
		});
	};

	/* enable / disable provider ───────────────────────── */
	const toggleProviderEnable = async () => {
		if (!localPreset.isEnabled) {
			await modelPresetStoreAPI.patchProviderPreset(provider, true);
			updateLocal(p => ({ ...p, isEnabled: true }));
			return;
		}

		// disabling
		if (provider === defaultProvider) {
			setDeniedMsg('Cannot disable the default provider. Pick another default first.');
			setShowDenied(true);
			return;
		}
		if (isLastEnabled) {
			setDeniedMsg('Cannot disable the last enabled provider.');
			setShowDenied(true);
			return;
		}

		try {
			await modelPresetStoreAPI.patchProviderPreset(provider, false);
			updateLocal(p => ({ ...p, isEnabled: false }));
		} catch (err) {
			console.error(err);
			setDeniedMsg('Failed toggling provider.');
			setShowDenied(true);
		}
	};

	/* expand */
	const toggleExpand = () => {
		if (localPreset.isEnabled) setExpanded(p => !p);
	};

	/* delete provider */
	const requestDeleteProvider = () => {
		if (providerIsBuiltIn) {
			setDeniedMsg('Built-in providers cannot be deleted.');
			setShowDenied(true);
			return;
		}
		setShowDelProviderModal(true);
	};
	const confirmDeleteProvider = async () => {
		await onProviderDelete(provider);
		setShowDelProviderModal(false);
	};

	/* default model change */
	const handleDefaultModelChange = async (id: ModelPresetID) => {
		try {
			await modelPresetStoreAPI.patchProviderPreset(provider, undefined, id);
			updateLocal(p => ({ ...p, defaultModelPresetID: id }));
		} catch {
			setDeniedMsg('Failed setting default model.');
			setShowDenied(true);
		}
	};

	/* enable/disable model */
	const toggleModelEnable = async (id: ModelPresetID) => {
		const m = localPreset.modelPresets[id];
		if (id === localPreset.defaultModelPresetID && m.isEnabled) {
			setDeniedMsg('Cannot disable the default preset. Choose another default first.');
			setShowDenied(true);
			return;
		}
		try {
			await modelPresetStoreAPI.patchModelPreset(provider, id, !m.isEnabled);
			updateLocal(p => ({
				...p,
				modelPresets: {
					...p.modelPresets,
					[id]: { ...m, isEnabled: !m.isEnabled },
				},
			}));
		} catch {
			setDeniedMsg('Failed toggling model.');
			setShowDenied(true);
		}
	};

	/* add / edit model */
	const openAddModel = () => {
		if (providerIsBuiltIn) {
			setDeniedMsg('Cannot add presets to a built-in provider.');
			setShowDenied(true);
			return;
		}
		setSelectedModelID(null);
		setShowModifyModal(true);
	};
	const openEditModel = (id: ModelPresetID) => {
		if (localPreset.modelPresets[id].isBuiltIn) {
			setDeniedMsg('Built-in presets cannot be edited.');
			setShowDenied(true);
			return;
		}
		setSelectedModelID(id);
		setShowModifyModal(true);
	};

	const handleModifyModelSubmit = async (id: ModelPresetID, data: ModelPreset) => {
		try {
			// strip fields not in payload
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			const { id: _id, isBuiltIn, ...payload } = data;
			await modelPresetStoreAPI.putModelPreset(provider, id, payload);
			updateLocal(p => ({
				...p,
				modelPresets: { ...p.modelPresets, [id]: data },
			}));
			setShowModifyModal(false);
		} catch {
			setDeniedMsg('Failed saving preset.');
			setShowDenied(true);
		}
	};

	/* delete model */
	const requestDeleteModel = (id: ModelPresetID) => {
		if (localPreset.modelPresets[id].isBuiltIn) {
			setDeniedMsg('Built-in presets cannot be deleted.');
			setShowDenied(true);
			return;
		}
		setSelectedModelID(id);
		setShowDelModelModal(true);
	};
	const confirmDeleteModel = async () => {
		if (!selectedModelID) return;
		try {
			await modelPresetStoreAPI.deleteModelPreset(provider, selectedModelID);
			updateLocal(p => {
				const { [selectedModelID]: _, ...rest } = p.modelPresets;
				return { ...p, modelPresets: rest };
			});
			setShowDelModelModal(false);
		} catch {
			setDeniedMsg('Failed deleting preset.');
			setShowDenied(true);
		}
	};

	/* render ───────────────────────────────────────────── */
	const { modelPresets, defaultModelPresetID } = localPreset;

	return (
		<div className="bg-base-100 rounded-2xl shadow-lg px-4 py-2 mb-8">
			{/* header row */}
			<div className="grid grid-cols-12 gap-2 items-center">
				<div className="col-span-3">
					<h3 className="text-sm font-semibold capitalize">{localPreset.displayName || provider}</h3>
				</div>

				{/* enable toggle */}
				<div className="col-span-3 flex items-center gap-2">
					<label className="text-sm">Enable</label>
					<input
						type="checkbox"
						className="toggle toggle-accent rounded-full"
						checked={localPreset.isEnabled}
						onChange={toggleProviderEnable}
					/>
				</div>

				{/* api-key status + chevron */}
				<div className="col-span-6 flex items-end justify-end gap-4 cursor-pointer" onClick={toggleExpand}>
					<div className="flex items-center">
						<span className="text-sm">API-Key</span>
						{authKeySet ? <FiCheckCircle className="text-success mx-1" /> : <FiXCircle className="text-error mx-1" />}
					</div>

					<div className="flex items-center">
						<span className="text-sm">Details</span>
						{expanded ? <FiChevronUp className="mx-1" /> : <FiChevronDown className="mx-1" />}
					</div>
				</div>
			</div>

			{/* body */}
			{localPreset.isEnabled && expanded && (
				<div className="mt-8 space-y-6">
					{/* readonly provider details */}
					<div className="grid grid-cols-12 gap-2">
						<label className="col-span-3 text-sm">Origin:</label>
						<p className="col-span-9 text-sm">{localPreset.origin}</p>

						<label className="col-span-3 text-sm">Chat Path:</label>
						<p className="col-span-9 text-sm">{localPreset.chatCompletionPathPrefix}</p>

						<label className="col-span-3 text-sm">API Type:</label>
						<p className="col-span-9 text-sm">{localPreset.apiType}</p>
					</div>

					{/* default model dropdown */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="col-span-3 text-sm">Default Model</label>
						<div className="col-span-9">
							<Dropdown<ModelPresetID>
								dropdownItems={modelPresets}
								selectedKey={defaultModelPresetID}
								onChange={handleDefaultModelChange}
								filterDisabled={false}
								title="Select default model"
								getDisplayName={k => modelPresets[k].displayName || k}
							/>
						</div>
					</div>

					{/* model presets table */}
					<div className="overflow-x-auto border border-base-content/10 rounded-2xl">
						<table className="table table-zebra w-full">
							<thead>
								<tr className="text-sm font-semibold bg-base-300">
									<th>Preset Label</th>
									<th>Model Name</th>
									<th className="text-center">Enabled</th>
									<th className="text-center">Reasoning</th>
									<th className="text-right pr-8">Actions</th>
								</tr>
							</thead>
							<tbody>
								{Object.entries(modelPresets).map(([id, m]) => {
									const canModifyModel = !m.isBuiltIn;
									return (
										<tr key={id} className="hover:bg-base-300">
											<td>{m.displayName || id}</td>
											<td>{m.name}</td>
											<td className="text-center">
												<input
													type="checkbox"
													className="toggle toggle-accent rounded-full"
													checked={m.isEnabled}
													onChange={() => toggleModelEnable(id)}
												/>
											</td>
											<td className="text-center">
												{'reasoning' in m && m.reasoning ? (
													<FiCheck className="mx-auto" />
												) : (
													<FiX className="mx-auto" />
												)}
											</td>
											<td className="text-right">
												{canModifyModel ? (
													<>
														{/* edit */}
														<button
															className="btn btn-sm btn-ghost rounded-2xl"
															onClick={() => {
																openEditModel(id);
															}}
															title="Edit Preset"
														>
															<FiEdit size={16} />
														</button>
														{/* delete */}
														<button
															className="btn btn-sm btn-ghost rounded-2xl"
															onClick={() => {
																requestDeleteModel(id);
															}}
															title="Delete Preset"
														>
															<FiTrash2 size={16} />
														</button>
													</>
												) : (
													<span className="text-xs opacity-50">built-in</span>
												)}
											</td>
										</tr>
									);
								})}
							</tbody>
						</table>
					</div>

					{/* bottom bar */}
					<div className="flex justify-between items-center mt-4">
						<div className="flex gap-2">
							{/* edit provider */}
							<button
								className={`btn btn-ghost rounded-2xl flex items-center ${
									providerIsBuiltIn ? 'btn-disabled opacity-50 cursor-not-allowed' : ''
								}`}
								onClick={() => {
									if (!providerIsBuiltIn) onRequestEdit(provider);
									else {
										setDeniedMsg('Built-in providers cannot be edited.');
										setShowDenied(true);
									}
								}}
								title="Edit provider details"
								disabled={providerIsBuiltIn}
							>
								<FiEdit /> <span className="ml-1">Edit</span>
							</button>

							{/* delete provider */}
							<button
								className={`btn btn-ghost rounded-2xl flex items-center ${
									providerIsBuiltIn ? 'btn-disabled opacity-50 cursor-not-allowed' : ''
								}`}
								onClick={requestDeleteProvider}
								title="Delete provider"
								disabled={providerIsBuiltIn}
							>
								<FiTrash2 /> <span className="ml-1">Delete</span>
							</button>
						</div>

						{/* add preset */}
						<button
							className={`btn btn-ghost rounded-2xl flex items-center ${
								providerIsBuiltIn ? 'btn-disabled opacity-50 cursor-not-allowed' : ''
							}`}
							onClick={openAddModel}
							disabled={providerIsBuiltIn}
						>
							<FiPlus /> <span className="ml-1">Add Preset</span>
						</button>
					</div>
				</div>
			)}

			{/* dialogs / alerts */}
			{showDelProviderModal && (
				<DeleteConfirmationModal
					isOpen={showDelProviderModal}
					onClose={() => {
						setShowDelProviderModal(false);
					}}
					onConfirm={confirmDeleteProvider}
					title="Delete Provider"
					message={`Delete provider "${provider}"? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>
			)}

			{showDelModelModal && (
				<DeleteConfirmationModal
					isOpen={showDelModelModal}
					onClose={() => {
						setShowDelModelModal(false);
					}}
					onConfirm={confirmDeleteModel}
					title="Delete Preset"
					message={`Delete preset "${selectedModelID}"? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>
			)}

			{showModifyModal && (
				<ModifyModelModal
					isOpen={showModifyModal}
					onClose={() => {
						setShowModifyModal(false);
					}}
					onSubmit={handleModifyModelSubmit}
					providerName={provider}
					initialModelID={selectedModelID ?? undefined}
					initialData={selectedModelID ? modelPresets[selectedModelID] : undefined}
					existingModels={modelPresets}
				/>
			)}

			{showDenied && (
				<ActionDeniedAlert
					isOpen={showDenied}
					onClose={() => {
						setShowDenied(false);
						setDeniedMsg('');
					}}
					message={deniedMsg}
				/>
			)}
		</div>
	);
};

export default ProviderPresetCard;
