import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import type { ChatOption } from '@/apis/chatoption_helper';

interface ModelDropdownProps {
	selectedModel: ChatOption;
	setSelectedModel: React.Dispatch<React.SetStateAction<ChatOption>>;
	allOptions: ChatOption[];
	isOpen: boolean;
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
	detailsRef: React.RefObject<HTMLDetailsElement | null>;
}

export default function ModelDropdown(props: ModelDropdownProps) {
	const { selectedModel, setSelectedModel, allOptions, isOpen, setIsOpen, detailsRef } = props;

	/* helper to decide which item shows the ✓ icon */
	const isCurrent = (m: ChatOption) =>
		m.providerName === selectedModel.providerName && m.modelPresetID === selectedModel.modelPresetID;

	return (
		<details
			ref={detailsRef}
			className="dropdown dropdown-top dropdown-end w-full"
			open={isOpen}
			onToggle={e => {
				setIsOpen((e.currentTarget as HTMLDetailsElement).open);
			}}
		>
			<summary
				className="btn btn-xs w-full text-left text-nowrap text-neutral/60 shadow-none border-none overflow-hidden"
				title="Select Model"
			>
				<div className="flex items-center">
					{/* truncate on very small screens */}
					<span className="text-xs font-normal sm:hidden">{selectedModel.modelDisplayName.slice(0, 8)}</span>
					<span className="text-xs font-normal hidden sm:inline">{selectedModel.modelDisplayName}</span>

					{isOpen ? (
						<FiChevronDown size={16} className="ml-1 md:ml-2" />
					) : (
						<FiChevronUp size={16} className="ml-1 md:ml-2" />
					)}
				</div>
			</summary>

			<ul className="dropdown-content menu bg-base-100 rounded-xl w-full">
				{allOptions.map(model => (
					<li
						key={`${model.providerName}-${model.modelPresetID}`}
						className="cursor-pointer text-xs"
						onClick={() => {
							setSelectedModel(model);
							/* close the <details> dropdown manually */
							if (detailsRef.current) {
								detailsRef.current.open = false;
							}
							setIsOpen(false);
						}}
					>
						<a className="flex justify-between items-center p-1 m-0">
							<span>{model.modelDisplayName}</span>
							{isCurrent(model) && <FiCheck />}
						</a>
					</li>
				))}
			</ul>
		</details>
	);
}
