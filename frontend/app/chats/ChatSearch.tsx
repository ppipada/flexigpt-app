import React, { useEffect, useState } from 'react';
import { FiSearch } from 'react-icons/fi';

interface ChatSearchProps {
	initialItems: string[];
	onSearch: (query: string) => Promise<string[]>;
}

const ChatSearch: React.FC<ChatSearchProps> = ({ initialItems, onSearch }) => {
	const [query, setQuery] = useState('');
	const [items, setItems] = useState<string[]>([]);
	const [showDropdown, setShowDropdown] = useState(false);

	useEffect(() => {
		setItems(initialItems);
	}, [initialItems]);

	const handleSearchChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
		const value = e.target.value;
		setQuery(value);

		if (value) {
			const results = await onSearch(value);
			setItems(results);
		} else {
			setItems(initialItems);
		}
	};

	const handleItemClick = (item: string) => {
		setQuery(item);
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
					placeholder="Search or load chats"
					className="w-full bg-base-100 border-none outline-none"
					style={{ fontSize: '14px' }}
				/>
			</div>
			{showDropdown && (
				<ul
					className="absolute left-0 right-0 mt-0 max-h-60 overflow-y-auto bg-base-200 rounded-2xl shadow-lg z-10"
					style={{ fontSize: '14px' }}
				>
					{items.map((item, index) => (
						<li
							key={index}
							onClick={() => handleItemClick(item)}
							className="px-12 py-2 cursor-pointer hover:bg-base-100"
						>
							{item}
						</li>
					))}
				</ul>
			)}
		</div>
	);
};

export default ChatSearch;
