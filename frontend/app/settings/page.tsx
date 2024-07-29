'use client';

import { getAllSettings, providers, setSetting } from '@/api/settings';
import DownloadButton from '@/components/DownloadButton';
import ThemeSwitch from '@/components/ThemeSwitch';
import React, { useEffect, useState } from 'react';
import AISettingsCard from './AISettings';

const SettingsPage: React.FC = () => {
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
		<div className="flex flex-col w-full h-screen mb-16">
			<div className="absolute top-0 p-4 shadow-none z-10 w-full">
				<div className="flex items-center justify-center h-16 relative">
					<h1 className="absolute left-1/2 transform -translate-x-1/2 text-2xl font-bold">Settings</h1>
					<div className="ml-auto">
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
							className="btn btn-sm bg-transparent border-none flex items-center shadow-none"
						/>
					</div>
				</div>
			</div>

			<div className="flex-1 mt-20 overflow-y-auto w-full flex justify-center p-4">
				<div className="w-full lg:w-2/3">
					<div className="bg-base-100 rounded-lg shadow-lg p-6 mb-6">
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
								className="select select-bordered col-span-9 w-full"
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

					{Object.keys(aiSettings).map(provider => {
						const typedProvider = provider as keyof typeof aiSettings;
						return (
							<AISettingsCard
								key={provider}
								provider={provider}
								settings={aiSettings[typedProvider]}
								onChange={(key, value) => handleAISettingsChange(typedProvider, key, value)}
								onSave={(key, value) => handleSaveAISettings(typedProvider, key, value)}
								additionalSettings={aiSettings[typedProvider].additionalSettings}
								onAdditionalSettingsChange={value => handleAISettingsChange(typedProvider, 'additionalSettings', value)}
							/>
						);
					})}
				</div>
			</div>
		</div>
	);
};

export default SettingsPage;
