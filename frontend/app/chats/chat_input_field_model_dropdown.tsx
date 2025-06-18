import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import type { ChatOptions } from '@/models/modelpresetsmodel';

export default function ModelDropdown(props: {
	selectedModel: ChatOptions;
	setSelectedModel: React.Dispatch<React.SetStateAction<ChatOptions>>;
	allOptions: ChatOptions[];
	isOpen: boolean;
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
	detailsRef: React.RefObject<HTMLDetailsElement | null>;
}) {
	const { selectedModel, setSelectedModel, allOptions, isOpen, setIsOpen, detailsRef } = props;

	return (
		<details
			ref={detailsRef}
			className="dropdown dropdown-top dropdown-end w-full"
			onToggle={(event: React.SyntheticEvent<HTMLElement>) => {
				setIsOpen((event.currentTarget as HTMLDetailsElement).open);
			}}
			open={isOpen}
		>
			<summary
				className="btn btn-xs w-full text-left text-nowrap text-neutral/60 shadow-none border-none overflow-hidden"
				title="Select Model"
			>
				<div className="flex">
					<span className="text-xs font-normal sm:hidden">{selectedModel.title.substring(0, 8)}</span>
					<span className="text-xs font-normal hidden sm:inline">{selectedModel.title} </span>
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
						key={`${model.provider}-${model.id}`}
						className="cursor-pointer text-xs"
						onClick={() => {
							setSelectedModel(model);
							if (detailsRef.current) {
								detailsRef.current.open = false;
							}
							setIsOpen(false);
						}}
					>
						<a className="justify-between items-center p-1 m-0">
							<span>{model.title}</span>
							{selectedModel.name === model.name && selectedModel.provider === model.provider && <FiCheck />}
						</a>
					</li>
				))}
			</ul>
		</details>
	);
}
