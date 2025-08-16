import { type FC, useCallback, useEffect, useState } from 'react';

import { FiEdit, FiPlus } from 'react-icons/fi';

import type { ConversationSearchItem } from '@/spec/conversation';

import { sanitizeConversationTitle } from '@/lib/text_utils';

import DownloadButton from '@/components/download_button';

import ChatSearch from '@/chats/chat_search';

interface ChatNavBarProps {
	onNewChat: () => void;
	onRenameTitle: (newTitle: string) => void;
	getConversationForExport: () => Promise<string>;
	onSelectConversation: (item: ConversationSearchItem) => Promise<void>;
	chatTitle: string;
	chatID: string;
	searchRefreshKey: number;
	disabled: boolean;
	renameEnabled: boolean;
}

const ChatNavBar: FC<ChatNavBarProps> = ({
	onNewChat,
	onRenameTitle,
	getConversationForExport,
	onSelectConversation,
	chatTitle,
	chatID,
	searchRefreshKey,
	disabled,
	renameEnabled,
}) => {
	const [isEditing, setIsEditing] = useState(false);
	const [draftTitle, setDraftTitle] = useState(chatTitle);

	/* keep draft in sync when conversation switches */
	useEffect(() => {
		if (!isEditing) setDraftTitle(chatTitle);
	}, [chatTitle, isEditing]);

	const finishEdit = useCallback(() => {
		const cleaned = sanitizeConversationTitle(draftTitle.trim());
		if (cleaned && cleaned !== chatTitle) onRenameTitle(cleaned);
		setIsEditing(false);
	}, [draftTitle, chatTitle, onRenameTitle]);

	useEffect(() => {
		if (!renameEnabled && isEditing) setIsEditing(false);
	}, [renameEnabled, isEditing]);

	const editDisabled = disabled || !renameEnabled;

	return (
		<div className="flex-1 flex-col items-center">
			{/* search ------------------------------------------------ */}
			<div className="flex-1 items-center justify-between py-1 px-2 bg-transparent ml-8 md:ml-0">
				<ChatSearch
					onSelectConversation={onSelectConversation}
					refreshKey={searchRefreshKey}
					currentConversationId={chatID}
				/>
			</div>

			{/* controls / title ------------------------------------ */}
			<div className="flex items-center justify-between p-2 bg-transparent">
				{/* new chat */}
				<button
					className="btn btn-sm btn-ghost mx-1"
					onClick={onNewChat}
					disabled={disabled}
					aria-label="Create New Chat"
					title="Create New Chat"
				>
					<FiPlus size={20} />
				</button>

				{/* title or editor */}
				<div className="flex-1 flex justify-center px-2">
					{isEditing ? (
						<input
							autoFocus
							value={draftTitle}
							onChange={e => {
								setDraftTitle(e.target.value);
							}}
							onBlur={finishEdit}
							onKeyDown={e => {
								if (e.key === 'Enter') finishEdit();
								else if (e.key === 'Escape') setIsEditing(false);
							}}
							className="text-xl font-semibold text-center bg-base-100 rounded px-2 py-1 w-full max-w-[60vw] outline-none"
						/>
					) : (
						<h1
							className="text-xl font-semibold text-center text-nowrap overflow-hidden cursor-pointer"
							title="Double-click or click the pencil to rename"
							onDoubleClick={() => {
								if (!editDisabled) {
									setIsEditing(true);
								}
							}}
						>
							{chatTitle}
						</h1>
					)}
				</div>

				{/* edit title */}
				<button
					className="btn btn-sm btn-ghost mx-1"
					onClick={() => {
						if (!editDisabled) setIsEditing(true);
					}}
					disabled={editDisabled}
					aria-label="Rename Conversation"
					title="Rename Conversation"
				>
					<FiEdit size={20} />
				</button>

				{/* download */}
				<DownloadButton
					language="json"
					valueFetcher={getConversationForExport}
					size={20}
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
