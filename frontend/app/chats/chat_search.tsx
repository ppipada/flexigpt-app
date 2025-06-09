import type { ChangeEvent, FC, KeyboardEvent } from 'react';
import { Fragment, useCallback, useEffect, useState } from 'react';

import { FiSearch } from 'react-icons/fi';

import type { ConversationItem } from '@/models/conversationmodel';

import { listAllConversations } from '@/apis/conversationstore_helper';

function groupItems(items: ConversationItem[]) {
	const norm = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate());
	const today = norm(new Date());
	const yesterday = new Date(today);
	yesterday.setDate(today.getDate() - 1);
	const last7 = new Date(today);
	last7.setDate(today.getDate() - 7);

	const buckets: Record<string, ConversationItem[]> = { Today: [], Yesterday: [], 'Last 7 Days': [], Older: [] };

	items.forEach(it => {
		const day = norm(new Date(it.createdAt));
		if (day >= today) buckets.Today.push(it);
		else if (day >= yesterday) buckets.Yesterday.push(it);
		else if (day >= last7) buckets['Last 7 Days'].push(it);
		else buckets.Older.push(it);
	});
	return buckets;
}

interface GroupedProps {
	items: ConversationItem[];
	onPick: (i: ConversationItem) => void;
	focused: number;
}
const GroupedDropdown: FC<GroupedProps> = ({ items, onPick, focused }) => (
	<ul className="absolute left-0 right-0 mt-0 max-h-80 overflow-y-auto bg-base-200 rounded-2xl shadow-lg text-sm">
		{Object.entries(groupItems(items))
			.filter(([, arr]) => arr.length)
			.map(([label, arr]) => (
				<Fragment key={label}>
					<li className="px-12 py-2 text-neutral/60">{label}</li>
					{arr.map((it, i) => (
						<li
							key={it.id}
							onClick={() => {
								onPick(it);
							}}
							className={`flex justify-between items-center px-12 py-2 cursor-pointer hover:bg-base-100 ${
								i === focused ? 'bg-base-100' : ''
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

interface ChatSearchProps {
	onSelectConversation: (item: ConversationItem) => Promise<void>;
	// Bump this value from the parent to force a reload.
	refreshKey: number;
}

const ChatSearch: FC<ChatSearchProps> = ({ onSelectConversation, refreshKey }) => {
	const [allItems, setAllItems] = useState<ConversationItem[]>([]);
	const [visible, setVisible] = useState<ConversationItem[]>([]);
	const [query, setQuery] = useState('');
	const [show, setShow] = useState(false);
	const [focusedIndex, setFocused] = useState(-1);

	// Fetch list (on mount AND when refreshKey changes).
	const fetchList = useCallback(async () => {
		const list = await listAllConversations();
		setAllItems(list);
		setVisible(list);
	}, []);

	useEffect(() => {
		fetchList();
	}, [fetchList, refreshKey]);

	const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
		const v = e.target.value;
		setQuery(v);

		if (!v.trim()) {
			setVisible(allItems);
			return;
		}

		const lower = v.toLowerCase();
		setVisible(allItems.filter(it => it.title.toLowerCase().includes(lower)));
		setFocused(-1);
	};

	const pick = async (it: ConversationItem) => {
		await onSelectConversation(it);
		setQuery('');
		setShow(false);
		setFocused(-1);
	};

	const handleKey = (e: KeyboardEvent<HTMLInputElement>) => {
		if (!visible.length) return;

		if (e.key === 'ArrowDown') setFocused(i => (i + 1) % visible.length);
		else if (e.key === 'ArrowUp') setFocused(i => (i - 1 + visible.length) % visible.length);
		else if (e.key === 'Enter' && focusedIndex >= 0) pick(visible[focusedIndex]);
		else if (e.key === 'Escape') {
			setShow(false);
			setFocused(-1);
		}
	};

	return (
		<div className="relative">
			<div className="flex items-center bg-base-100 py-3 rounded-2xl border">
				<FiSearch size={24} className="mx-2 text-neutral/60" />
				<input
					type="text"
					value={query}
					onChange={handleChange}
					onFocus={() => {
						setShow(true);
						setVisible(allItems);
					}}
					onBlur={() =>
						setTimeout(() => {
							setShow(false);
						}, 100)
					}
					onKeyDown={handleKey}
					placeholder="Previous chats…"
					className="w-full bg-base-100 outline-none"
					style={{ fontSize: 14 }}
					spellCheck={false}
				/>
			</div>

			{show && <GroupedDropdown items={visible} onPick={pick} focused={focusedIndex} />}
		</div>
	);
};

export default ChatSearch;
