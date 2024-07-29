import React from 'react';
import { FiDownload, FiPlus } from 'react-icons/fi';
import ChatSearch from './ChatSearch';

interface ChatNavBarProps {
	onNewChat: () => void;
	onExport: () => void;
	initialSearchItems: string[];
	onSearch: (query: string) => Promise<string[]>;
}

const ChatNavBar: React.FC<ChatNavBarProps> = ({ onNewChat, onExport, initialSearchItems, onSearch }) => {
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
				<ChatSearch initialItems={initialSearchItems} onSearch={onSearch} />
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
