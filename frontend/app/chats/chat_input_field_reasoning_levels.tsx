import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import { ReasoningLevel } from '@/models/aiprovidermodel';

/**
 * Subcomponent for Reasoning Level Selection
 * -------------------------------------
 * Displays three reasoning levels (low, medium, high) the user can select.
 */
export default function SingleReasoningDropdown(props: {
	reasoningLevel: ReasoningLevel;
	setReasoningLevel: (level: ReasoningLevel) => void;
	isOpen: boolean;
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
	detailsRef: React.RefObject<HTMLDetailsElement | null>;
}) {
	const { reasoningLevel, setReasoningLevel, isOpen, setIsOpen, detailsRef } = props;

	// Map reasoning levels to display names
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
					className="btn btn-xs text-left text-nowrap text-neutral/60 shadow-none border-none overflow-hidden"
					title="Set Reasoning Level"
				>
					<div className="flex">
						<span className="text-xs font-normal sm:hidden mr-2">Reasoning: </span>
						<span className="text-xs font-normal hidden sm:inline mr-2">Reasoning Level: </span>
						<span className="text-xs font-normal"> {levelDisplayNames[reasoningLevel]} </span>
						{isOpen ? (
							<FiChevronDown size={16} className="ml-1 md:ml-2" />
						) : (
							<FiChevronUp size={16} className="ml-1 md:ml-2" />
						)}
					</div>
				</summary>

				<ul className="dropdown-content menu bg-base-100 rounded-xl w-full p-4">
					{/* Reasoning level options */}
					{Object.values(ReasoningLevel).map(level => (
						<li
							key={level}
							className="cursor-pointer text-xs"
							onClick={() => {
								setReasoningLevel(level);
								if (detailsRef.current) {
									detailsRef.current.open = false;
								}
								setIsOpen(false);
							}}
						>
							<a className="justify-between items-center p-1 m-0">
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
