import type { ProviderName } from '@/models/aiprovidermodel';
import { ProviderInfoDescription } from '@/models/aiprovidermodel';
import type { AISetting } from '@/models/settingmodel';
import type { FC } from 'react';
import { useState } from 'react';
import { FiAlertTriangle, FiCheckCircle, FiChevronDown, FiChevronUp, FiXCircle } from 'react-icons/fi';

interface AISettingsCardProps {
	provider: ProviderName;
	settings: AISetting;
	onChange: (key: string, value: any) => void;
	onSave: (key: string, value: any) => Promise<void>;
	aiSettings: Record<string, AISetting>;
}

const AISettingsCard: FC<AISettingsCardProps> = ({ provider, settings, onChange, onSave, aiSettings }) => {
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
			const enabledProviders = Object.keys(aiSettings).filter(k => aiSettings[k].isEnabled && k !== provider);
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
			<div className="grid grid-cols-12 gap-4 items-center">
				{/* Provider Title*/}
				<div className="col-span-3 flex items-center space-x-4">
					<h3 className="text-sm font-medium capitalize">{provider}</h3>
				</div>
				{/* Enable/Disable Toggle */}
				<div className="col-span-3 flex items-center space-x-4 ml-1">
					<label className="text-sm font-medium">Enable</label>
					<input
						type="checkbox"
						checked={isEnabled}
						onChange={toggleEnable}
						className="toggle toggle-primary rounded-full"
						spellCheck="false"
					/>
				</div>
				{/* Full Settings with Chevron */}
				<div className="col-span-6 cursor-pointer space-x-4 flex items-end justify-end" onClick={toggleExpand}>
					<div className="flex items-center">
						<span className="text-sm font-medium">API Key</span>
						{settings.apiKey ? (
							<FiCheckCircle className="text-green-500 mx-1" title="API Key Configured" />
						) : (
							<FiXCircle className="text-red-500 mx-1" title="API Key Not Configured" />
						)}
					</div>
					<div className="flex items-center">
						<span className="text-sm font-medium">Full Settings</span>
						{isExpanded ? (
							<FiChevronUp size={16} className="mx-1 text-gray-500" />
						) : (
							<FiChevronDown size={16} className="mx-1 text-gray-500" />
						)}
					</div>
				</div>
			</div>

			{isEnabled && isExpanded && (
				<div className="m-1 mt-8 space-y-4">
					{/* API Key */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm text-left tooltip" data-tip={ProviderInfoDescription['apiKey']}>
							API Key
						</label>
						<input
							type="password"
							className="input col-span-9 w-full h-10 rounded-lg px-4 py-2"
							style={{ fontSize: '14px' }}
							value={settings.apiKey}
							onChange={e => {
								onChange('apiKey', e.target.value);
							}}
							onBlur={e => {
								onSave('apiKey', e.target.value);
							}}
							spellCheck="false"
						/>
					</div>

					{/* Origin */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm text-left tooltip" data-tip={ProviderInfoDescription['origin']}>
							Origin
						</label>
						<input
							type="text"
							className="input col-span-9 w-full h-10 rounded-lg px-4 py-2"
							style={{ fontSize: '14px' }}
							value={settings.origin}
							onChange={e => {
								onChange('origin', e.target.value);
							}}
							onBlur={e => {
								onSave('origin', e.target.value);
							}}
							spellCheck="false"
						/>
					</div>

					{/* Default Model */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label className="col-span-3 text-sm text-left tooltip" data-tip={ProviderInfoDescription['defaultModel']}>
							Default Model
						</label>
						<input
							type="text"
							className="input col-span-9 w-full h-10 rounded-lg px-4 py-2"
							style={{ fontSize: '14px' }}
							value={settings.defaultModel}
							onChange={e => {
								onChange('defaultModel', e.target.value);
							}}
							onBlur={e => {
								onSave('defaultModel', e.target.value);
							}}
							spellCheck="false"
						/>
					</div>

					{/* Default Temperature */}
					<div className="grid grid-cols-12 gap-4 items-center">
						<label
							className="col-span-3 text-sm text-left tooltip"
							data-tip={ProviderInfoDescription['defaultTemperature']}
						>
							Default Temperature
						</label>
						<input
							type="number"
							step="0.01"
							min="0"
							max="1"
							className="input col-span-9 w-full h-10 rounded-lg px-4 py-2"
							style={{ fontSize: '14px' }}
							value={settings.defaultTemperature}
							onChange={e => {
								onChange('defaultTemperature', parseFloat(e.target.value));
							}}
							onBlur={e => {
								onSave('defaultTemperature', parseFloat(e.target.value));
							}}
							spellCheck="false"
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
						<button
							onClick={() => {
								setShowModal(false);
							}}
						>
							OK
						</button>
					</form>
				</dialog>
			)}
		</div>
	);
};

export default AISettingsCard;
