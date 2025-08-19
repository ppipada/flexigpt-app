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
			<div className="flex w-full justify-center">
				<details
					ref={detailsRef}
					className="dropdown dropdown-top dropdown-end w-full justify-center"
					open={isOpen}
					onToggle={e => {
						setIsOpen((e.currentTarget as HTMLDetailsElement).open);
					}}
				>
					<summary
						className="btn btn-xs text-neutral-custom w-full flex-1 items-center overflow-hidden border-none text-center text-nowrap shadow-none"
						title="Select Model"
					>
						<span className="min-w-0 truncate text-center text-xs font-normal">{selectedModel.modelDisplayName}</span>
						{isOpen ? (
							<FiChevronDown size={16} className="ml-1 shrink-0 md:ml-2" />
						) : (
							<FiChevronUp size={16} className="ml-1 shrink-0 md:ml-2" />
						)}
					</summary>

					<ul className="dropdown-content menu bg-base-100 w-full rounded-xl">
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
								<a className="m-0 flex items-center justify-between p-1">
									<span>{model.modelDisplayName}</span>
									{isCurrent(model) && <FiCheck />}
								</a>
							</li>
						))}
					</ul>
				</details>
			</div>
		);
	}
);

ModelDropdown.displayName = 'ModelDropdown';
export default ModelDropdown;
