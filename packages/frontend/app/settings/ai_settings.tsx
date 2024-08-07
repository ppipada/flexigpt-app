import { ALL_AI_PROVIDERS, ProviderName } from 'aiprovidermodel';
import { FC } from 'react';
import { AISetting } from 'settingmodel';

interface AISettingsCardProps {
	provider: ProviderName;
	settings: AISetting;
	onChange: (key: string, value: any) => void;
	onSave: (key: string, value: any) => void;
}

const AISettingsCard: FC<AISettingsCardProps> = ({ provider, settings, onChange, onSave }) => {
	const providerinfo = ALL_AI_PROVIDERS[provider];

	return (
		<div className="bg-base-100 rounded-lg shadow-lg p-4 mb-4">
			<h3 className="text-xl font-semibold mb-8 capitalize">{provider}</h3>

			{/* API Key */}
			<div className="grid grid-cols-12 gap-4 mb-4 items-center">
				<label
					className="col-span-3 text-sm font-medium text-left tooltip"
					data-tip={providerinfo?.getDescription('apiKey')}
				>
					API Key
				</label>
				<input
					type="password"
					className="input input-bordered col-span-9 w-full h-10 rounded-lg px-4 py-2"
					style={{ fontSize: '14px' }}
					value={settings.apiKey}
					onChange={e => onChange('apiKey', e.target.value)}
					onBlur={e => onSave('apiKey', e.target.value)}
				/>
			</div>

			{/* Default Model */}
			<div className="grid grid-cols-12 gap-4 mb-4 items-center">
				<label
					className="col-span-3 text-sm font-medium text-left tooltip"
					data-tip={providerinfo?.getDescription('defaultModel')}
				>
					Default Model
				</label>
				<input
					type="text"
					className="input input-bordered col-span-9 w-full h-10 rounded-lg px-4 py-2"
					style={{ fontSize: '14px' }}
					value={settings.defaultModel}
					onChange={e => onChange('defaultModel', e.target.value)}
					onBlur={e => onSave('defaultModel', e.target.value)}
				/>
			</div>

			{/* Default Temperature */}
			<div className="grid grid-cols-12 gap-4 mb-4 items-center">
				<label
					className="col-span-3 text-sm font-medium text-left tooltip"
					data-tip={providerinfo?.getDescription('defaultTemperature')}
				>
					Default Temperature
				</label>
				<input
					type="number"
					step="0.01"
					min="0"
					max="1"
					className="input input-bordered col-span-9 w-full h-10 rounded-lg px-4 py-2"
					style={{ fontSize: '14px' }}
					value={settings.defaultTemperature}
					onChange={e => onChange('defaultTemperature', parseFloat(e.target.value))}
					onBlur={e => onSave('defaultTemperature', parseFloat(e.target.value))}
				/>
			</div>

			{/* Default Origin */}
			<div className="grid grid-cols-12 gap-4 mb-4 items-center">
				<label
					className="col-span-3 text-sm font-medium text-left tooltip"
					data-tip={providerinfo?.getDescription('defaultOrigin')}
				>
					Default Origin
				</label>
				<input
					type="text"
					className="input input-bordered col-span-9 w-full h-10 rounded-lg px-4 py-2"
					style={{ fontSize: '14px' }}
					value={settings.defaultOrigin}
					onChange={e => onChange('defaultOrigin', e.target.value)}
					onBlur={e => onSave('defaultOrigin', e.target.value)}
				/>
			</div>
		</div>
	);
};

export default AISettingsCard;
