import { useEffect, useMemo, useState } from 'react';

import {
	FiCheck,
	FiCheckCircle,
	FiChevronDown,
	FiChevronUp,
	FiEdit2,
	FiKey,
	FiPlus,
	FiTrash2,
	FiX,
	FiXCircle,
} from 'react-icons/fi';

import {
	type ModelPreset,
	type ModelPresetID,
	type ProviderName,
	type ProviderPreset,
	SDK_DISPLAY_NAME,
} from '@/spec/modelpreset';
import type { AuthKeyMeta } from '@/spec/setting';
import { AuthKeyTypeProvider } from '@/spec/setting';

import { modelPresetStoreAPI, settingstoreAPI } from '@/apis/baseapi';

import { ActionDeniedAlertModal } from '@/components/action_denied_modal';
import { DeleteConfirmationModal } from '@/components/delete_confirmation_modal';
import { Dropdown } from '@/components/dropdown';

import { AddEditModelPresetModal } from '@/modelpresets/modelpreset_add_edit_modal';
import { AddEditAuthKeyModal } from '@/settings/authkey_add_edit_modal';

interface ProviderPresetCardProps {
	provider: ProviderName;
	preset: ProviderPreset;
	defaultProvider: ProviderName;
	authKeySet: boolean;
	enabledProviders: ProviderName[];
	allProviderPresets: Record<ProviderName, ProviderPreset>;

	onProviderPresetChange: (provider: ProviderName, newPreset: ProviderPreset) => void;
	onProviderDelete: (provider: ProviderName) => Promise<void>;
	onRequestEdit: (provider: ProviderName) => void; // open edit-modal
}

