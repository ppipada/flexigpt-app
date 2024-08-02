import { ChangeEvent, FC, KeyboardEvent, useEffect, useState } from 'react';
import { FiSearch } from 'react-icons/fi';

// Define the type for search items
export type SearchItem = {
	id: string;
	title: string;
};

interface ChatSearchProps {
	initialItems: SearchItem[];
	onSearch: (query: string) => Promise<SearchItem[]>;
	onSelectConversation: (item: SearchItem) => Promise<void>;
}

export const ChatSearch: FC<ChatSearchProps> = ({ initialItems, onSearch, onSelectConversation }) => {
	const [query, setQuery] = useState('');
	const [items, setItems] = useState<SearchItem[]>([]);
	const [showDropdown, setShowDropdown] = useState(false);
	const [focusedIndex, setFocusedIndex] = useState(-1);

	useEffect(() => {
		setItems(initialItems);
	}, [initialItems]);

	const handleSearchChange = async (e: ChangeEvent<HTMLInputElement>) => {
		const value = e.target.value;
		setQuery(value);

		if (value) {
			const results = await onSearch(value);
			setItems(results);
			setFocusedIndex(-1); // Reset focus index when search results change
		} else {
			setItems(initialItems);
		}
	};

	const handleItemClick = (item: SearchItem) => {
		onSelectConversation(item);
		setQuery(item.title);
		setShowDropdown(false);
	};

	const handleFocus = () => {
		setShowDropdown(true);
		setItems(initialItems); // Show all items initially when focused
	};

	const handleBlur = () => {
		// Delay hiding the dropdown to allow item click event to be processed
		setTimeout(() => setShowDropdown(false), 100);
	};

	const handleKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
		if (e.key === 'ArrowDown') {
			setFocusedIndex(prevIndex => (prevIndex + 1) % items.length);
		} else if (e.key === 'ArrowUp') {
			setFocusedIndex(prevIndex => (prevIndex - 1 + items.length) % items.length);
		} else if (e.key === 'Escape') {
			setFocusedIndex(-1);
		} else if (e.key === 'Enter') {
			if (focusedIndex >= 0 && focusedIndex < items.length) {
				handleItemClick(items[focusedIndex]);
			}
		}
	};

	return (
		<div className="relative">
			<div className="flex items-center bg-base-100 py-3 mx-0 rounded-2xl border">
				<FiSearch size={24} className="mx-2 text-neutral-400" />
				<input
					type="text"
					value={query}
					onChange={handleSearchChange}
					onFocus={handleFocus}
					onBlur={handleBlur}
					onKeyDown={handleKeyDown}
					placeholder="Search or load chats"
					className="w-full bg-base-100 border-none outline-none"
					style={{ fontSize: '14px' }}
				/>
			</div>
			{showDropdown && (
				<ul
					className="absolute left-0 right-0 mt-0 max-h-90 overflow-y-auto bg-base-200 rounded-2xl shadow-lg z-10"
					style={{ fontSize: '14px' }}
				>
					{items.map((item, index) => (
						<li
							key={item.id}
							onClick={() => handleItemClick(item)}
							className={`px-12 py-2 cursor-pointer hover:bg-base-100 ${index === focusedIndex ? 'bg-base-100' : ''}`}
						>
							{item.title}
						</li>
					))}
				</ul>
			)}
		</div>
	);
};
