import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import { type ProviderName, type ProviderPreset } from '@/spec/modelpreset';
import { AuthKeyTypeProvider } from '@/spec/setting';

import { modelPresetStoreAPI, settingstoreAPI } from '@/apis/baseapi';
import { getAllProviderPresetsMap } from '@/apis/modelpreset_helper';

import ActionDeniedAlert from '@/components/action_denied';
import DownloadButton from '@/components/download_button';
import Dropdown from '@/components/dropdown';
import Loader from '@/components/loader';

import AddEditProviderPresetModal from '@/modelpresets/provider_add_edit';
import ProviderPresetCard from '@/modelpresets/provider_presets_card';

const ModelPresetPage: FC = () => {
	/* ─────────────── state ─────────────── */
	// default provider is *unknown* until it is fetched
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

	/* ─────────────── data fetch ─────────────── */
	useEffect(() => {
		(async () => {
			setLoading(true);
			setError(null);
			try {
				const settingsPromise = settingstoreAPI.getSettings();
				const defProvPromise = modelPresetStoreAPI.getDefaultProvider();
				const presetsPromise = getAllProviderPresetsMap(true);

				const [settings, defProv, allProviderPresets] = await Promise.all([
					settingsPromise,
					defProvPromise,
					presetsPromise,
				]);

				setDefaultProvider(defProv);
				setProviderPresets(allProviderPresets);
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

	/* ─────────────── derived ─────────────── */
	const enabledProviders = useMemo(
		() =>
			Object.values(providerPresets)
				.filter(p => p.isEnabled)
				.map(p => p.name),
		[providerPresets]
	);

	/* ─────────────── handlers ─────────────── */
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

	const handleProviderPresetChange = (name: ProviderName, newPreset: ProviderPreset) => {
		setProviderPresets(prev => ({ ...prev, [name]: newPreset }));
	};

	const handleProviderDelete = async (name: ProviderName) => {
		if (providerPresets[name].isBuiltIn) {
			setDeniedMsg('Built-in providers cannot be deleted.');
			setShowDenied(true);
			return;
		}

		try {
			if (name === defaultProvider) {
				setDeniedMsg('Cannot delete the current default provider. Pick another default first.');
				setShowDenied(true);
				return;
			}

			await modelPresetStoreAPI.deleteProviderPreset(name);
			await settingstoreAPI.deleteAuthKey(AuthKeyTypeProvider, name).catch(() => void 0); // ignore if not existing

			setProviderPresets(prev => {
				const { [name]: _, ...rest } = prev;
				return rest;
			});
			setProviderKeySet(prev => {
				const { [name]: _, ...rest } = prev;
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
		try {
			await modelPresetStoreAPI.putProviderPreset(name, payload);
			if (apiKey && apiKey.trim()) {
				await settingstoreAPI.setAuthKey(AuthKeyTypeProvider, name, apiKey.trim());
			}

			// optimistic update
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
			if (apiKey && apiKey.trim()) setProviderKeySet(prev => ({ ...prev, [name]: true }));
		} catch (err) {
			console.error(err);
			setDeniedMsg(isEdit ? 'Failed updating provider.' : 'Failed adding provider.');
			setShowDenied(true);
		}
	};

	/* ---------- helpers ---------- */
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

	// show full-page loader until the first batch of data arrives
	if (loading) {
		return <Loader text="Loading model presets..." />;
	}

	return (
		<div className="flex flex-col items-center w-full h-full">
			{/* header */}
			<div className="w-full flex justify-center fixed top-8 z-10">
				<div className="w-10/12 lg:w-2/3 flex items-center justify-between p-2">
					<h1 className="text-xl font-semibold flex-grow text-center">Model Presets</h1>

					<DownloadButton
						title="Download Model Presets"
						language="json"
						valueFetcher={fetchValue}
						size={24}
						fileprefix="modelpresets"
						className="btn btn-sm btn-ghost"
					/>
				</div>
			</div>

			{/* body */}
			<div
				className="flex flex-col items-center w-full grow mt-24 overflow-y-auto"
				style={{ maxHeight: 'calc(100vh - 128px)' }}
			>
				<div className="flex flex-col space-y-4 w-5/6 xl:w-2/3">
					{/* default-provider row */}
					<div className="bg-base-100 rounded-2xl shadow-lg px-4 py-2 mb-8">
						<div className="grid grid-cols-12 items-center gap-4">
							<label className="col-span-3 text-sm font-medium">Default Provider</label>

							<div className="col-span-6">
								{/* render dropdown only when we have both data pieces */}
								{defaultProvider && Object.keys(providerPresets).length > 0 && (
									<Dropdown<ProviderName>
										dropdownItems={providerPresets}
										selectedKey={defaultProvider}
										onChange={handleDefaultProviderChange}
										filterDisabled={false}
										title="Select default provider"
										getDisplayName={k => providerPresets[k].displayName}
									/>
								)}
							</div>

							<div className="col-span-3 flex justify-end">
								<button className="btn btn-ghost rounded-2xl flex items-center" onClick={openAddModal}>
									<FiPlus /> <span className="ml-1">Add Provider</span>
								</button>
							</div>
						</div>
					</div>

					{/* provider cards / errors */}
					{error && <p className="text-error text-center mt-8">{error}</p>}

					{!error &&
						Object.entries(providerPresets).map(([name, p]) => (
							<ProviderPresetCard
								key={name}
								provider={name}
								preset={p}
								defaultProvider={defaultProvider ?? ''}
								authKeySet={!!providerKeySet[name]}
								enabledProviders={enabledProviders}
								onProviderPresetChange={handleProviderPresetChange}
								onProviderDelete={handleProviderDelete}
								onRequestEdit={openEditModal}
							/>
						))}
				</div>
			</div>

			{/* add / edit modal */}
			{modalOpen && (
				<AddEditProviderPresetModal
					isOpen={modalOpen}
					mode={modalMode}
					onClose={() => {
						setModalOpen(false);
					}}
					onSubmit={(n, payload, key) => handleProviderModalSubmit(n, payload, key, modalMode === 'edit')}
					existingProviderNames={Object.keys(providerPresets)}
					initialPreset={modalMode === 'edit' && editProvider ? providerPresets[editProvider] : undefined}
					apiKeyAlreadySet={editProvider ? !!providerKeySet[editProvider] : false}
				/>
			)}

			{/* alerts */}
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