export function ProviderPresetCard({
	provider,
	preset,
	defaultProvider,
	authKeySet,
	enabledProviders,
	allProviderPresets,
	onProviderPresetChange,
	onProviderDelete,
	onRequestEdit,
}: ProviderPresetCardProps) {
	/* ───────── local ui-state ───────── */
	const [expanded, setExpanded] = useState(false);

	/* keep key status fresh */
	const [keySet, setKeySet] = useState(authKeySet);
	useEffect(() => {
		setKeySet(authKeySet);
	}, [authKeySet]);

	const [selectedID, setSelectedID] = useState<ModelPresetID | null>(null);
	const [showModModal, setShowModModal] = useState(false);
	const [showDelProv, setShowDelProv] = useState(false);
	const [showDelModel, setShowDelModel] = useState(false);
	const [showDenied, setShowDenied] = useState(false);
	const [deniedMsg, setDeniedMsg] = useState('');

	/* api-key modal */
	const [showKeyModal, setShowKeyModal] = useState(false);
	const [authKeys, setAuthKeys] = useState<AuthKeyMeta[]>([]);
	const [keyModalInitial, setKeyModalInitial] = useState<AuthKeyMeta | null>(null);

	/* ───────── derived helpers ───────── */
	const isLastEnabled = preset.isEnabled && enabledProviders.length === 1;
	const providerIsBuiltIn = preset.isBuiltIn;

	const modelPresets = preset.modelPresets;
	const defaultModelPresetID = preset.defaultModelPresetID;
	const modelEntries = Object.entries(modelPresets);
	const hasModels = modelEntries.length > 0;
	const canDeleteProvider = !providerIsBuiltIn && !hasModels;

	const allModelPresets = useMemo(() => {
		const o: Record<ProviderName, Record<ModelPresetID, ModelPreset>> = {};
		for (const [prov, pp] of Object.entries(allProviderPresets)) {
			o[prov] = pp.modelPresets;
		}
		return o;
	}, [allProviderPresets]);

	/* ───────── provider enable / disable ───────── */
	const toggleProviderEnable = async () => {
		if (!preset.isEnabled) {
			await modelPresetStoreAPI.patchProviderPreset(provider, true);
			onProviderPresetChange(provider, { ...preset, isEnabled: true });
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
			onProviderPresetChange(provider, { ...preset, isEnabled: false });
		} catch {
			setDeniedMsg('Failed toggling provider.');
			setShowDenied(true);
		}
	};

	/* ───────── expand / collapse ───────── */
	const toggleExpand = () => {
		if (preset.isEnabled) setExpanded(p => !p);
	};

	/* ───────── delete provider ───────── */
	const requestDeleteProvider = () => {
		if (providerIsBuiltIn) {
			setDeniedMsg('Built-in providers cannot be deleted.');
			setShowDenied(true);
			return;
		}
		if (hasModels) {
			setDeniedMsg('Only empty providers can be deleted. Remove all model presets first.');
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
			onProviderPresetChange(provider, { ...preset, defaultModelPresetID: id });
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
			onProviderPresetChange(provider, {
				...preset,
				modelPresets: { ...modelPresets, [id]: { ...m, isEnabled: !m.isEnabled } },
			});
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
			let newDefault = defaultModelPresetID;
			if (!defaultModelPresetID) {
				await modelPresetStoreAPI.patchProviderPreset(provider, undefined, id);
				newDefault = id;
			}

			onProviderPresetChange(provider, {
				...preset,
				modelPresets: { ...modelPresets, [id]: data },
				defaultModelPresetID: newDefault,
			});

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
			const { [selectedID]: _removed, ...rest } = modelPresets;

			/* ── if the deleted one was the default, pick a new default (or undefined) ── */
			let newDefault = defaultModelPresetID;
			if (selectedID === defaultModelPresetID) {
				newDefault = Object.keys(rest)[0] ?? '';
				if (newDefault !== '') {
					await modelPresetStoreAPI.patchProviderPreset(provider, undefined, newDefault);
				}
			}

			onProviderPresetChange(provider, {
				...preset,
				modelPresets: rest,
				defaultModelPresetID: newDefault,
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
		<div className="bg-base-100 mb-8 rounded-2xl px-4 py-2 shadow-lg">
			{/* ─── header ─── */}
			<div className="grid grid-cols-12 items-center gap-2 py-2">
				<div className="col-span-3">
					<h3 className="text-sm font-semibold capitalize">{preset.displayName || provider}</h3>
				</div>

				{/* enable toggle */}
				<div className="col-span-3 flex items-center gap-2">
					<label className="text-sm">Enable</label>
					<input
						type="checkbox"
						className="toggle toggle-accent"
						checked={preset.isEnabled}
						onChange={toggleProviderEnable}
					/>
				</div>

				{/* key status & expand */}
				<div className="col-span-6 flex cursor-pointer items-end justify-end gap-4" onClick={toggleExpand}>
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
			{preset.isEnabled && expanded && (
				<div className="mt-4 space-y-6">
					{/* provider-details table */}
					<div className="border-base-content/10 mb-4 overflow-x-auto rounded-2xl border">
						<table className="table w-full">
							<tbody>
								{/* actions row */}
								<tr>
									<td colSpan={2} className="py-0.5">
										<div className="flex items-center justify-between">
											{/* delete */}
											<span
												className="label-text-alt tooltip tooltip-right"
												data-tip={
													providerIsBuiltIn
														? 'Built-in providers cannot be deleted.'
														: hasModels
															? 'Only empty providers can be deleted.'
															: 'Delete Empty Provider'
												}
											>
												<button
													className={`btn btn-ghost flex items-center rounded-2xl ${
														!canDeleteProvider ? 'btn-disabled cursor-not-allowed opacity-50' : ''
													}`}
													onClick={canDeleteProvider ? requestDeleteProvider : undefined}
													title={'Delete Provider'}
													disabled={!canDeleteProvider}
												>
													<FiTrash2 />
													<span className="ml-1 hidden md:inline">Delete Provider</span>
												</button>
											</span>
											{/* api-key + edit */}
											<div className="flex gap-2">
												<button
													className="btn btn-ghost flex items-center rounded-2xl"
													onClick={openSetApiKey}
													title={keySet ? 'Update API Key' : 'Set API Key'}
												>
													<FiKey />
													<span className="ml-1 hidden md:inline">{keySet ? 'Update API Key' : 'Set API Key'}</span>
												</button>

												<button
													className={`btn btn-ghost flex items-center rounded-2xl ${
														providerIsBuiltIn ? 'btn-disabled cursor-not-allowed opacity-50' : ''
													}`}
													onClick={handleEditProvider}
													title="Edit Provider"
													disabled={providerIsBuiltIn}
												>
													<FiEdit2 />
													<span className="ml-1 hidden md:inline">Edit Provider</span>
												</button>
											</div>
										</div>
									</td>
								</tr>
								<tr className="hover:bg-base-300">
									<td className="w-1/3 text-sm">ID</td>
									<td className="text-sm">{preset.name}</td>
								</tr>
								<tr className="hover:bg-base-300">
									<td className="w-1/3 text-sm">SDK Type</td>
									<td className="text-sm">{SDK_DISPLAY_NAME[preset.sdkType]}</td>
								</tr>
								<tr className="hover:bg-base-300">
									<td className="w-1/3 text-sm">Origin</td>
									<td className="text-sm">{preset.origin}</td>
								</tr>
								<tr className="hover:bg-base-300">
									<td className="w-1/3 text-sm">Chat Path</td>
									<td className="text-sm">{preset.chatCompletionPathPrefix}</td>
								</tr>
							</tbody>
						</table>
					</div>

					{/* ─── model presets section ─── */}
					<div className="border-base-content/10 mb-2 overflow-x-auto rounded-2xl border">
						{/* default model selector row (always present) */}
						<div className="grid grid-cols-12 items-center gap-4 px-4 py-2">
							<span className="col-span-3 text-sm font-semibold">Default Model</span>

							<div className="col-span-6 ml-8">
								{hasModels ? (
									<Dropdown<ModelPresetID>
										dropdownItems={modelPresets}
										selectedKey={defaultModelPresetID}
										onChange={handleDefaultModelChange}
										filterDisabled={false}
										title="Select default model"
										getDisplayName={k => modelPresets[k].displayName || k}
									/>
								) : (
									<span className="text-sm italic">No model presets configured.</span>
								)}
							</div>

							{/* add model btn */}
							<div className="col-span-3 flex justify-end">
								<button
									className={`btn btn-ghost flex items-center rounded-2xl ${
										providerIsBuiltIn ? 'btn-disabled cursor-not-allowed opacity-50' : ''
									}`}
									onClick={openAddModel}
									disabled={providerIsBuiltIn}
									title="Add Model Preset"
								>
									<FiPlus size={16} />
									<span className="ml-1 hidden md:inline">Add Model Preset</span>
								</button>
							</div>
						</div>

						{/* table only if models exist */}
						{hasModels && (
							<table className="table-zebra table w-full">
								<thead>
									<tr className="bg-base-300 text-sm font-semibold">
										<th>Model Preset Label</th>
										<th>Model Name</th>
										<th className="text-center">Enabled</th>
										<th className="text-center">Reasoning</th>
										<th className="text-center">Actions</th>
									</tr>
								</thead>

								<tbody>
									{modelEntries.map(([id, m]) => {
										const canModify = !m.isBuiltIn;
										return (
											<tr key={id} className="hover:bg-base-300">
												<td>{m.displayName || id}</td>
												<td>{m.name}</td>
												<td className="text-center">
													<input
														type="checkbox"
														className="toggle toggle-accent"
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
												<td className="space-x-1 text-center">
													{canModify ? (
														<>
															<button
																className="btn btn-sm btn-ghost rounded-2xl"
																onClick={() => {
																	openEditModel(id);
																}}
																title="Edit Model Preset"
															>
																<FiEdit2 size={16} />
															</button>
															<button
																className="btn btn-sm btn-ghost rounded-2xl"
																onClick={() => {
																	requestDeleteModel(id);
																}}
																title="Delete Model Preset"
															>
																<FiTrash2 size={16} />
															</button>
														</>
													) : (
														<span className="text-xs opacity-50">Built-in</span>
													)}
												</td>
											</tr>
										);
									})}
								</tbody>
							</table>
						)}
					</div>
				</div>
			)}

			{/* ───────── dialogs & alerts ───────── */}
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
					allModelPresets={allModelPresets}
				/>
			)}

			{/* ───────── auth-key modal ───────── */}
			{showKeyModal && (
				<AddEditAuthKeyModal
					isOpen={showKeyModal}
					initial={keyModalInitial} // edit-mode when not null
					existing={authKeys}
					/* NEW – provide the (type,keyName) to pre-select when we are in pure “add” mode */
					prefill={!keyModalInitial ? { type: AuthKeyTypeProvider, keyName: provider } : undefined}
					onClose={() => {
						setShowKeyModal(false);
					}}
					onChanged={() => {
						setShowKeyModal(false);
						setKeySet(true);
					}}
				/>
			)}

			{showDenied && (
				<ActionDeniedAlertModal
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
}
