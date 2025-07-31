import type { FC } from 'react';
import { useEffect, useState } from 'react';

import { type ProviderName, type ProviderPreset } from '@/spec/modelpreset';
import type { AISetting } from '@/spec/setting';

import { settingstoreAPI } from '@/apis/baseapi';
import { getAllProviderPresetsMap, getBuiltInPresets } from '@/apis/modelpresetstore_helper';

import ActionDeniedAlert from '@/components/action_denied';
import DownloadButton from '@/components/download_button';

import ProviderPresetCard from '@/modelpresets/provider_presets_card';

const ModelPresetPage: FC = () => {
	/* ── state ─────────────────────────────────────────────── */
	const [aiSettings, setAISettings] = useState<Record<ProviderName, AISetting>>({});
	const [providerPresets, setProviderPresets] = useState<Record<ProviderName, ProviderPreset>>({});
	const [BuiltInProviderInfo, setBuiltInProviderInfo] = useState<Record<ProviderName, ProviderPreset>>({});

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
				const presets = await getAllProviderPresetsMap();
				const builtInPresets = getBuiltInPresets(presets);

				setAISettings(settings.aiSettings);
				setBuiltInProviderInfo(builtInPresets);
				setProviderPresets(presets);
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

	/* download helper */
	const fetchValue = async () => {
		try {
			const presets = await getAllProviderPresetsMap();
			return JSON.stringify(presets, null, 2);
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
			<div className="w-full flex justify-center fixed top-8">
				<div className="w-10/12 lg:w-2/3 flex items-center justify-between p-2">
					<h1 className="text-xl font-semibold text-center flex-grow">Model Presets</h1>
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
				style={{ maxHeight: `calc(100vh - 128px)` }}
			>
				<div className="flex flex-col space-y-4 w-5/6 xl:w-2/3">
					{loading && <p className="text-center text-sm mt-8">Loading model presets...</p>}
					{error && <p className="text-center text-error mt-8">{error}</p>}

					{!loading && !error && Object.keys(providerPresets).length === 0 && (
						<p className="text-center text-sm mt-8">No model presets configured yet.</p>
					)}

					{!loading &&
						!error &&
						Object.entries(providerPresets).map(([provider, preset]) => (
							<ProviderPresetCard
								key={provider}
								provider={provider}
								isEnabled={aiSettings[provider].isEnabled}
								preset={preset}
								inbuiltProviderPresets={BuiltInProviderInfo[provider].modelPresets}
								onPresetChange={handlePresetChange}
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

export default ModelPresetPage;
