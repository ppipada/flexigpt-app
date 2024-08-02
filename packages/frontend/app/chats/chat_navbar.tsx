import { ChatSearch, SearchItem } from '@/chats/chat_search';
import DownloadButton from '@/components/download_button';
import { FC } from 'react';
import { FiPlus } from 'react-icons/fi';

interface ChatNavBarProps {
	onNewChat: () => void;
	getConversationForExport: () => Promise<string>;
	initialSearchItems: SearchItem[];
	onSearch: (query: string) => Promise<SearchItem[]>;
	onSelectConversation: (item: SearchItem) => Promise<void>;
}

const ChatNavBar: FC<ChatNavBarProps> = ({
	onNewChat,
	getConversationForExport,
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
			<DownloadButton
				language="json"
				valueFetcher={getConversationForExport}
				size={24}
				fileprefix="conversation"
				className="btn btn-sm mx-1 bg-transparent shadow-none"
				aria-label="Export chat"
				title="Export chat"
			/>
		</div>
	);
};

export default ChatNavBar;
