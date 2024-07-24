import React from "react";
import { FiDownload, FiPlus } from "react-icons/fi";

interface ChatNavBarProps {
  onNewChat: () => void;
  onExport: () => void;
  title: string;
}

const ChatNavBar: React.FC<ChatNavBarProps> = ({
  onNewChat,
  onExport,
  title,
}) => {
  return (
    <div className="flex items-center justify-between p-4 bg-transparent">
      <button
        className="btn bg-transparent shadow-none"
        onClick={onNewChat}
        aria-label="Create new chat"
        title="Create new chat"
      >
        <FiPlus size={24} />
      </button>
      <h1 className="text-xl font-semibold text-center flex-1">{title}</h1>
      <button
        className="btn bg-transparent shadow-none"
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
