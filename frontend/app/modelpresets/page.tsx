import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import { type ProviderName, type ProviderPreset } from '@/spec/modelpreset';
import { AuthKeyTypeProvider } from '@/spec/setting';

import { modelPresetStoreAPI, settingstoreAPI } from '@/apis/baseapi';
import { getAllProviderPresetsMap } from '@/apis/modelpreset_helper';

import ActionDeniedAlert from '@/components/action_denied';
import DownloadButton from '@/components/download_button';
import Dropdown, { type DropdownItem } from '@/components/dropdown';
import Loader from '@/components/loader';

import AddEditProviderPresetModal from '@/modelpresets/provider_add_edit';
import ProviderPresetCard from '@/modelpresets/provider_presets_card';

// put it somewhere near the top of the file
const sortByDisplayName = ([, a]: [string, ProviderPreset], [, b]: [string, ProviderPreset]) =>
	a.displayName.localeCompare(b.displayName);
/* -------------------------------------------------------------------------- */

const ModelPresetPage: FC = () => {
	/* ----------------------------- local states ---------------------------- */
	const [defaultProvider, setDefaultProvider] = useState<ProviderName | undefined>(undefined);
	const [providerPresets, setProviderPresets] = useState<Record<ProviderName, ProviderPreset>>({});
	const [providerKeySet, setProviderKeySet] = useState<Record<ProviderName, boolean>>({});

	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const [showDenied, setShowDenied] = useState(false);
	const [deniedMsg, setDeniedMsg] = useState('');

	const [modalOpen, setModalOpen] = useState(false);
	const [modalMode, setModalMode] = useState<'add' | 'edit'>('add');
	const [editProvider, setEditProvider] = useState<ProviderName | null>(null);

	/* ---------------------------------------------------------------------- */
	/*                             initial loading                            */
	/* ---------------------------------------------------------------------- */
	useEffect(() => {
		(async () => {
			try {
				setLoading(true);
				const [settings, defProv, presets] = await Promise.all([
					settingstoreAPI.getSettings(),
					modelPresetStoreAPI.getDefaultProvider(),
					getAllProviderPresetsMap(true),
				]);

				setDefaultProvider(defProv);
				setProviderPresets(presets);
				setProviderKeySet(
					Object.fromEntries(
						settings.authKeys.filter(k => k.type === AuthKeyTypeProvider).map(k => [k.keyName, k.nonEmpty])
					)
				);
			} catch (err) {
				console.error('init model presets error', err);
				setError('Failed to load provider presets. Try again.');
			} finally {
				setLoading(false);
			}
		})();
	}, []);

	/* ---------------------------------------------------------------------- */
	/*                            derived helpers                             */
	/* ---------------------------------------------------------------------- */
	const enabledProviderNames = useMemo(
		() =>
			Object.values(providerPresets)
				.filter(p => p.isEnabled)
				.map(p => p.name),
		[providerPresets]
	);

	/* ------------------------------------------------------------------ */
	/* enabledProviderPresets – contains ONLY the enabled providers       */
	/* ------------------------------------------------------------------ */
	const enabledProviderPresets = useMemo<Partial<Record<ProviderName, ProviderPreset>>>(() => {
		const obj: Partial<Record<ProviderName, ProviderPreset>> = {};
		for (const [name, preset] of Object.entries(providerPresets)) {
			if (preset.isEnabled) obj[name] = preset;
		}
		return obj;
	}, [providerPresets]);

	/* ------------------------------------------------------------------ */
	/* safeDefaultKey – use it only if it exists inside the partial map   */
	/* ------------------------------------------------------------------ */
	const safeDefaultKey: ProviderName | undefined =
		defaultProvider && defaultProvider in enabledProviderPresets ? defaultProvider : undefined;
	/* ---------------------------------------------------------------------- */
	/*                          automatic default fix                         */
	/* ---------------------------------------------------------------------- */
	/* If no default is set OR the current default somehow became disabled,
	   automatically pick the first enabled provider (if any).                */
	useEffect(() => {
		if (enabledProviderNames.length === 0) return; // nothing enabled ⇒ nothing to do

		if (!defaultProvider || !providerPresets[defaultProvider].isEnabled) {
			const first = enabledProviderNames[0];
			(async () => {
				try {
					await modelPresetStoreAPI.patchDefaultProvider(first);
					setDefaultProvider(first);
				} catch (err) {
					console.error(err);
					setDeniedMsg('Failed setting default provider.');
					setShowDenied(true);
				}
			})();
		}
	}, [enabledProviderNames, defaultProvider, providerPresets]);

	/* ---------------------------------------------------------------------- */
	/*                              event handlers                            */
	/* ---------------------------------------------------------------------- */
	const handleDefaultProviderChange = async (prov: ProviderName) => {
		try {
			setDefaultProvider(prov);
			await modelPresetStoreAPI.patchDefaultProvider(prov);
		} catch (err) {
			console.error(err);
			setDeniedMsg('Failed changing default provider.');
			setShowDenied(true);
		}
	};

	/* Prevent disabling the current default provider */
	const handleProviderPresetChange = (name: ProviderName, newPreset: ProviderPreset) => {
		if (name === defaultProvider && !newPreset.isEnabled) {
			setDeniedMsg('Cannot disable the current default provider. Pick another default first.');
			setShowDenied(true);
			return;
		}

		setProviderPresets(prev => ({ ...prev, [name]: newPreset }));

		/* If there is no default yet and this preset has just been enabled,
		   make it the default automatically. */
		if (!defaultProvider && newPreset.isEnabled) {
			handleDefaultProviderChange(name);
		}
	};

	const handleProviderDelete = async (name: ProviderName) => {
		if (providerPresets[name].isBuiltIn) {
			setDeniedMsg('Built-in providers cannot be deleted.');
			setShowDenied(true);
			return;
		}

		if (name === defaultProvider) {
			setDeniedMsg('Cannot delete the current default provider. Pick another default first.');
			setShowDenied(true);
			return;
		}

		try {
			await modelPresetStoreAPI.deleteProviderPreset(name);
			await settingstoreAPI.deleteAuthKey(AuthKeyTypeProvider, name).catch(() => void 0); // ignore missing key

			setProviderPresets(prev => {
				const { [name]: _removed, ...rest } = prev;
				return rest;
			});
			setProviderKeySet(prev => {
				const { [name]: _removed, ...rest } = prev;
				return rest;
			});
		} catch (err) {
			console.error(err);
			setDeniedMsg('Failed deleting provider.');
			setShowDenied(true);
		}
	};

	const handleProviderModalSubmit = async (
		name: ProviderName,
		payload: Omit<ProviderPreset, 'isBuiltIn' | 'defaultModelPresetID' | 'modelPresets'>,
		apiKey: string | null,
		isEdit: boolean
	) => {
		/* Cannot disable current default */
		if (isEdit && name === defaultProvider && !payload.isEnabled) {
			setDeniedMsg('Cannot disable the current default provider. Pick another default first.');
			setShowDenied(true);
			return;
		}

		try {
			await modelPresetStoreAPI.putProviderPreset(name, payload);
			if (apiKey && apiKey.trim()) {
				await settingstoreAPI.setAuthKey(AuthKeyTypeProvider, name, apiKey.trim());
			}

			setProviderPresets(prev => ({
				...prev,
				[name]: {
					...(prev[name] ?? {}),
					...payload,
					defaultModelPresetID: '',
					isBuiltIn: false,
					modelPresets: {},
				} as ProviderPreset,
			}));

			if (apiKey && apiKey.trim()) {
				setProviderKeySet(prev => ({ ...prev, [name]: true }));
			}

			/* Auto-set default if none exists & this one is enabled */
			if (!defaultProvider && payload.isEnabled) {
				await modelPresetStoreAPI.patchDefaultProvider(name);
				setDefaultProvider(name);
			}
		} catch (err) {
			console.error(err);
			setDeniedMsg(isEdit ? 'Failed updating provider.' : 'Failed adding provider.');
			setShowDenied(true);
		}
	};

	/* ---------------------------- misc helpers ---------------------------- */
	const fetchValue = async () => {
		try {
			const [{ providers }, defProv] = await Promise.all([
				modelPresetStoreAPI.listProviderPresets(undefined, true),
				modelPresetStoreAPI.getDefaultProvider(),
			]);
			return JSON.stringify({ defaultProvider: defProv, providers }, null, 2);
		} catch (err) {
			console.log('fetch preset error', err);
			setDeniedMsg('Failed fetching presets.');
			setShowDenied(true);
			return '';
		}
	};

	const openAddModal = () => {
		setModalMode('add');
		setEditProvider(null);
		setModalOpen(true);
	};

	const openEditModal = (name: ProviderName) => {
		if (providerPresets[name].isBuiltIn) {
			setDeniedMsg('Built-in providers cannot be edited.');
			setShowDenied(true);
			return;
		}
		setModalMode('edit');
		setEditProvider(name);
		setModalOpen(true);
	};

	/* ---------------------------------------------------------------------- */
	/*                                  UI                                    */
	/* ---------------------------------------------------------------------- */

	if (loading) return <Loader text="Loading model presets…" />;

	return (
		<div className="flex flex-col items-center w-full h-full">
			{/* ------------------------------ header ----------------------------- */}
			<div className="flex w-10/12 lg:w-2/3 items-center fixed mt-8 p-2">
				<h1 className="flex grow items-center justify-center text-xl font-semibold">Model Presets</h1>
				<DownloadButton
					title="Download Model Presets"
					language="json"
					valueFetcher={fetchValue}
					size={20}
					fileprefix="modelpresets"
					className="btn btn-sm btn-ghost"
				/>
			</div>

			{/* ------------------------------ body ------------------------------ */}
			<div
				className="flex flex-col items-center w-full grow mt-24 overflow-y-auto"
				style={{ maxHeight: 'calc(100vh - 128px)' }}
			>
				<div className="flex flex-col space-y-4 w-5/6 xl:w-2/3">
					{/* --------------- default-provider selection row ---------------- */}
					<div className="bg-base-100 rounded-2xl shadow-lg px-4 py-2 mb-8">
						<div className="grid grid-cols-12 items-center gap-4">
							<label className="col-span-3 text-sm font-medium">Default Provider</label>

							<div className="col-span-6">
								{enabledProviderNames.length > 0 && safeDefaultKey ? (
									<Dropdown<ProviderName>
										dropdownItems={enabledProviderPresets as Record<string, DropdownItem>}
										/* use the safe key */
										selectedKey={safeDefaultKey}
										onChange={handleDefaultProviderChange}
										filterDisabled={false}
										title="Select default provider"
										/* fallback so it never crashes */
										getDisplayName={k => enabledProviderPresets[k]?.displayName ?? ''}
									/>
								) : (
									<span className="text-sm text-error">Enable at least one provider first.</span>
								)}
							</div>

							<div className="col-span-3 flex justify-end">
								<button className="btn btn-ghost rounded-2xl flex items-center" onClick={openAddModal}>
									<FiPlus /> <span className="ml-1">Add Provider</span>
								</button>
							</div>
						</div>
					</div>

					{/* -------------------- provider cards / errors ------------------ */}
					{error && <p className="text-error text-center mt-8">{error}</p>}

					{!error &&
						Object.entries(providerPresets)
							.sort(sortByDisplayName)
							.map(([name, p]) => (
								<ProviderPresetCard
									key={name}
									provider={name}
									preset={p}
									defaultProvider={defaultProvider ?? ''}
									authKeySet={providerKeySet[name]}
									enabledProviders={enabledProviderNames}
									onProviderPresetChange={handleProviderPresetChange}
									onProviderDelete={handleProviderDelete}
									onRequestEdit={openEditModal}
									allProviderPresets={providerPresets}
								/>
							))}
				</div>
			</div>

			{/* ----------------------- Add / Edit modal ------------------------ */}
			{modalOpen && (
				<AddEditProviderPresetModal
					isOpen={modalOpen}
					mode={modalMode}
					onClose={() => {
						setModalOpen(false);
					}}
					onSubmit={(n, payload, key) => handleProviderModalSubmit(n, payload, key, modalMode === 'edit')}
					existingProviderNames={Object.keys(providerPresets)}
					allProviderPresets={providerPresets}
					initialPreset={modalMode === 'edit' && editProvider ? providerPresets[editProvider] : undefined}
					apiKeyAlreadySet={editProvider ? providerKeySet[editProvider] : false}
				/>
			)}

			{/* ---------------------------- alerts ----------------------------- */}
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

export default ModelPresetPage;
