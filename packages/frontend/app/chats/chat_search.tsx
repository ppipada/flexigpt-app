import { ConversationItem } from 'conversationmodel';
// import { log } from 'logger';
import { ChangeEvent, FC, Fragment, KeyboardEvent, useEffect, useState } from 'react';
import { FiSearch } from 'react-icons/fi';

interface ChatSearchProps {
	initialItems: ConversationItem[];
	onSearch: (query: string) => Promise<ConversationItem[]>;
	onSelectConversation: (item: ConversationItem) => Promise<void>;
}

export const ChatSearch: FC<ChatSearchProps> = ({ initialItems, onSearch, onSelectConversation }) => {
	const [query, setQuery] = useState('');
	const [items, setItems] = useState<ConversationItem[]>([]);
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

	const handleItemClick = (item: ConversationItem) => {
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
			{showDropdown && <GroupedDropdown items={items} handleItemClick={handleItemClick} focusedIndex={focusedIndex} />}
		</div>
	);
};

interface GroupedDropdownProps {
	items: ConversationItem[];
	handleItemClick: (item: ConversationItem) => void;
	focusedIndex: number;
}

const GroupedDropdown: FC<GroupedDropdownProps> = ({ items, handleItemClick, focusedIndex }) => {
	// Function to group items by date
	const groupItems = (items: ConversationItem[]) => {
		const now = new Date();
		const yesterday = new Date(now);
		yesterday.setDate(now.getDate() - 1);
		const last7Days = new Date(now);
		last7Days.setDate(now.getDate() - 7);

		const groups: { [key: string]: ConversationItem[] } = {
			Yesterday: [],
			'Last 7 Days': [],
			Older: [],
		};

		items.forEach(item => {
			if (item.createdAt >= yesterday && item.createdAt < now) {
				groups['Yesterday'].push(item);
			} else if (item.createdAt >= last7Days && item.createdAt < now) {
				groups['Last 7 Days'].push(item);
			} else {
				groups['Older'].push(item);
			}
		});

		return groups;
	};

	const groupedItems = groupItems(items);
	// log.info('Grouped Items:', JSON.stringify(groupedItems, null, 2));
	return (
		<ul
			className="absolute left-0 right-0 mt-0 max-h-80 overflow-y-auto bg-base-200 rounded-2xl shadow-lg z-10"
			style={{ fontSize: '14px' }}
		>
			{Object.entries(groupedItems).map(([groupTitle, groupItems]) => (
				<Fragment key={groupTitle}>
					<li className="text-neutral-400 text-sm px-12 py-2">{groupTitle}</li>
					{groupItems.map((item, index) => (
						<li
							key={item.id}
							onClick={() => handleItemClick(item)}
							className={`flex justify-between items-center px-12 py-2 cursor-pointer hover:bg-base-100 ${index === focusedIndex ? 'bg-base-100' : ''}`}
						>
							<span>{item.title}</span>
							<span className="text-neutral-400 text-xs hidden lg:block">
								{new Date(item.createdAt).getDate()}{' '}
								{new Intl.DateTimeFormat('en-US', { month: 'short' }).format(new Date(item.createdAt))}{' '}
								{new Date(item.createdAt).getFullYear()}
							</span>
						</li>
					))}
				</Fragment>
			))}
		</ul>
	);
};
