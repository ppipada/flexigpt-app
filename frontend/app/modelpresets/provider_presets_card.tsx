import type { FC } from 'react';
import { useEffect, useState } from 'react';

import {
	FiCheck,
	FiCheckCircle,
	FiChevronDown,
	FiChevronUp,
	FiEdit,
	FiKey,
	FiPlus,
	FiTrash2,
	FiX,
	FiXCircle,
} from 'react-icons/fi';

import { type ModelPreset, type ModelPresetID, type ProviderName, type ProviderPreset } from '@/spec/modelpreset';
import type { AuthKeyMeta } from '@/spec/setting';
import { AuthKeyTypeProvider } from '@/spec/setting';

import { modelPresetStoreAPI, settingstoreAPI } from '@/apis/baseapi';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';
import Dropdown from '@/components/dropdown';

import AddEditModelPresetModal from '@/modelpresets/modelpreset_add_edit';
import AddEditAuthKeyModal from '@/settings/authkey_add_edit';

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
	/* ───────── local & derived ───────── */
	const [localPreset, setLocalPreset] = useState<ProviderPreset>(preset);
	useEffect(() => {
		setLocalPreset(preset);
	}, [preset]);

	/* api-key state (keeps up-to-date once user sets key) */
	const [keySet, setKeySet] = useState(authKeySet);
	useEffect(() => {
		setKeySet(authKeySet);
	}, [authKeySet]);

	const [expanded, setExpanded] = useState(false);

	const [selectedID, setSelectedID] = useState<ModelPresetID | null>(null);
	const [showModModal, setShowModModal] = useState(false);

	const [showDelProv, setShowDelProv] = useState(false);
	const [showDelModel, setShowDelModel] = useState(false);

	const [showDenied, setShowDenied] = useState(false);
	const [deniedMsg, setDeniedMsg] = useState('');

	/* auth-key modal */
	const [showKeyModal, setShowKeyModal] = useState(false);
	const [authKeys, setAuthKeys] = useState<AuthKeyMeta[]>([]);
	const [keyModalInitial, setKeyModalInitial] = useState<AuthKeyMeta | null>(null);

	/* helpers */
	const isLastEnabled = localPreset.isEnabled && enabledProviders.length === 1;
	const providerIsBuiltIn = localPreset.isBuiltIn;

	const modelPresets = localPreset.modelPresets;
	const defaultModelPresetID = localPreset.defaultModelPresetID;
	const modelEntries = Object.entries(modelPresets);
	const hasModels = modelEntries.length > 0;

	const updateLocal = (updater: (p: ProviderPreset) => ProviderPreset) => {
		setLocalPreset(prev => {
			const upd = updater(prev);
			onProviderPresetChange(provider, upd);
			return upd;
		});
	};

	/* ───────── provider enable / disable ───────── */
	const toggleProviderEnable = async () => {
		if (!localPreset.isEnabled) {
			await modelPresetStoreAPI.patchProviderPreset(provider, true);
			updateLocal(p => ({ ...p, isEnabled: true }));
			return;
		}

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
		} catch {
			setDeniedMsg('Failed toggling provider.');
			setShowDenied(true);
		}
	};

	/* ───────── expand / collapse ───────── */
	const toggleExpand = () => {
		if (localPreset.isEnabled) setExpanded(p => !p);
	};

	/* ───────── delete provider ───────── */
	const requestDeleteProvider = () => {
		if (providerIsBuiltIn) {
			setDeniedMsg('Built-in providers cannot be deleted.');
			setShowDenied(true);
			return;
		}
		setShowDelProv(true);
	};
	const confirmDeleteProvider = async () => {
		await onProviderDelete(provider);
		setShowDelProv(false);
	};

	/* ───────── default model change ───────── */
	const handleDefaultModelChange = async (id: ModelPresetID) => {
		try {
			await modelPresetStoreAPI.patchProviderPreset(provider, undefined, id);
			updateLocal(p => ({ ...p, defaultModelPresetID: id }));
		} catch {
			setDeniedMsg('Failed setting default model.');
			setShowDenied(true);
		}
	};

	/* ───────── enable / disable model ───────── */
	const toggleModelEnable = async (id: ModelPresetID) => {
		const m = modelPresets[id];
		if (id === defaultModelPresetID && m.isEnabled) {
			setDeniedMsg('Cannot disable the default model preset. Choose another default first.');
			setShowDenied(true);
			return;
		}
		try {
			await modelPresetStoreAPI.patchModelPreset(provider, id, !m.isEnabled);
			updateLocal(p => ({
				...p,
				modelPresets: { ...p.modelPresets, [id]: { ...m, isEnabled: !m.isEnabled } },
			}));
		} catch {
			setDeniedMsg('Failed toggling model.');
			setShowDenied(true);
		}
	};

	/* ───────── add / edit model ───────── */
	const openAddModel = () => {
		if (providerIsBuiltIn) {
			setDeniedMsg('Cannot add model presets to a built-in provider.');
			setShowDenied(true);
			return;
		}
		setSelectedID(null);
		setShowModModal(true);
	};
	const openEditModel = (id: ModelPresetID) => {
		if (modelPresets[id].isBuiltIn) {
			setDeniedMsg('Built-in model presets cannot be edited.');
			setShowDenied(true);
			return;
		}
		setSelectedID(id);
		setShowModModal(true);
	};

	const handleModifyModelSubmit = async (id: ModelPresetID, data: ModelPreset) => {
		try {
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			const { id: _id, isBuiltIn, ...payload } = data;
			await modelPresetStoreAPI.putModelPreset(provider, id, payload);
			updateLocal(p => ({
				...p,
				modelPresets: { ...p.modelPresets, [id]: data },
			}));
			setShowModModal(false);
		} catch {
			setDeniedMsg('Failed saving model preset.');
			setShowDenied(true);
		}
	};

	/* ───────── delete model ───────── */
	const requestDeleteModel = (id: ModelPresetID) => {
		if (modelPresets[id].isBuiltIn) {
			setDeniedMsg('Built-in model presets cannot be deleted.');
			setShowDenied(true);
			return;
		}
		setSelectedID(id);
		setShowDelModel(true);
	};
	const confirmDeleteModel = async () => {
		if (!selectedID) return;
		try {
			await modelPresetStoreAPI.deleteModelPreset(provider, selectedID);
			updateLocal(p => {
				const { [selectedID]: _, ...rest } = p.modelPresets;
				return { ...p, modelPresets: rest };
			});
			setShowDelModel(false);
		} catch {
			setDeniedMsg('Failed deleting model preset.');
			setShowDenied(true);
		}
	};

	/* ───────── api-key modal helpers ───────── */
	const openSetApiKey = async () => {
		try {
			const settings = await settingstoreAPI.getSettings();
			setAuthKeys(settings.authKeys);

			const meta = settings.authKeys.find(k => k.type === AuthKeyTypeProvider && k.keyName === provider) ?? null;

			setKeyModalInitial(meta);
			setShowKeyModal(true);
		} catch (err) {
			console.error(err);
			setDeniedMsg('Failed loading auth keys.');
			setShowDenied(true);
		}
	};

	const handleEditProvider = () => {
		if (!providerIsBuiltIn) onRequestEdit(provider);
		else {
			setDeniedMsg('Built-in providers cannot be edited.');
			setShowDenied(true);
		}
	};

	/* ─────────────────────────── render ─────────────────────────── */
	return (
		<div className="bg-base-100 rounded-2xl shadow-lg px-4 py-2 mb-8">
			{/* ─── header ─── */}
			<div className="grid grid-cols-12 gap-2 items-center py-2">
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

				{/* key status & expand */}
				<div className="col-span-6 flex items-end justify-end gap-4 cursor-pointer" onClick={toggleExpand}>
					<div className="flex items-center">
						<span className="text-sm">API-Key</span>
						{keySet ? <FiCheckCircle className="text-success mx-1" /> : <FiXCircle className="text-error mx-1" />}
					</div>

					<div className="flex items-center">
						<span className="text-sm">Details</span>
						{expanded ? <FiChevronUp className="mx-1" /> : <FiChevronDown className="mx-1" />}
					</div>
				</div>
			</div>

			{/* ─── body ─── */}
			{localPreset.isEnabled && expanded && (
				<div className="mt-4 space-y-6">
					{/* provider-details table (no header) */}
					<div className="overflow-x-auto border border-base-content/10 rounded-2xl mb-4">
						<table className="table w-full">
							<tbody>
								{/* actions row (delete / api-key / edit) */}
								<tr>
									<td colSpan={2} className="py-0.5">
										<div className="flex justify-between items-center">
											{/* delete on left */}
											<button
												className={`btn btn-ghost rounded-2xl flex items-center ${
													providerIsBuiltIn ? 'btn-disabled opacity-50 cursor-not-allowed' : ''
												}`}
												onClick={requestDeleteProvider}
												title="Delete Provider"
												disabled={providerIsBuiltIn}
											>
												<FiTrash2 /> <span className="ml-1 hidden md:inline">Delete Provider</span>
											</button>

											{/* api-key + edit on right */}
											<div className="flex gap-2">
												<button
													className="btn btn-ghost rounded-2xl flex items-center"
													onClick={openSetApiKey}
													title={keySet ? 'Update API Key' : 'Set API Key'}
												>
													<FiKey />{' '}
													<span className="ml-1 hidden md:inline">{keySet ? 'Update Key' : 'Set API Key'}</span>
												</button>

												<button
													className={`btn btn-ghost rounded-2xl flex items-center ${
														providerIsBuiltIn ? 'btn-disabled opacity-50 cursor-not-allowed' : ''
													}`}
													onClick={handleEditProvider}
													title="Edit Provider"
													disabled={providerIsBuiltIn}
												>
													<FiEdit /> <span className="ml-1 hidden md:inline">Edit Provider</span>
												</button>
											</div>
										</div>
									</td>
								</tr>
								<tr className="hover:bg-base-300">
									<td className="w-1/3 text-sm">Origin</td>
									<td className="text-sm">{localPreset.origin}</td>
								</tr>
								<tr className="hover:bg-base-300">
									<td className="w-1/3 text-sm">Chat Path</td>
									<td className="text-sm">{localPreset.chatCompletionPathPrefix}</td>
								</tr>
							</tbody>
						</table>
					</div>

					{/* model preset table */}
					<div className="overflow-x-auto border border-base-content/10 rounded-2xl mb-2">
						{/* default-model selector */}
						{hasModels && (
							<div className="grid grid-cols-12 items-center gap-4 px-4 py-2">
								<span className="col-span-3 text-sm font-semibold">Default Model</span>
								<div className="col-span-9 flex-1">
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
						)}
						<table className="table table-zebra w-full">
							<thead>
								<tr className="text-sm font-semibold bg-base-300">
									<th>Model Preset Label</th>
									<th>Model Name</th>
									<th className="text-center">Enabled</th>
									<th className="text-center">Reasoning</th>
									<th className="text-center">Actions</th>
								</tr>
							</thead>

							<tbody>
								{/* model rows */}
								{hasModels ? (
									modelEntries.map(([id, m]) => {
										const canModify = !m.isBuiltIn;
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
												<td className="text-center space-x-1">
													{canModify ? (
														<>
															<button
																className="btn btn-sm btn-ghost rounded-2xl"
																onClick={() => {
																	openEditModel(id);
																}}
																title="Edit Model Preset"
															>
																<FiEdit size={16} />
																<span className="ml-1 hidden md:inline">Edit Preset</span>
															</button>
															<button
																className="btn btn-sm btn-ghost rounded-2xl"
																onClick={() => {
																	requestDeleteModel(id);
																}}
																title="Delete Model Preset"
															>
																<FiTrash2 size={16} />
																<span className="ml-1 hidden md:inline">Delete Preset</span>
															</button>
														</>
													) : (
														<span className="text-xs opacity-50">built-in</span>
													)}
												</td>
											</tr>
										);
									})
								) : (
									<tr>
										<td colSpan={5} className="py-4 text-center text-sm italic">
											No model presets configured.
										</td>
									</tr>
								)}
							</tbody>
						</table>
					</div>

					{/* add model preset button */}
					<div className="flex justify-end items-center mt-2">
						<button
							className={`btn btn-ghost rounded-2xl flex items-center ${
								providerIsBuiltIn ? 'btn-disabled opacity-50 cursor-not-allowed' : ''
							}`}
							onClick={openAddModel}
							disabled={providerIsBuiltIn}
						>
							<FiPlus /> <span className="ml-1">Add Model Preset</span>
						</button>
					</div>
				</div>
			)}

			{/* ───────── dialogs & alerts ───────── */}
			{/* provider delete */}
			{showDelProv && (
				<DeleteConfirmationModal
					isOpen={showDelProv}
					onClose={() => {
						setShowDelProv(false);
					}}
					onConfirm={confirmDeleteProvider}
					title="Delete Provider"
					message={`Delete provider “${provider}”? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>
			)}

			{/* model delete */}
			{showDelModel && (
				<DeleteConfirmationModal
					isOpen={showDelModel}
					onClose={() => {
						setShowDelModel(false);
					}}
					onConfirm={confirmDeleteModel}
					title="Delete Model Preset"
					message={`Delete model preset “${selectedID}”? This action cannot be undone.`}
					confirmButtonText="Delete"
				/>
			)}

			{/* model add / edit */}
			{showModModal && (
				<AddEditModelPresetModal
					isOpen={showModModal}
					onClose={() => {
						setShowModModal(false);
					}}
					onSubmit={handleModifyModelSubmit}
					providerName={provider}
					initialModelID={selectedID ?? undefined}
					initialData={selectedID ? modelPresets[selectedID] : undefined}
					existingModels={modelPresets}
				/>
			)}

			{/* api-key modal */}
			{showKeyModal && (
				<AddEditAuthKeyModal
					isOpen={showKeyModal}
					initial={keyModalInitial}
					existing={authKeys}
					onClose={() => {
						setShowKeyModal(false);
					}}
					onChanged={() => {
						setShowKeyModal(false);
						setKeySet(true);
					}}
				/>
			)}

			{/* denied alert */}
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
