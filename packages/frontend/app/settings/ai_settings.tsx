import { ALL_AI_PROVIDERS, ProviderName } from '@/models/aiprovidermodel';
import { AISetting } from '@/models/settingmodel';
import { FC, useState } from 'react';
import { FiAlertTriangle, FiCheckCircle, FiChevronDown, FiChevronUp, FiXCircle } from 'react-icons/fi';

interface AISettingsCardProps {
	provider: ProviderName;
	settings: AISetting;
	onChange: (key: string, value: any) => void;
	onSave: (key: string, value: any) => void;
	aiSettings: Record<string, AISetting>;
}

const AISettingsCard: FC<AISettingsCardProps> = ({ provider, settings, onChange, onSave, aiSettings }) => {
	const providerinfo = ALL_AI_PROVIDERS[provider];
	const [isExpanded, setIsExpanded] = useState(false);
	const [isEnabled, setIsEnabled] = useState(!!settings.isEnabled);
	const [showModal, setShowModal] = useState(false);

	const toggleExpand = () => {
		if (isEnabled) {
			setIsExpanded(!isExpanded);
		}
	};

	const toggleEnable = () => {
		const newIsEnabled = !isEnabled;
		if (!newIsEnabled) {
			const enabledProviders = Object.keys(aiSettings).filter(
				k => aiSettings[k as ProviderName]?.isEnabled && k !== provider
			);
			console.log(`Remaining providers ${enabledProviders}`);
			if (enabledProviders.length === 0) {
				setShowModal(true);
				return;
			}
		}
		setIsEnabled(newIsEnabled);
		onChange('isEnabled', newIsEnabled);
		onSave('isEnabled', newIsEnabled);
	};

	return (
		<div className="bg-base-100 rounded-lg shadow-lg p-4 mb-4">
			<div className="flex justify-between items-center">
				<div className="flex items-center space-x-4">
					<label className="text-sm font-medium">Enable</label>
					<input
						type="checkbox"
						checked={isEnabled}
						onChange={toggleEnable}
						className="toggle toggle-primary rounded-full"
					/>
				</div>
				<h3 className="text-xl font-semibold capitalize cursor-pointer" onClick={toggleExpand}>
					{provider}
				</h3>
				<div className="flex items-center space-x-4 cursor-pointer" onClick={toggleExpand}>
					<div className="flex items-center space-x-1">
						<span className="text-sm">API Key</span>
						{settings.apiKey ? (
							<FiCheckCircle className="text-green-500" title="API Key Configured" />
						) : (
							<FiXCircle className="text-red-500" title="API Key Not Configured" />
						)}
					</div>
					<div className="flex items-center space-x-1">
						<span className="text-sm">Full Settings</span>
						{isExpanded ? <FiChevronUp className="text-gray-500" /> : <FiChevronDown className="text-gray-500" />}
					</div>
				</div>
			</div>

			{isEnabled && isExpanded && (
				<div className="m-4 space-y-4">
					{/* API Key */}
					<div className="grid grid-cols-12 gap-4 items-center">
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
					<div className="grid grid-cols-12 gap-4 items-center">
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
					<div className="grid grid-cols-12 gap-4 items-center">
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
					<div className="grid grid-cols-12 gap-4 items-center">
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
			)}

			{showModal && (
				<dialog className="modal modal-open">
					<div className="modal-box w-5/6 max-w-4xl">
						<div className="flex flex-row items-center">
							<FiAlertTriangle size={24} />
							<p className="text-lg px-4">Cannot disable the last provider !!!</p>
						</div>
					</div>
					<form method="dialog" className="modal-backdrop w-full">
						<button onClick={() => setShowModal(false)}>OK</button>
					</form>
				</dialog>
			)}
		</div>
	);
};

export default AISettingsCard;
