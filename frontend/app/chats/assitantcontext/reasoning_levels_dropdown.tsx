import { forwardRef } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import { ReasoningLevel } from '@/spec/modelpreset';

type SingleReasoningDropdownProps = {
	reasoningLevel: ReasoningLevel;
	setReasoningLevel: (level: ReasoningLevel) => void;
	isOpen: boolean;
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
};

const SingleReasoningDropdown = forwardRef<HTMLDetailsElement, SingleReasoningDropdownProps>(
	({ reasoningLevel, setReasoningLevel, isOpen, setIsOpen }, detailsRef) => {
		// Map reasoning levels to display names.
		const levelDisplayNames = {
			[ReasoningLevel.Low]: 'Low',
			[ReasoningLevel.Medium]: 'Medium',
			[ReasoningLevel.High]: 'High',
		};

		return (
			<div className="flex w-full justify-center">
				<details
					ref={detailsRef}
					className="dropdown dropdown-top dropdown-end"
					onToggle={(event: React.SyntheticEvent<HTMLElement>) => {
						setIsOpen((event.currentTarget as HTMLDetailsElement).open);
					}}
					open={isOpen}
				>
					<summary
						className="btn btn-xs text-neutral-custom overflow-hidden border-none text-left text-nowrap shadow-none"
						title="Set Reasoning Level"
					>
						<div className="flex">
							<span className="mr-2 text-xs font-normal sm:hidden">Reasoning: </span>
							<span className="mr-2 hidden text-xs font-normal sm:inline">Reasoning Level: </span>
							<span className="text-xs font-normal"> {levelDisplayNames[reasoningLevel]} </span>
							{isOpen ? (
								<FiChevronDown size={16} className="ml-1 md:ml-2" />
							) : (
								<FiChevronUp size={16} className="ml-1 md:ml-2" />
							)}
						</div>
					</summary>

					<ul className="dropdown-content menu bg-base-100 w-full rounded-xl p-4">
						{/* Reasoning level options */}
						{Object.values(ReasoningLevel).map(level => (
							<li
								key={level}
								className="cursor-pointer text-xs"
								onClick={() => {
									setReasoningLevel(level);
									if (detailsRef && typeof detailsRef !== 'function' && detailsRef.current) {
										detailsRef.current.open = false;
									}
									setIsOpen(false);
								}}
							>
								<a className="m-0 items-center justify-between p-1">
									<span>{levelDisplayNames[level]}</span>
									{reasoningLevel === level && <FiCheck />}
								</a>
							</li>
						))}
					</ul>
				</details>
			</div>
		);
	}
);

SingleReasoningDropdown.displayName = 'SingleReasoningDropdown';
export default SingleReasoningDropdown;
