import { ProviderName } from '@/models/aiprovidermodel';
import { FC, useState } from 'react';
import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

interface ProviderDropdownProps {
	aiSettings: Record<string, any>;
	defaultProvider: ProviderName;
	onProviderChange: (provider: ProviderName) => void;
}

const ProviderDropdown: FC<ProviderDropdownProps> = ({ aiSettings, defaultProvider, onProviderChange }) => {
	const [isOpen, setIsOpen] = useState(false);
	const [selectedProvider, setSelectedProvider] = useState(defaultProvider);

	const toggleDropdown = () => setIsOpen(prev => !prev);

	const handleSelection = (provider: ProviderName) => {
		setSelectedProvider(provider);
		onProviderChange(provider);
		setIsOpen(false);
	};

	const capitalize = (str: string) => str.charAt(0).toUpperCase() + str.slice(1);

	return (
		<div className="dropdown relative w-full">
			<label
				tabIndex={0}
				className="flex btn w-full text-left shadow-none border-none rounded-box bg-base-100 justify-between items-center px-1"
				onClick={toggleDropdown}
				title="Select Provider"
			>
				<h3 className="text-sm font-medium capitalize">{selectedProvider}</h3>
				{isOpen ? <FiChevronUp /> : <FiChevronDown />}
			</label>
			<ul
				tabIndex={0}
				className={`dropdown-content menu rounded-box w-full bg-base-200 ${isOpen ? 'block' : 'hidden'}`}
				onClick={() => setIsOpen(false)}
			>
				{Object.keys(aiSettings)
					.filter(provider => aiSettings[provider]?.isEnabled)
					.map(provider => (
						<li
							key={provider}
							className="cursor-pointer rounded-box"
							onClick={() => handleSelection(provider as ProviderName)}
						>
							<a className="justify-between items-center p-2 m-1 flex">
								<span>{capitalize(provider)}</span>
								{provider === selectedProvider && <FiCheck />}
							</a>
						</li>
					))}
			</ul>
		</div>
	);
};

export default ProviderDropdown;
