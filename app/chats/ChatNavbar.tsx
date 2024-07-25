import React from "react";
import { FiDownload, FiPlus } from "react-icons/fi";
import ChatSearch from "./ChatSearch";

interface ChatNavBarProps {
  onNewChat: () => void;
  onExport: () => void;
  initialSearchItems: string[];
  onSearch: (query: string) => Promise<string[]>;
}

const ChatNavBar: React.FC<ChatNavBarProps> = ({
  onNewChat,
  onExport,
  initialSearchItems,
  onSearch,
}) => {
  return (
    <div className="flex items-center justify-between p-1 bg-transparent">
      <button
        className="btn bg-transparent shadow-none ml-4 mr-2"
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
        className="btn bg-transparent shadow-none ml-2 mr-4"
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
