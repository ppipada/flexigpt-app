import type { ChangeEvent, FC, KeyboardEvent } from 'react';
import { useCallback, useEffect, useState } from 'react';

import { FiSearch } from 'react-icons/fi';

import type { ConversationItem } from '@/models/conversationmodel';

import { listAllConversations } from '@/apis/conversationstore_helper';

import { GroupedDropdown } from '@/components/date_grouped_dropdown';

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

			{show && (
				<GroupedDropdown
					items={visible}
					getDate={item => new Date(item.createdAt)}
					getKey={item => item.id}
					getLabel={item => item.title}
					onPick={item => {
						pick(item);
					}}
					focused={focusedIndex}
					renderItemExtra={item =>
						new Date(item.createdAt).toLocaleDateString('en-US', {
							day: '2-digit',
							month: 'short',
							year: 'numeric',
						})
					}
				/>
			)}
		</div>
	);
};

export default ChatSearch;
