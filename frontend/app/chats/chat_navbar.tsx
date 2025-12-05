import { forwardRef, useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react';

import { FiEdit2, FiPlus } from 'react-icons/fi';

import type { ConversationSearchItem } from '@/spec/conversation';

import { formatShortcut, type ShortcutConfig } from '@/lib/keyboard_shortcuts';
import { sanitizeConversationTitle } from '@/lib/text_utils';

import { DownloadButton } from '@/components/download_button';

import { ChatSearch, type ChatSearchHandle } from '@/chats/chat_search';

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
	shortcutConfig: ShortcutConfig;
}

export interface ChatNavBarHandle {
	focusSearch: () => void;
}

export const ChatNavBar = forwardRef<ChatNavBarHandle, ChatNavBarProps>(function ChatNavBar(
	{
		onNewChat,
		onRenameTitle,
		getConversationForExport,
		onSelectConversation,
		chatTitle,
		chatID,
		searchRefreshKey,
		disabled,
		renameEnabled,
		shortcutConfig,
	}: ChatNavBarProps,
	ref
) {
	const [isEditing, setIsEditing] = useState(false);
	const [draftTitle, setDraftTitle] = useState(chatTitle);
	const searchRef = useRef<ChatSearchHandle | null>(null);
	const shortcutLabels = useMemo(
		() => ({
			newChat: formatShortcut(shortcutConfig.newChat),
		}),
		[shortcutConfig]
	);

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

	useImperativeHandle(ref, () => ({
		focusSearch: () => {
			searchRef.current?.focusInput();
		},
	}));

	const editDisabled = disabled || !renameEnabled;

	return (
		<div className="w-full justify-center p-2">
			<div className="flex items-center justify-between bg-transparent">
				<ChatSearch
					ref={searchRef}
					onSelectConversation={onSelectConversation}
					refreshKey={searchRefreshKey}
					currentConversationId={chatID}
				/>
			</div>

			{/* controls / title ------------------------------------ */}
			<div className="flex items-center justify-between bg-transparent p-2">
				{/* new chat */}
				<div
					className="tooltip tooltip-right"
					data-tip={shortcutLabels.newChat ? `Create New Chat (${shortcutLabels.newChat})` : 'Create New Chat'}
				>
					<button
						className="btn btn-sm btn-ghost mx-1"
						onClick={onNewChat}
						disabled={disabled}
						aria-label="Create New Chat"
						title="Create New Chat"
					>
						<FiPlus size={20} />
					</button>
				</div>
				{/* title or editor */}
				<div className="flex flex-1 justify-center px-2">
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
							className="bg-base-100 w-full max-w-[60vw] rounded px-2 py-1 text-center text-xl font-semibold outline-none"
						/>
					) : (
						<h1
							className="cursor-pointer overflow-hidden text-center text-xl font-semibold text-nowrap"
							title="Click the title or the pencil to rename"
							onClick={() => {
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
				<div className="tooltip tooltip-left" data-tip="Rename Conversation">
					<button
						className="btn btn-sm btn-ghost mx-1"
						onClick={() => {
							if (!editDisabled) setIsEditing(true);
						}}
						disabled={editDisabled}
						aria-label="Rename Conversation"
						title="Rename Conversation"
					>
						<FiEdit2 size={20} />
					</button>
				</div>

				{/* download */}
				<div className="tooltip tooltip-left" data-tip="Export Conversation As JSON">
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
		</div>
	);
});
