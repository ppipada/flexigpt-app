import { type Dispatch, forwardRef, type SetStateAction, type SyntheticEvent } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import { ReasoningLevel } from '@/spec/modelpreset';

type SingleReasoningDropdownProps = {
	reasoningLevel: ReasoningLevel;
	setReasoningLevel: (level: ReasoningLevel) => void;
	isOpen: boolean;
	setIsOpen: Dispatch<SetStateAction<boolean>>;
};

export const SingleReasoningDropdown = forwardRef<HTMLDetailsElement, SingleReasoningDropdownProps>(
	function SingleReasoningDropdown({ reasoningLevel, setReasoningLevel, isOpen, setIsOpen }, detailsRef) {
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
					className="dropdown dropdown-top dropdown-end w-full justify-center"
					onToggle={(event: SyntheticEvent<HTMLElement>) => {
						setIsOpen((event.currentTarget as HTMLDetailsElement).open);
					}}
					open={isOpen}
				>
					<summary
						className="btn btn-xs text-neutral-custom w-full flex-1 items-center overflow-hidden border-none text-center text-nowrap shadow-none"
						title="Set Reasoning Level"
					>
						<span className="min-w-0 truncate text-center text-xs font-normal">
							Reasoning Level: {levelDisplayNames[reasoningLevel]}
						</span>
						{isOpen ? (
							<FiChevronDown size={16} className="ml-1 shrink-0 md:ml-2" />
						) : (
							<FiChevronUp size={16} className="ml-1 shrink-0 md:ml-2" />
						)}
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
