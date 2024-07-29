import { FC } from 'react';
import { aiSettingsDescriptions } from 'sharedpkg/settings/consts';

interface AISettingsCardProps {
	provider: string;
	settings: any;
	onChange: (key: string, value: any) => void;
	onSave: (key: string, value: any) => void;
	additionalSettings: string;
	onAdditionalSettingsChange: (value: string) => void;
}

const formatKey = (key: string) => {
	return key
		.replace(/([a-z])([A-Z])/g, '$1 $2')
		.replace(/([A-Z])([A-Z][a-z])/g, '$1 $2')
		.replace(/([0-9])/g, ' $1')
		.split(' ')
		.map(word => (word.toLowerCase() === 'api' ? 'API' : word.charAt(0).toUpperCase() + word.slice(1)))
		.join(' ')
		.trim();
};

function getInputType(key: string, value: any): string {
	if (typeof value === 'string') {
		if (key.toLowerCase().includes('key') || key.toLowerCase().includes('password')) {
			return 'password';
		}
		return 'text';
	}
	return 'number';
}

const AISettingsCard: FC<AISettingsCardProps> = ({
	provider,
	settings,
	onChange,
	onSave,
	additionalSettings,
	onAdditionalSettingsChange,
}) => {
	return (
		<div className="bg-base-100 rounded-lg shadow-lg p-4 mb-4">
			<h3 className="text-xl font-semibold mb-8 capitalize">{provider}</h3>
			{Object.keys(settings)
				.filter(key => key !== 'additionalSettings')
				.map(key => (
					<div key={key} className="grid grid-cols-12 gap-4 mb-4 items-center">
						<label
							className="col-span-3 text-sm font-medium text-left tooltip"
							data-tip={aiSettingsDescriptions[`${provider}.${key}`]}
						>
							{formatKey(key)}
						</label>
						<input
							type={getInputType(key, settings[key])}
							step={typeof settings[key] === 'number' ? '0.01' : undefined}
							min={typeof settings[key] === 'number' ? '0' : undefined}
							max={typeof settings[key] === 'number' ? '1' : undefined}
							className="input input-bordered col-span-9 w-full h-10 rounded-lg px-4 py-2"
							style={{ fontSize: '14px' }}
							value={settings[key]}
							onChange={e => onChange(key, e.target.value)}
							onBlur={e => onSave(key, e.target.value)}
						/>
					</div>
				))}
			<div className="grid grid-cols-12 gap-4 mb-4 items-center">
				<label
					className="col-span-3 text-sm font-medium text-left tooltip"
					data-tip={aiSettingsDescriptions[`${provider}.additionalSettings`]}
				>
					Additional Settings
				</label>
				<textarea
					className="textarea textarea-bordered col-span-9 w-full rounded-lg px-4 py-2"
					rows={2}
					value={additionalSettings}
					onChange={e => onAdditionalSettingsChange(e.target.value)}
				/>
			</div>
		</div>
	);
};

export default AISettingsCard;
