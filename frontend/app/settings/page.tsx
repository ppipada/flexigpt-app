'use client';

import { SettingsSchema, getAllSettings, setSetting } from '@/api/settings';
import DownloadButton from '@/components/DownloadButton';
import ThemeSwitch from '@/components/ThemeSwitch';
import React, { useEffect, useState } from 'react';

const SettingsPage: React.FC = () => {
	const [openAiSettings, setOpenAiSettings] = useState<SettingsSchema['openai']>({
		apiKey: '',
		defaultModel: '',
		defaultTemperature: 0.1,
	});

	const [anthropicSettings, setAnthropicSettings] = useState<SettingsSchema['anthropic']>({
		apiKey: '',
		defaultModel: '',
		defaultTemperature: 0.1,
	});

	useEffect(() => {
		(async () => {
			const settings = await getAllSettings();
			if (settings) {
				setOpenAiSettings(settings.openai);
				setAnthropicSettings(settings.anthropic);
			}
		})();
	}, []);

	const handleChange = <T,>(setter: React.Dispatch<React.SetStateAction<T>>, key: keyof T, value: any) => {
		setter(prevState => ({
			...prevState,
			[key]: value,
		}));
	};

	const handleSave = async (category: 'openai' | 'anthropic', key: string, value: any) => {
		await setSetting(`${category}.${key}`, value);
	};

	return (
		<div className="flex flex-col items-center justify-center p-2">
			<div className="w-11/12 lg:w-2/3 mb-8 mt-8">
				<div className="relative flex items-center justify-center h-full mb-16">
					<h1 className="absolute left-1/2 transform -translate-x-1/2 text-2xl font-bold">Settings</h1>
					<div className="ml-auto">
						<DownloadButton
							title="Download Settings"
							language="json"
							value={JSON.stringify(
								{
									openai: openAiSettings,
									anthropic: anthropicSettings,
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

				<div className="bg-base-100 rounded-lg shadow-lg p-6 mb-6">
					<div className="grid grid-cols-12 items-center">
						<h3 className="col-span-3 text-xl font-semibold">Theme</h3>
						<div className="col-span-9">
							<ThemeSwitch />
						</div>
					</div>
				</div>

				{/* OpenAI Settings */}
				<div className="bg-base-100 rounded-lg shadow-lg p-6 mb-6">
					<h3 className="text-xl font-semibold mb-8">OpenAI</h3>
					<div className="grid grid-cols-12 gap-4 mb-4 items-center">
						<label className="col-span-3 text-sm font-medium">API Key</label>
						<input
							type="password"
							className="input input-bordered col-span-9 w-full"
							value={openAiSettings.apiKey}
							onChange={e => handleChange(setOpenAiSettings, 'apiKey', e.target.value)}
							onBlur={e => handleSave('openai', 'apiKey', e.target.value)}
						/>
					</div>
					<div className="grid grid-cols-12 gap-4 mb-4 items-center">
						<label className="col-span-3 text-sm font-medium">Default Model</label>
						<input
							type="text"
							className="input input-bordered col-span-9 w-full"
							value={openAiSettings.defaultModel}
							onChange={e => handleChange(setOpenAiSettings, 'defaultModel', e.target.value)}
							onBlur={e => handleSave('openai', 'defaultModel', e.target.value)}
						/>
					</div>
					<div className="grid grid-cols-12 gap-4 mb-4 items-center">
						<label className="col-span-3 text-sm font-medium">Default Temperature</label>
						<input
							type="number"
							step="0.01"
							min="0"
							max="1"
							className="input input-bordered col-span-9 w-full"
							value={openAiSettings.defaultTemperature}
							onChange={e =>
								handleChange(
									setOpenAiSettings,
									'defaultTemperature',
									e.target.value === '' ? '' : Number(e.target.value)
								)
							}
							onBlur={e =>
								handleSave('openai', 'defaultTemperature', e.target.value === '' ? 0.1 : Number(e.target.value))
							}
						/>
					</div>
				</div>

				{/* Anthropic Settings */}
				<div className="bg-base-100 rounded-lg shadow-lg p-6">
					<h3 className="text-xl font-semibold mb-8">Anthropic</h3>
					<div className="grid grid-cols-12 gap-4 mb-4 items-center">
						<label className="col-span-3 text-sm font-medium">API Key</label>
						<input
							type="password"
							className="input input-bordered col-span-9 w-full"
							value={anthropicSettings.apiKey}
							onChange={e => handleChange(setAnthropicSettings, 'apiKey', e.target.value)}
							onBlur={e => handleSave('anthropic', 'apiKey', e.target.value)}
						/>
					</div>
					<div className="grid grid-cols-12 gap-4 mb-4 items-center">
						<label className="col-span-3 text-sm font-medium">Default Model</label>
						<input
							type="text"
							className="input input-bordered col-span-9 w-full"
							value={anthropicSettings.defaultModel}
							onChange={e => handleChange(setAnthropicSettings, 'defaultModel', e.target.value)}
							onBlur={e => handleSave('anthropic', 'defaultModel', e.target.value)}
						/>
					</div>
					<div className="grid grid-cols-12 gap-4 mb-4 items-center">
						<label className="col-span-3 text-sm font-medium">Default Temperature</label>
						<input
							type="number"
							step="0.01"
							min="0"
							max="1"
							className="input input-bordered col-span-9 w-full"
							value={anthropicSettings.defaultTemperature}
							onChange={e =>
								handleChange(
									setAnthropicSettings,
									'defaultTemperature',
									e.target.value === '' ? '' : Number(e.target.value)
								)
							}
							onBlur={e =>
								handleSave('anthropic', 'defaultTemperature', e.target.value === '' ? 0.1 : Number(e.target.value))
							}
						/>
					</div>
				</div>
			</div>
		</div>
	);
};

export default SettingsPage;
