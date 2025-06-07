import { useRef, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp } from 'react-icons/fi';

import { UseCloseDetails } from '@/lib/use_close_details';

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
	// The mapped dropdownItems (like modelSettings or aiSettings).
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
}

// A single reusable dropdown that can be used by passing the appropriate config.
const Dropdown = <K extends string>(props: DropdownProps<K>) => {
	const {
		dropdownItems,
		selectedKey,
		onChange,
		filterDisabled = true,
		title = 'Select an option',
		getDisplayName,
	} = props;

	const [isOpen, setIsOpen] = useState(false);
	const detailsRef = useRef<HTMLDetailsElement>(null);

	// Close the details dropdown if the user clicks outside.
	UseCloseDetails({
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
			onToggle={(event: React.SyntheticEvent<HTMLElement>) => {
				setIsOpen((event.currentTarget as HTMLDetailsElement).open);
			}}
		>
			<summary
				className="flex btn w-full text-left shadow-none rounded-xl border-neutral/20 bg-base-100 justify-between items-center px-4 py-2 cursor-pointer"
				title={title}
			>
				<span className="font-normal">{selectedKey ? getItemDisplayName(selectedKey) : 'Select an option'}</span>
				{isOpen ? <FiChevronUp size={16} /> : <FiChevronDown size={16} />}
			</summary>

			<ul tabIndex={0} className="dropdown-content menu rounded-xl w-full bg-base-200 z-10">
				{filteredKeys.map(key => (
					<li
						key={key}
						className="cursor-pointer rounded-xl"
						onClick={() => {
							handleSelection(key);
						}}
					>
						<a className="flex justify-between items-center p-2 m-1">
							<span>{getItemDisplayName(key)}</span>
							{key === selectedKey && <FiCheck />}
						</a>
					</li>
				))}
			</ul>
		</details>
	);
};

export default Dropdown;
