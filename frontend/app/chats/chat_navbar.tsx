import type { FC } from 'react';

import { FiPlus } from 'react-icons/fi';

import type { ConversationItem } from '@/models/conversationmodel';

import DownloadButton from '@/components/download_button';

import ChatSearch from '@/chats/chat_search';

interface ChatNavBarProps {
	onNewChat: () => void;
	getConversationForExport: () => Promise<string>;
	onSelectConversation: (item: ConversationItem) => Promise<void>;
	chatTitle: string;
	searchRefreshKey: number;
}

const ChatNavBar: FC<ChatNavBarProps> = ({
	onNewChat,
	getConversationForExport,
	onSelectConversation,
	chatTitle,
	searchRefreshKey,
}) => {
	return (
		<div className="flex-1 flex-col items-center">
			<div className="flex-1 items-center justify-between p-2 bg-transparent ml-8 md:ml-0">
				<ChatSearch onSelectConversation={onSelectConversation} refreshKey={searchRefreshKey} />
			</div>
			<div className="flex items-center justify-between p-0 mt-2 max-h-8 bg-transparent">
				<button
					className="btn btn-sm btn-ghost mx-1"
					onClick={onNewChat}
					aria-label="Create New Chat"
					title="Create New Chat"
				>
					<FiPlus size={24} />
				</button>
				<h1 className="text-xl font-semibold text-center text-nowrap overflow-hidden">{chatTitle}</h1>
				<DownloadButton
					language="json"
					valueFetcher={getConversationForExport}
					size={24}
					fileprefix="conversation"
					className="btn btn-sm btn-ghost mx-1"
					aria-label="Export Chat"
					title="Export Chat"
				/>
			</div>
		</div>
	);
};

export default ChatNavBar;
