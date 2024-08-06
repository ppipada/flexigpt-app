'use client';

import { setSetting } from '@/api/settings';
import DownloadButton from '@/components/download_button';
import ThemeSwitch from '@/components/theme_switch';
import { loadProviderSettings, updateProviderAISettings } from '@/lib/loadSettings';
import AISettingsCard from '@/settings/ai_settings';
import { ALL_AI_PROVIDERS, ProviderName, providerSet } from 'aiprovider';
import { FC, useEffect, useState } from 'react';
import { defaultAISettings } from 'settingmodel';

const SettingsPage: FC = () => {
	const [defaultProvider, setDefaultProvider] = useState(providerSet.getDefaultProvider());
	const [aiSettings, setAISettings] = useState(defaultAISettings);

	useEffect(() => {
		(async () => {
			const settings = await loadProviderSettings();
			if (settings) {
				const defaultProvider = settings.app.defaultProvider as ProviderName;
				setDefaultProvider(defaultProvider);

				setAISettings({
					anthropic: settings[ProviderName.ANTHROPIC],
					google: settings[ProviderName.GOOGLE],
					huggingface: settings[ProviderName.HUGGINGFACE],
					llamacpp: settings[ProviderName.LLAMACPP],
					openai: settings[ProviderName.OPENAI],
				});
			}
		})();
	}, []);

	const handleDefaultProviderChange = (value: ProviderName) => {
		setDefaultProvider(value);
		providerSet.setDefaultProvider(defaultProvider);
		setSetting('app.defaultProvider', value);
	};

	const handleAISettingsChange = (provider: keyof typeof aiSettings, key: string, value: any) => {
		const updatedSettings = {
			...aiSettings,
			[provider]: {
				...aiSettings[provider],
				[key]: value,
			},
		};
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		setAISettings(prevState => {
			return updatedSettings;
		});
	};

	const handleSaveAISettings = async (provider: keyof typeof aiSettings, key: string, value: any) => {
		await setSetting(`${provider}.${key}`, value);
		updateProviderAISettings(provider as ProviderName, aiSettings[provider]);
	};

	const fetchValue = async (): Promise<string> => {
		const value = JSON.stringify(
			{
				app: {
					defaultProvider,
				},
				[ProviderName.ANTHROPIC]: aiSettings.anthropic,
				[ProviderName.GOOGLE]: aiSettings.google,
				[ProviderName.HUGGINGFACE]: aiSettings.huggingface,
				[ProviderName.LLAMACPP]: aiSettings.llamacpp,
				[ProviderName.OPENAI]: aiSettings.openai,
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
						<div className="flex-grow text-center">
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

			<div className="flex flex-col items-center w-full flex-grow bg-transparent  mt-20">
				<div
					className="w-full flex-grow flex justify-center overflow-y-auto"
					style={{ maxHeight: `calc(100vh - 128px` }}
				>
					<div className="w-11/12 lg:w-2/3">
						<div className="bg-base-100 rounded-lg shadow-lg p-4 mb-4">
							<h3 className="text-xl font-semibold mb-8">App</h3>
							<div className="grid grid-cols-12 gap-4 mb-4 items-center">
								<h3 className="col-span-3 text-sm font-medium">Theme</h3>
								<div className="col-span-9">
									<ThemeSwitch />
								</div>
							</div>
							<div className="grid grid-cols-12 gap-4 mb-4 items-center">
								<label className="col-span-3 text-sm font-medium">Default Provider</label>
								<select
									className="select select-bordered col-span-9 w-full rounded-lg min-h-2 h-10"
									value={defaultProvider}
									onChange={e => handleDefaultProviderChange(e.target.value as ProviderName)}
								>
									{Object.keys(ALL_AI_PROVIDERS).map(provider => (
										<option key={provider} value={provider}>
											{provider.charAt(0).toUpperCase() + provider.slice(1)}
										</option>
									))}
								</select>
							</div>
						</div>
						<div className="w-full flex-1 pb-4">
							{Object.keys(aiSettings).map(providerStr => {
								const typedProvider = providerStr as ProviderName;
								const oneSettings = aiSettings[typedProvider];
								return (
									<AISettingsCard
										key={typedProvider}
										provider={typedProvider}
										settings={oneSettings}
										onChange={(key, value) => handleAISettingsChange(typedProvider, key, value)}
										onSave={(key, value) => handleSaveAISettings(typedProvider, key, value)}
									/>
								);
							})}
						</div>
					</div>
				</div>
			</div>
		</div>
	);
};

export default SettingsPage;
