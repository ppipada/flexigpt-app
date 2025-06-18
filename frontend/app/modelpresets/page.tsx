import type { FC } from 'react';
import { useEffect, useState } from 'react';

import { DefaultProviderName, type ProviderName, type ProviderPreset } from '@/models/aimodelmodel';
import type { AISetting } from '@/models/settingmodel';

import { modelPresetStoreAPI, providerSetAPI, settingstoreAPI } from '@/apis/baseapi';
import { MergeInbuiltModelsWithPresets } from '@/apis/modelpresetstore_helper';

import { omitManyKeys } from '@/lib/obj_utils';

import ActionDeniedAlert from '@/components/action_denied';
import DownloadButton from '@/components/download_button';

import ProviderPresetCard from '@/modelpresets/provider_preset';

const ModelPresetsPage: FC = () => {
	/* ── state ─────────────────────────────────────────────── */
	const [aiSettings, setAISettings] = useState<Record<ProviderName, AISetting>>({});
	const [providerPresets, setProviderPresets] = useState<Record<ProviderName, ProviderPreset>>({});
	const [inbuiltProviderInfo, setInbuiltProviderInfo] = useState<Record<ProviderName, ProviderPreset>>({});

	const [defaultProvider, setDefaultProvider] = useState<ProviderName>(DefaultProviderName);

	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);

	const [showActionDenied, setShowActionDenied] = useState(false);
	const [actionDeniedMsg, setActionDeniedMsg] = useState('');

	/* ── initial load ──────────────────────────────────────── */
	useEffect(() => {
		(async () => {
			setLoading(true);
			setError(null);
			try {
				const settings = await settingstoreAPI.getAllSettings();
				const info = await providerSetAPI.getConfigurationInfo();
				const schema = await modelPresetStoreAPI.getAllModelPresets();

				setAISettings(settings.aiSettings);
				setDefaultProvider(settings.app.defaultProvider);
				setInbuiltProviderInfo(info.inbuiltProviderModels);

				setProviderPresets(MergeInbuiltModelsWithPresets(schema.providerPresets, info.inbuiltProviderModels));
			} catch (err) {
				console.error('Error loading model presets:', err, (err as Error).stack || '');
				setError('Failed to load model presets. Please try again.');
			} finally {
				setLoading(false);
			}
		})();
	}, []);

	/* ── handlers ──────────────────────────────────────────── */
	const handlePresetChange = (provider: ProviderName, newPreset: ProviderPreset) => {
		setProviderPresets(prev => ({ ...prev, [provider]: newPreset }));
	};

	const handleProviderDelete = async (provider: ProviderName) => {
		if (provider === defaultProvider) {
			setActionDeniedMsg('Cannot delete the default provider. Please select a different default provider first.');
			setShowActionDenied(true);
			return;
		}

		try {
			setProviderPresets(prev => omitManyKeys(prev, [provider]));
			setAISettings(prev => omitManyKeys(prev, [provider]));

			await modelPresetStoreAPI.deleteProviderPreset(provider);
			await settingstoreAPI.deleteAISetting(provider);
		} catch (err) {
			console.error('Failed to delete provider:', err, (err as Error).stack || '');
			setActionDeniedMsg('Failed to delete provider. Please try again.');
			setShowActionDenied(true);
		}
	};

	/* download helper */
	const fetchValue = async () => {
		try {
			const schema = await modelPresetStoreAPI.getAllModelPresets();
			return JSON.stringify(schema, null, 2);
		} catch (err) {
			console.error('Failed to fetch presets for download:', err, (err as Error).stack || '');
			setActionDeniedMsg('Failed to fetch presets for download.');
			setShowActionDenied(true);
			return '';
		}
	};

	/* ── render ────────────────────────────────────────────── */
	return (
		<div className="flex flex-col items-center w-full h-full">
			{/* header */}
			<div className="w-full flex justify-center fixed top-2 z-10">
				<div className="w-10/12 lg:w-2/3 flex items-center justify-between p-2">
					<h1 className="text-xl font-semibold text-center flex-grow">Model Presets</h1>
					<DownloadButton
						title="Download Presets"
						language="json"
						valueFetcher={fetchValue}
						size={24}
						fileprefix="presets"
						className="btn btn-sm btn-ghost"
					/>
				</div>
			</div>

			{/* body */}
			<div className="flex flex-col items-center w-full grow mt-20 overflow-y-auto">
				<div className="flex flex-col space-y-4 w-11/12 lg:w-2/3">
					{loading && <p className="text-center text-sm mt-8">Loading model presets...</p>}
					{error && <p className="text-center text-error mt-8">{error}</p>}

					{!loading && !error && Object.keys(providerPresets).length === 0 && (
						<p className="text-center text-sm mt-8">No providers configured yet.</p>
					)}

					{!loading &&
						!error &&
						Object.entries(providerPresets).map(([provider, preset]) => (
							<ProviderPresetCard
								key={provider}
								provider={provider}
								isEnabled={aiSettings[provider].isEnabled}
								preset={preset}
								inbuiltProviderPresets={inbuiltProviderInfo[provider].modelPresets}
								defaultProvider={defaultProvider}
								onPresetChange={handlePresetChange}
								onProviderDelete={handleProviderDelete}
							/>
						))}
				</div>
			</div>

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
		</div>
	);
};

export default ModelPresetsPage;
