import { type SyntheticEvent, useRef, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import { useCloseDetails } from '@/hooks/use_close_details';

/**
 * @public
 */
export interface DropdownItem {
	isEnabled: boolean;
}

/**
 * @public
 */
export interface DropdownProps<K extends string> {
	// The mapped dropdownItems (like modelPresets or aiSettings).
	dropdownItems: Record<K, DropdownItem>;
	//  The currently selected key (e.g. 'gpt-3.5' or 'openai').
	selectedKey: K;
	// Called when user picks a new key.
	onChange: (key: K) => void;
	// Optional. If you want to filter out items that
	// are not enabled unless they are the current selection,
	// set this to true. (Defaults to true).
	filterDisabled?: boolean;
	// Optional text that appears in a tooltip or as a title
	// on the summary element.
	title?: string;
	// Optional callback to get the display name for an item in the dropdown.
	// If the item itself has a `displayName` property, that takes precedence.
	// Otherwise, this function (if present) will be used to determine the label.
	getDisplayName?: (key: K) => string;
	maxMenuHeight?: number | string; // Optional, default 300
}

// A single reusable dropdown that can be used by passing the appropriate config.
export const Dropdown = <K extends string>(props: DropdownProps<K>) => {
	const {
		dropdownItems,
		selectedKey,
		onChange,
		filterDisabled = true,
		title = 'Select an option',
		getDisplayName,
		maxMenuHeight = 300,
	} = props;

	const [isOpen, setIsOpen] = useState(false);
	const detailsRef = useRef<HTMLDetailsElement>(null);

	// Close the details dropdown if the user clicks outside.
	useCloseDetails({
		detailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsOpen(false);
		},
	});

	const handleSelection = (key: K) => {
		onChange(key);
		if (detailsRef.current) {
			// Force-close the details dropdown.
			detailsRef.current.open = false;
		}
		setIsOpen(false);
	};

	// Helper function to derive display name:
	// 1. Use the consumer's getDisplayName callback if provided
	// 2. Otherwise, fall back to the key itself.
	const getItemDisplayName = (key: K) => {
		if (typeof getDisplayName === 'function') {
			return getDisplayName(key);
		}
		return key;
	};

	// We can optionally filter out disabled items, unless they are already selected.
	const filteredKeys = Object.keys(dropdownItems).filter(k => {
		const typedKey = k as K;
		const item = dropdownItems[typedKey];
		if (!filterDisabled) return true;
		// If the item is enabled or it is the selected key, we keep it.
		return item.isEnabled || typedKey === selectedKey;
	}) as K[];

	return (
		<details
			ref={detailsRef}
			className="dropdown relative w-full"
			onToggle={(event: SyntheticEvent<HTMLElement>) => {
				setIsOpen((event.currentTarget as HTMLDetailsElement).open);
			}}
		>
			<summary
				className="btn border-neutral/20 bg-base-200 flex w-full cursor-pointer items-center justify-between rounded-2xl px-4 py-2 text-left shadow-none"
				title={title}
			>
				<span className="truncate font-normal">
					{selectedKey ? getItemDisplayName(selectedKey) : 'Select an option'}
				</span>
				{isOpen ? <FiChevronUp size={16} /> : <FiChevronDown size={16} />}
			</summary>

			<ul
				tabIndex={0}
				className="dropdown-content menu border-neutral/20 bg-base-300 flex w-full flex-col flex-nowrap overflow-x-hidden overflow-y-auto rounded-2xl shadow-sm"
				style={{
					maxHeight: typeof maxMenuHeight === 'number' ? `${maxMenuHeight}px` : maxMenuHeight,
				}}
			>
				{filteredKeys.sort().map(key => (
					<li
						key={key}
						className="w-full cursor-pointer rounded-2xl"
						onClick={() => {
							handleSelection(key);
						}}
					>
						<a className="m-1 flex items-center justify-between p-2">
							<span className="truncate">{getItemDisplayName(key)}</span>
							{key === selectedKey && <FiCheck />}
						</a>
					</li>
				))}
			</ul>
		</details>
	);
};
