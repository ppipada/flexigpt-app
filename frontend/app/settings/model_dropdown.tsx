import type { FC } from 'react';
import { useRef, useState } from 'react';
import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import { UseCloseDetails } from '@/lib/use_close_details';

import type { ModelName } from '@/models/aiprovidermodel';
import type { ModelSetting } from '@/models/settingmodel';

interface ModelDropdownProps {
	modelSettings: Record<ModelName, ModelSetting>;
	defaultModel: string;
	onModelChange: (modelName: string) => void;
}

const ModelDropdown: FC<ModelDropdownProps> = ({ modelSettings, defaultModel, onModelChange }) => {
	const [isOpen, setIsOpen] = useState(false);
	const [selectedModel, setSelectedModel] = useState(defaultModel);
	const detailsRef = useRef<HTMLDetailsElement>(null);

	UseCloseDetails({
		detailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsOpen(false);
		},
	});

	const handleSelection = (modelName: string) => {
		setSelectedModel(modelName);
		onModelChange(modelName);
		// Force-close the native dropdown
		if (detailsRef.current) {
			detailsRef.current.open = false;
		}
		setIsOpen(false);
	};

	return (
		<details
			ref={detailsRef}
			className="dropdown relative w-full"
			onToggle={(event: React.SyntheticEvent<HTMLElement>) => {
				setIsOpen((event.currentTarget as HTMLDetailsElement).open);
			}}
		>
			<summary
				className="flex btn w-full text-left shadow-none rounded-xl border border-base-300 bg-base-100 justify-between items-center px-4 py-2 cursor-pointer"
				title="Select Model"
			>
				<span className="font-normal">
					{selectedModel !== ''
						? selectedModel in modelSettings
							? modelSettings[selectedModel].displayName
							: selectedModel
						: 'Select a model'}
				</span>
				{isOpen ? <FiChevronUp size={16} /> : <FiChevronDown size={16} />}
			</summary>
			<ul tabIndex={0} className="dropdown-content menu rounded-xl w-full bg-base-200 z-10">
				{Object.entries(modelSettings).map(([modelName, model]) => (
					<li
						key={modelName}
						className="cursor-pointer rounded-xl"
						onClick={() => {
							handleSelection(modelName);
						}}
					>
						<a className="flex justify-between items-center p-2 m-1">
							<span>{model.displayName || modelName}</span>
							{modelName === selectedModel && <FiCheck />}
						</a>
					</li>
				))}
			</ul>
		</details>
	);
};

export default ModelDropdown;
