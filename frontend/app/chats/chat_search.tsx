import type { ChangeEvent, FC, KeyboardEvent } from 'react';
import { Fragment, useEffect, useState } from 'react';

import { FiSearch } from 'react-icons/fi';

import type { ConversationItem } from '@/models/conversationmodel';

function groupItems(items: ConversationItem[]) {
	const normalize = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate());
	const today = normalize(new Date());
	const yesterday = new Date(today);
	yesterday.setDate(today.getDate() - 1);
	const last7Days = new Date(today);
	last7Days.setDate(today.getDate() - 7);

	const buckets: Record<string, ConversationItem[]> = {
		Today: [],
		Yesterday: [],
		'Last 7 Days': [],
		Older: [],
	};

	items.forEach(it => {
		const day = normalize(new Date(it.createdAt));
		if (day >= today) buckets['Today'].push(it);
		else if (day >= yesterday) buckets['Yesterday'].push(it);
		else if (day >= last7Days) buckets['Last 7 Days'].push(it);
		else buckets['Older'].push(it);
	});

	return buckets;
}

interface GroupedDropdownProps {
	items: ConversationItem[];
	onClick: (item: ConversationItem) => void;
	focusedIndex: number;
}
const GroupedDropdown: FC<GroupedDropdownProps> = ({ items, onClick, focusedIndex }) => {
	const grouped = groupItems(items);

	return (
		<ul className="absolute left-0 right-0 mt-0 max-h-80 overflow-y-auto bg-base-200 rounded-2xl shadow-lg text-sm">
			{Object.entries(grouped)
				.filter(([, arr]) => arr.length)
				.map(([title, arr]) => (
					<Fragment key={title}>
						<li className="px-12 py-2 text-neutral/60">{title}</li>
						{arr.map((it, i) => (
							<li
								key={it.id}
								onClick={() => {
									onClick(it);
								}}
								className={`flex justify-between items-center px-12 py-2 cursor-pointer hover:bg-base-100 ${
									i === focusedIndex ? 'bg-base-100' : ''
								}`}
							>
								<span className="truncate">{it.title}</span>
								<span className="hidden lg:block text-neutral text-xs">
									{new Date(it.createdAt).toLocaleDateString('en-US', {
										day: '2-digit',
										month: 'short',
										year: 'numeric',
									})}
								</span>
							</li>
						))}
					</Fragment>
				))}
		</ul>
	);
};

interface ChatSearchProps {
	initialItems: ConversationItem[];
	onSearch: (query: string) => Promise<ConversationItem[]>;
	onSelectConversation: (item: ConversationItem) => Promise<void>;
}

const ChatSearch: FC<ChatSearchProps> = ({ initialItems, onSearch, onSelectConversation }) => {
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
		// setQuery(item.title);
		setQuery('');
		setShowDropdown(false);
	};

	const handleFocus = () => {
		setShowDropdown(true);
		setItems(initialItems); // Show all items initially when focused
	};

	const handleBlur = () => {
		// Delay hiding the dropdown to allow item click event to be processed
		setTimeout(() => {
			setShowDropdown(false);
		}, 100);
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
				<FiSearch size={24} className="mx-2 text-neutral/60" />
				<input
					type="text"
					value={query}
					onChange={handleSearchChange}
					onFocus={handleFocus}
					onBlur={handleBlur}
					onKeyDown={handleKeyDown}
					placeholder="Previous chats..."
					className="w-full bg-base-100 border-none outline-hidden"
					style={{ fontSize: '14px' }}
					spellCheck="false"
				/>
			</div>
			{showDropdown && <GroupedDropdown items={items} onClick={handleItemClick} focusedIndex={focusedIndex} />}
		</div>
	);
};
export default ChatSearch;
