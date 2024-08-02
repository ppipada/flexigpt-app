import { ChatSearch, SearchItem } from '@/chats/ChatSearch';
import { FC } from 'react';
import { FiDownload, FiPlus } from 'react-icons/fi';

interface ChatNavBarProps {
	onNewChat: () => void;
	onExport: () => void;
	initialSearchItems: SearchItem[];
	onSearch: (query: string) => Promise<SearchItem[]>;
	onSelectConversation: (item: SearchItem) => Promise<void>;
}

const ChatNavBar: FC<ChatNavBarProps> = ({
	onNewChat,
	onExport,
	initialSearchItems,
	onSearch,
	onSelectConversation,
}) => {
	return (
		<div className="flex items-center justify-between p-2 bg-transparent">
			<button
				className="btn btn-sm mx-1 bg-transparent shadow-none"
				onClick={onNewChat}
				aria-label="Create new chat"
				title="Create new chat"
			>
				<FiPlus size={24} />
			</button>
			<div className="flex-1">
				<ChatSearch initialItems={initialSearchItems} onSearch={onSearch} onSelectConversation={onSelectConversation} />
			</div>
			<button
				className="btn btn-sm mx-1 bg-transparent shadow-none"
				onClick={onExport}
				aria-label="Export chat"
				title="Export chat"
			>
				<FiDownload size={24} />
			</button>
		</div>
	);
};

export default ChatNavBar;
