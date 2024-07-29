'use client';

import { getAllSettings, setSetting } from '@/api/settings';

import DownloadButton from '@/components/DownloadButton';
import ThemeSwitch from '@/components/ThemeSwitch';
import { FC, useEffect, useState } from 'react';
import { providers } from 'sharedpkg/settings/consts';
import AISettingsCard from './AISettings';

const SettingsPage: FC = () => {
	const [defaultProvider, setDefaultProvider] = useState('');
	const [aiSettings, setAISettings] = useState({
		openai: {
			apiKey: '',
			defaultModel: '',
			defaultOrigin: '',
			defaultTemperature: 0.1,
			additionalSettings: '{}',
		},
		anthropic: {
			apiKey: '',
			defaultModel: '',
			defaultOrigin: '',
			defaultTemperature: 0.1,
			additionalSettings: '{}',
		},
		huggingface: {
			apiKey: '',
			defaultModel: '',
			defaultOrigin: '',
			defaultTemperature: 0.1,
			additionalSettings: '{}',
		},
		googlegl: {
			apiKey: '',
			defaultModel: '',
			defaultOrigin: '',
			defaultTemperature: 0.1,
			additionalSettings: '{}',
		},
		llamacpp: {
			apiKey: '',
			defaultModel: '',
			defaultOrigin: '',
			defaultTemperature: 0.1,
			additionalSettings: '{}',
		},
	});

	useEffect(() => {
		(async () => {
			const settings = await getAllSettings();
			if (settings) {
				setDefaultProvider(settings.app.defaultProvider);
				setAISettings({
					openai: settings.openai,
					anthropic: settings.anthropic,
					huggingface: settings.huggingface,
					googlegl: settings.googlegl,
					llamacpp: settings.llamacpp,
				});
			}
		})();
	}, []);

	const handleDefaultProviderChange = (value: string) => {
		setDefaultProvider(value);
		setSetting('app.defaultProvider', value);
	};

	const handleAISettingsChange = (provider: keyof typeof aiSettings, key: string, value: any) => {
		setAISettings(prevState => ({
			...prevState,
			[provider]: {
				...prevState[provider],
				[key]: value,
			},
		}));
	};

	const handleSaveAISettings = async (provider: keyof typeof aiSettings, key: string, value: any) => {
		await setSetting(`${provider}.${key}`, value);
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
								value={JSON.stringify(
									{
										app: {
											defaultProvider,
										},
										openai: aiSettings.openai,
										anthropic: aiSettings.anthropic,
										huggingface: aiSettings.huggingface,
										googlegl: aiSettings.googlegl,
										llamacpp: aiSettings.llamacpp,
									},
									null,
									2
								)}
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
									onChange={e => handleDefaultProviderChange(e.target.value)}
								>
									{providers.map(provider => (
										<option key={provider} value={provider}>
											{provider.charAt(0).toUpperCase() + provider.slice(1)}
										</option>
									))}
								</select>
							</div>
						</div>
						<div className="w-full flex-1 pb-4">
							{Object.keys(aiSettings).map(provider => {
								const typedProvider = provider as keyof typeof aiSettings;
								const oneSettings = aiSettings[typedProvider];
								return (
									<AISettingsCard
										key={provider}
										provider={provider}
										settings={oneSettings}
										onChange={(key, value) => handleAISettingsChange(typedProvider, key, value)}
										onSave={(key, value) => handleSaveAISettings(typedProvider, key, value)}
										additionalSettings={aiSettings[typedProvider].additionalSettings}
										onAdditionalSettingsChange={value =>
											handleAISettingsChange(typedProvider, 'additionalSettings', value)
										}
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
