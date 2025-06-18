import type { FC } from 'react';
import { useEffect, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import type { ProviderName, ProviderPreset } from '@/models/modelpresetsmodel';
import { DefaultProviderName } from '@/models/modelpresetsmodel';
import type { AISetting } from '@/models/settingmodel';

import { providerSetAPI, settingstoreAPI } from '@/apis/baseapi';
import { AddAISetting, DeleteAISetting, SetAppSettings } from '@/apis/settingstore_helper';

import { omitManyKeys } from '@/lib/obj_utils';

import DownloadButton from '@/components/download_button';
import Dropdown from '@/components/dropdown';
import ThemeSwitch from '@/components/theme_switch';

import AddProviderModal from '@/settings/provider_add_modal';
import ProviderSettingsCard from '@/settings/provider_card';

const SettingsPage: FC = () => {
	/* ── state ─────────────────────────────────────────────── */
	const [defaultProvider, setDefaultProvider] = useState<ProviderName>(DefaultProviderName);
	const [aiSettings, setAISettings] = useState<Record<ProviderName, AISetting>>({});
	const [inbuiltProviderInfo, setInbuiltProviderInfo] = useState<Record<ProviderName, ProviderPreset>>({});

	const [isAddProviderModalOpen, setIsAddProviderModalOpen] = useState(false);

	/* ── initial load ──────────────────────────────────────── */
	useEffect(() => {
		(async () => {
			const settings = await settingstoreAPI.getAllSettings();
			const info = await providerSetAPI.getConfigurationInfo();
			// const schema = await modelPresetStoreAPI.getAllModelPresets();

			setInbuiltProviderInfo(info.inbuiltProviderModels);
			setDefaultProvider(settings.app.defaultProvider);

			setAISettings(settings.aiSettings);
		})();
	}, []);

	/* ── handlers ──────────────────────────────────────────── */
	const handleDefaultProviderChange = async (value: ProviderName) => {
		setDefaultProvider(value);
		await SetAppSettings(value);
	};

	const handleProviderSettingChange = (provider: ProviderName, updated: AISetting) => {
		setAISettings(prev => ({ ...prev, [provider]: updated }));
	};

	const handleAddProviderSubmit = async (providerName: ProviderName, newSettings: AISetting) => {
		setAISettings(prev => ({ ...prev, [providerName]: newSettings }));

		await AddAISetting(providerName, newSettings);
	};

	const handleRemoveProvider = async (providerName: ProviderName) => {
		if (providerName === defaultProvider) return;

		setAISettings(prev => {
			return omitManyKeys(prev, [providerName]);
		});

		await DeleteAISetting(providerName);
	};

	const enabledProviders = Object.entries(aiSettings)
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		.filter(([_, s]) => s.isEnabled)
		.map(([name]) => name);

	/* download settings helper */
	const fetchValue = async () =>
		JSON.stringify(
			{
				app: { defaultProvider },
				aiSettings,
			},
			null,
			2
		);

	/* ── render ────────────────────────────────────────────── */
	return (
		<div className="flex flex-col items-center w-full h-full">
			{/* sticky header bar */}
			<div className="w-full flex justify-center fixed top-8">
				<div className="w-10/12 lg:w-2/3 flex items-center justify-between p-2">
					<h1 className="text-xl font-semibold text-center flex-grow">Settings</h1>
					<DownloadButton
						title="Download Settings"
						language="json"
						valueFetcher={fetchValue}
						size={24}
						fileprefix="settings"
						className="btn btn-sm btn-ghost"
					/>
				</div>
			</div>

			{/* content */}
			<div
				className="flex flex-col items-center w-full grow mt-24 overflow-y-auto"
				style={{ maxHeight: `calc(100vh - 128px)` }}
			>
				<div className="flex flex-col space-y-4 w-5/6 xl:w-2/3">
					{/* ── Theme switch ─ */}
					<div className="bg-base-100 rounded-xl shadow-lg px-4 py-2 mb-8">
						<div className="grid grid-cols-12 items-center gap-4">
							<div className="col-span-3 text-sm font-medium">Theme</div>
							<div className="col-span-9">
								<ThemeSwitch />
							</div>
						</div>
					</div>

					{/* ── Default provider & add ─ */}
					<div className="bg-base-100 rounded-xl shadow-lg px-4 py-2 mb-8">
						<div className="grid grid-cols-12 items-center gap-4">
							<label className="col-span-3 text-sm font-medium">Default Provider</label>
							<div className="col-span-6">
								<Dropdown<ProviderName>
									dropdownItems={aiSettings}
									selectedKey={defaultProvider}
									onChange={handleDefaultProviderChange}
									filterDisabled={true}
									title="Select Provider"
									getDisplayName={(key: string) => {
										return key.charAt(0).toUpperCase() + key.slice(1);
									}}
								/>
							</div>

							<div className="col-span-3 flex justify-end">
								<button
									className="btn btn-ghost rounded-2xl flex items-center"
									onClick={() => {
										setIsAddProviderModalOpen(true);
									}}
								>
									<FiPlus /> <span className="ml-1">Add Provider</span>
								</button>
							</div>
						</div>
					</div>

					{/* ── provider cards ─ */}
					{Object.keys(aiSettings).map(p => (
						<ProviderSettingsCard
							key={p}
							provider={p}
							settings={aiSettings[p]}
							defaultProvider={defaultProvider}
							inbuiltProvider={Boolean(inbuiltProviderInfo[p])}
							enabledProviders={enabledProviders}
							onProviderSettingChange={handleProviderSettingChange}
							onProviderDelete={handleRemoveProvider}
						/>
					))}
				</div>
			</div>

			{/* modal */}
			{isAddProviderModalOpen && (
				<AddProviderModal
					isOpen={isAddProviderModalOpen}
					onClose={() => {
						setIsAddProviderModalOpen(false);
					}}
					onSubmit={handleAddProviderSubmit}
					existingProviderNames={Object.keys(aiSettings)}
				/>
			)}
		</div>
	);
};

export default SettingsPage;
