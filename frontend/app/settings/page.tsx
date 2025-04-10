import { type FC, useEffect, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import { DefaultModelName, DefaultProviderName, type ProviderName } from '@/models/aiprovidermodel';
import type { AISetting } from '@/models/settingmodel';

import { settingstoreAPI } from '@/apis/baseapi';
import { AddAISetting, SetAppSettings } from '@/apis/settingstore_helper';

import DownloadButton from '@/components/download_button';
import ThemeSwitch from '@/components/theme_switch';

import AddProviderModal from '@/settings/provider_add_modal';
import AISettingsCard from '@/settings/provider_card';
import ProviderDropdown from '@/settings/provider_dropdown';

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

	useEffect(() => {
		(async () => {
			const settings = await settingstoreAPI.getAllSettings();

			const enabledProviders = Object.keys(settings.aiSettings).filter(
				provider => settings.aiSettings[provider].isEnabled
			);

			if (enabledProviders.length === 0) {
				enabledProviders.push(DefaultProviderName);
			}

			const defaultProv = settings.app.defaultProvider;
			setComponentDefaultProvider(enabledProviders.includes(defaultProv) ? defaultProv : enabledProviders[0]);
			setAISettings(settings.aiSettings);
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
		// Build the updated settings yourself:
		const updatedSettings = {
			...aiSettings,
			[providerName]: newProviderSettings,
		};

		// Update React state with the same object
		setAISettings(updatedSettings);

		// Persist to setting store and update the provider api
		await AddAISetting(providerName, newProviderSettings);
	};

	const fetchValue = async (): Promise<string> => {
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
									<ProviderDropdown
										aiSettings={aiSettings}
										defaultProvider={defaultProvider}
										onProviderChange={handleDefaultProviderChange}
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
										onProviderSettingChange={handleProviderSettingChange}
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
				/>
			)}
		</div>
	);
};

export default SettingsPage;
