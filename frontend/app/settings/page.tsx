'use client';

import { providerSetAPI, settingstoreAPI } from '@/backendapibase';
import { loadProviderSettings, updateProviderAISettings } from '@/backendapihelper/settings_helper';
import DownloadButton from '@/components/download_button';
import ThemeSwitch from '@/components/theme_switch';
import { DefaultModelName, DefaultProviderName, ProviderName } from '@/models/aiprovidermodel';
import { AISetting } from '@/models/settingmodel';
import ProviderDropdown from '@/settings/ai_provider'; // Import the new component
import AISettingsCard from '@/settings/ai_settings';
import { FC, useEffect, useState } from 'react';

const defaultAISettings: Record<ProviderName, AISetting> = {
	[DefaultProviderName]: {
		isEnabled: true,
		apiKey: '',
		defaultModel: DefaultModelName,
		defaultOrigin: '',
		defaultTemperature: 0.0,
		additionalSettings: {},
	},
};

const SettingsPage: FC = () => {
	const [defaultProvider, setComponentDefaultProvider] = useState(DefaultProviderName);
	const [aiSettings, setAISettings] = useState(defaultAISettings);

	useEffect(() => {
		(async () => {
			const settings = await loadProviderSettings();
			if (settings) {
				const enabledProviders = Object.keys(settings.aiSettings).filter(
					provider => settings.aiSettings[provider as ProviderName]?.isEnabled
				);

				if (enabledProviders.length === 0) {
					enabledProviders.push(DefaultProviderName);
				}

				const defaultProvider = settings.app.defaultProvider as ProviderName;
				setComponentDefaultProvider(
					enabledProviders.includes(defaultProvider)
						? (defaultProvider as ProviderName)
						: (enabledProviders[0] as ProviderName)
				);

				setAISettings(settings.aiSettings);
			}
		})();
	}, []);

	const handleDefaultProviderChange = (value: ProviderName) => {
		setComponentDefaultProvider(value);
		providerSetAPI.setDefaultProvider(value);
		// console.log('Set a new default provider', value);
		settingstoreAPI.setSetting('app.defaultProvider', value);
	};

	const handleAISettingsChange = (provider: keyof typeof aiSettings, key: string, value: any) => {
		const updatedSettings = {
			...aiSettings,
			[provider]: {
				...aiSettings[provider],
				[key]: value,
			},
		};
		setAISettings(updatedSettings);
	};

	const handleSaveAISettings = async (provider: keyof typeof aiSettings, key: string, value: any) => {
		if (key === 'isEnabled') {
			const enabledProviders = Object.keys(aiSettings).filter(
				provider => aiSettings[provider as ProviderName]?.isEnabled
			);

			if (enabledProviders.length === 1 && !value) {
				return;
			}
		}

		await settingstoreAPI.setSetting(`aiSettings.${provider}.${key}`, value);
		updateProviderAISettings(provider as ProviderName, aiSettings[provider]);
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
								className="btn btn-sm bg-transparent border-none shadow-none"
							/>
						</div>
					</div>
				</div>
			</div>

			<div className="flex flex-col items-center w-full grow bg-transparent mt-20">
				<div className="w-full grow flex justify-center overflow-y-auto" style={{ maxHeight: `calc(100vh - 128px)` }}>
					<div className="flex flex-col space-y-4 w-11/12 lg:w-2/3">
						{/* Theme Switch Card */}
						<div className="bg-base-100 rounded-lg shadow-lg px-4 py-2">
							<div className="grid grid-cols-12 gap-4 items-center">
								<div className="col-span-3 text-sm font-medium">Theme</div>
								<div className="col-span-9">
									<ThemeSwitch />
								</div>
							</div>
						</div>

						{/* Default Provider Card */}
						<div className="bg-base-100 rounded-lg shadow-lg px-4 py-2">
							<div className="grid grid-cols-12 gap-4 items-center">
								<div className="col-span-3 text-sm font-medium">Default Provider</div>
								<div className="col-span-9">
									<ProviderDropdown
										aiSettings={aiSettings}
										defaultProvider={defaultProvider}
										onProviderChange={handleDefaultProviderChange}
									/>
								</div>
							</div>
						</div>

						{/* AI Settings Cards */}
						{Object.keys(aiSettings).map(providerStr => {
							const typedProvider = providerStr as ProviderName;
							const oneSettings = aiSettings[typedProvider];
							return (
								<div key={typedProvider} className=" rounded-lg">
									<AISettingsCard
										provider={typedProvider}
										settings={oneSettings}
										onChange={(key, value) => handleAISettingsChange(typedProvider, key, value)}
										onSave={(key, value) => handleSaveAISettings(typedProvider, key, value)}
										aiSettings={aiSettings}
									/>
								</div>
							);
						})}
					</div>
				</div>
			</div>
		</div>
	);
};

export default SettingsPage;
