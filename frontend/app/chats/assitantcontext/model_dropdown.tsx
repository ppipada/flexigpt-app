import { forwardRef } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import type { ChatOption } from '@/apis/chatoption_helper';

type ModelDropdownProps = {
	selectedModel: ChatOption;
	setSelectedModel: React.Dispatch<React.SetStateAction<ChatOption>>;
	allOptions: ChatOption[];
	isOpen: boolean;
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
};

const ModelDropdown = forwardRef<HTMLDetailsElement, ModelDropdownProps>(
	({ selectedModel, setSelectedModel, allOptions, isOpen, setIsOpen }, detailsRef) => {
		/* helper to decide which item shows the ✓ icon */
		const isCurrent = (m: ChatOption) =>
			m.providerName === selectedModel.providerName && m.modelPresetID === selectedModel.modelPresetID;

		return (
			<details
				ref={detailsRef}
				className="dropdown dropdown-top dropdown-end w-full justify-center"
				open={isOpen}
				onToggle={e => {
					setIsOpen((e.currentTarget as HTMLDetailsElement).open);
				}}
			>
				<summary
					className="btn btn-xs w-full text-left text-nowrap text-neutral-custom shadow-none border-none overflow-hidden"
					title="Select Model"
				>
					<div className="flex items-center w-full min-w-0">
						<span className="text-xs text-center font-normal truncate min-w-0 flex-1">
							{selectedModel.modelDisplayName}
						</span>
						{isOpen ? (
							<FiChevronDown size={16} className="ml-1 md:ml-2 shrink-0" />
						) : (
							<FiChevronUp size={16} className="ml-1 md:ml-2 shrink-0" />
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
								if (detailsRef && typeof detailsRef !== 'function' && detailsRef.current) {
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
);

ModelDropdown.displayName = 'ModelDropdown';
export default ModelDropdown;
