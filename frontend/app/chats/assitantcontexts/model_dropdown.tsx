import type { Dispatch, SetStateAction } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import { Select, SelectItem, SelectPopover, useSelectStore, useStoreState } from '@ariakit/react';

import type { ChatOption } from '@/chats/chat_option_helper';

const modelKey = (m: ChatOption) => `${m.providerName}::${m.modelPresetID}`;

type ModelDropdownProps = {
	selectedModel: ChatOption;
	setSelectedModel: Dispatch<SetStateAction<ChatOption>>;
	allOptions: ChatOption[];
	isOpen: boolean;
	setIsOpen: Dispatch<SetStateAction<boolean>>;
};

export function ModelDropdown({ selectedModel, setSelectedModel, allOptions, isOpen, setIsOpen }: ModelDropdownProps) {
	const currentKey = modelKey(selectedModel);

	const select = useSelectStore({
		// value is a string key, not the full ChatOption
		value: currentKey,
		setValue: key => {
			if (typeof key !== 'string') return;
			const model = allOptions.find(m => modelKey(m) === key);
			if (model) setSelectedModel(model);
		},

		// external open state, so parent can still control it
		open: isOpen,
		setOpen: setIsOpen,

		placement: 'top-start', // open above the trigger

		focusLoop: true, // circular keyboard navigation
	});

	const open = useStoreState(select, 'open');

	const isCurrent = (m: ChatOption) => modelKey(m) === currentKey;

	return (
		<div className="flex w-full justify-center">
			{/* This wrapper is just for layout; no Daisy dropdown widget here */}
			<div className="relative w-full">
				{/* Trigger button */}
				<Select
					store={select}
					className="btn btn-xs text-neutral-custom w-full flex-1 items-center overflow-hidden border-none text-center text-nowrap shadow-none"
					title="Select Model"
				>
					<span className="min-w-0 truncate text-center text-xs font-normal">{selectedModel.modelDisplayName}</span>
					{/* Chevron UP when closed, DOWN when open */}
					{open ? (
						<FiChevronDown size={16} className="ml-1 shrink-0 md:ml-2" />
					) : (
						<FiChevronUp size={16} className="ml-1 shrink-0 md:ml-2" />
					)}
				</Select>

				{/* Popover with DaisyUI-like styling, managed by Ariakit */}
				<SelectPopover
					store={select}
					portal={false}
					gutter={4}
					autoFocusOnShow
					sameWidth
					className="border-base-300 bg-base-100 z-50 mt-1 max-h-72 w-full overflow-y-auto rounded-xl border p-1 text-xs shadow-lg outline-none"
				>
					{allOptions.map(model => (
						<SelectItem
							key={modelKey(model)}
							value={modelKey(model)}
							className="hover:bg-base-200 data-active-item:bg-base-300 m-0 flex cursor-pointer items-center justify-between rounded-md px-2 py-1 transition-colors outline-none"
						>
							<span>{model.modelDisplayName}</span>
							{isCurrent(model) && <FiCheck />}
						</SelectItem>
					))}
				</SelectPopover>
			</div>
		</div>
	);
}
