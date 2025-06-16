import { type FC, useEffect, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import {
	DefaultModelName,
	DefaultProviderName,
	type ModelName,
	type ModelPreset,
	type ProviderName,
} from '@/models/aiprovidermodel';
import type { AISetting } from '@/models/settingmodel';

import { providerSetAPI, settingstoreAPI } from '@/apis/baseapi';
import {
	AddAISetting,
	DeleteAISetting,
	MergeInbuiltModelsWithSettings,
	SetAppSettings,
} from '@/apis/settingstore_helper';

import DownloadButton from '@/components/download_button';
import Dropdown from '@/components/dropdown';
import ThemeSwitch from '@/components/theme_switch';

import AddProviderModal from '@/settings/provider_add_modal';
import AISettingsCard from '@/settings/provider_card';

const defaultAISettings: Record<ProviderName, AISetting> = {
	[DefaultProviderName]: {
		isEnabled: true,
		apiKey: '',
		origin: '',
		chatCompletionPathPrefix: '',
		defaultModel: DefaultModelName,
		modelSettings: {},
	},
};

const SettingsPage: FC = () => {
	const [defaultProvider, setComponentDefaultProvider] = useState<ProviderName>(DefaultProviderName);
	const [aiSettings, setAISettings] = useState<Record<string, AISetting>>(defaultAISettings);
	const [isAddProviderModalOpen, setIsAddProviderModalOpen] = useState(false);
	const [inbuiltProviderInfo, setInbuiltProviderInfo] = useState<Record<ProviderName, Record<ModelName, ModelPreset>>>(
		{}
	);

	useEffect(() => {
		(async () => {
			const settings = await settingstoreAPI.getAllSettings();
			const info = await providerSetAPI.getConfigurationInfo();
			const enabledProviders = Object.keys(settings.aiSettings).filter(
				provider => settings.aiSettings[provider].isEnabled
			);

			if (enabledProviders.length === 0) {
				enabledProviders.push(DefaultProviderName);
			}
			const defaultProv = settings.app.defaultProvider;

			setInbuiltProviderInfo(info.inbuiltProviderModels);
			setComponentDefaultProvider(enabledProviders.includes(defaultProv) ? defaultProv : enabledProviders[0]);
			const newSettings = MergeInbuiltModelsWithSettings(settings.aiSettings, info.inbuiltProviderModels);
			setAISettings(newSettings);
		})();
	}, []);

	const handleDefaultProviderChange = async (value: ProviderName) => {
		setComponentDefaultProvider(value);
		await SetAppSettings(value);
	};

	const handleProviderSettingChange = (provider: ProviderName, updatedSettings: AISetting) => {
		setAISettings(prev => ({
			...prev,
			[provider]: updatedSettings,
		}));
	};

	const handleAddProviderSubmit = async (providerName: ProviderName, newProviderSettings: AISetting) => {
		// Build the updated settings
		const updatedSettings = {
			...aiSettings,
			[providerName]: newProviderSettings,
		};

		// Update React state with the same object
		setAISettings(updatedSettings);

		// Persist to setting store
		await AddAISetting(providerName, newProviderSettings);
		console.log(`Provider "${providerName}" added successfully.`);
	};

	const handleRemoveProvider = async (providerName: ProviderName) => {
		if (providerName === defaultProvider) {
			// This should never happen
			console.log(`Provider "${providerName}" not removed.`);
			return;
		}
		// Remove from local state
		const newAISettings = Object.fromEntries(Object.entries(aiSettings).filter(([key]) => key !== providerName));
		setAISettings(newAISettings);

		await DeleteAISetting(providerName);
		console.log(`Provider "${providerName}" removed successfully.`);
	};

	const fetchValue = async (): Promise<string> => {
		// For download button
		const value = JSON.stringify(
			{
				app: {
					defaultProvider,
				},
				aiSettings: aiSettings,
			},
			null,
			2
		);
		return value;
	};

	return (
		<div className="flex flex-col items-center w-full h-full overflow-hidden">
			<div className="w-full flex justify-center bg-transparent fixed top-2">
				<div className="w-10/12 lg:w-2/3">
					<div className="flex flex-row items-center justify-between p-2 mt-2 bg-transparent">
						<div className="grow text-center">
							<h1 className="text-xl font-semibold">Settings</h1>
						</div>
						<div className="ml-auto lg:pr-8">
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
				</div>
			</div>

			<div className="flex flex-col items-center w-full grow bg-transparent mt-18">
				<div
					className="flex flex-col items-center w-full grow overflow-y-auto"
					style={{ maxHeight: `calc(100vh - 128px)` }}
				>
					<div className="flex flex-col space-y-4 w-11/12 lg:w-2/3">
						{/* Theme Switch Card */}
						<div className="bg-base-100 rounded-xl shadow-lg px-4 py-2 mb-8">
							<div className="grid grid-cols-12 gap-4 items-center">
								<div className="col-span-3 text-sm font-medium">Theme</div>
								<div className="col-span-9">
									<ThemeSwitch />
								</div>
							</div>
						</div>

						{/* Default Provider + Add Provider */}
						<div className="bg-base-100 rounded-xl shadow-lg px-4 py-2 mb-8">
							<div className="grid grid-cols-12 gap-4 items-center">
								<label className="col-span-3 text-sm font-medium">Default Provider</label>
								<div className="col-span-6">
									<Dropdown<ProviderName>
										dropdownItems={aiSettings}
										selectedKey={defaultProvider}
										onChange={handleDefaultProviderChange}
										filterDisabled={true}
										title="Select Default Provider"
										getDisplayName={(key: string) => {
											return key.charAt(0).toUpperCase() + key.slice(1);
										}}
									/>
								</div>
								<div className="col-span-3 flex justify-end">
									<button
										className="btn btn-md btn-ghost rounded-2xl flex items-center"
										onClick={() => {
											setIsAddProviderModalOpen(true);
										}}
									>
										<FiPlus size={16} /> Add Provider
									</button>
								</div>
							</div>
						</div>

						{/* AI Settings Cards */}
						{Object.keys(aiSettings).map(providerStr => {
							return (
								<div key={providerStr} className="rounded-xl">
									<AISettingsCard
										provider={providerStr}
										settings={aiSettings[providerStr]}
										aiSettings={aiSettings}
										defaultProvider={defaultProvider}
										inbuiltProviderModels={
											providerStr in inbuiltProviderInfo ? inbuiltProviderInfo[providerStr] : undefined
										}
										onProviderSettingChange={handleProviderSettingChange}
										onProviderDelete={handleRemoveProvider}
									/>
								</div>
							);
						})}
					</div>
				</div>
			</div>

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
