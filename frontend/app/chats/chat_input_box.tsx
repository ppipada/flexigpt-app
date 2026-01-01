import { forwardRef, type RefObject, useEffect, useImperativeHandle, useRef, useState } from 'react';

import type { ToolCallBinding } from '@/spec/inference';
import type { ToolStoreChoice, UIToolCallChip } from '@/spec/tool';

import type { ShortcutConfig } from '@/lib/keyboard_shortcuts';

import { DeleteConfirmationModal } from '@/components/delete_confirmation_modal';

import { AssistantContextBar } from '@/chats/assitantcontexts/context_bar';
import {
	EditorArea,
	type EditorAreaHandle,
	type EditorExternalMessage,
	type EditorSubmitPayload,
} from '@/chats/chat_input_editor';
import { type ChatOption, DefaultChatOptions } from '@/chats/chat_option_helper';

export interface InputBoxHandle {
	getChatOptions: () => ChatOption;
	focus: () => void;
	openTemplateMenu: () => void;
	openToolMenu: () => void;
	openAttachmentMenu: () => void;
	loadExternalMessage: (msg: EditorExternalMessage) => void;
	loadToolCalls: (toolCalls: UIToolCallChip[], bindings?: ToolCallBinding[]) => void;
	setConversationToolsFromChoices: (tools: ToolStoreChoice[]) => void;
}

interface InputBoxProps {
	onSend: (message: EditorSubmitPayload, options: ChatOption) => Promise<void>;
	isBusy: boolean;
	abortRef: RefObject<AbortController | null>;
	shortcutConfig: ShortcutConfig;
	editingMessageId: string | null;
	cancelEditing: () => void;
}

export const InputBox = forwardRef<InputBoxHandle, InputBoxProps>(function InputBox(
	{ onSend, isBusy, abortRef, shortcutConfig, editingMessageId, cancelEditing },
	ref
) {
	/* ------------------------------------------------------------------
	 * Aggregated chat-options (provided by <AssistantContextBar />)
	 * ------------------------------------------------------------------ */
	const [chatOptions, setChatOptions] = useState<ChatOption>(DefaultChatOptions);

	/* ------------------------------------------------------------------
	 * Abort-handling helpers
	 * ------------------------------------------------------------------ */
	const [showAbortModal, setShowAbortModal] = useState(false);

	useEffect(() => {
		if (!isBusy) setShowAbortModal(false);
	}, [isBusy]);

	/* ------------------------------------------------------------------
	 * <EditorArea /> ref utilities
	 * ------------------------------------------------------------------ */
	const inputAreaRef = useRef<EditorAreaHandle>(null);

	/* ------------------------------------------------------------------
	 * Send-message
	 * ------------------------------------------------------------------ */
	const handleSubmitMessage = async (payload: EditorSubmitPayload) => {
		onSend(payload, chatOptions);
	};

	useImperativeHandle(ref, () => ({
		getChatOptions: () => chatOptions,
		focus: () => {
			inputAreaRef.current?.focus();
		},
		openTemplateMenu: () => {
			inputAreaRef.current?.openTemplateMenu();
		},
		openToolMenu: () => {
			inputAreaRef.current?.openToolMenu();
		},
		openAttachmentMenu: () => {
			inputAreaRef.current?.openAttachmentMenu();
		},
		loadExternalMessage: msg => {
			inputAreaRef.current?.loadExternalMessage(msg);
		},
		loadToolCalls: toolCalls => {
			inputAreaRef.current?.loadToolCalls(toolCalls);
		},
		setConversationToolsFromChoices: tools => {
			inputAreaRef.current?.setConversationToolsFromChoices(tools);
		},
	}));

	/* ------------------------------------------------------------------
	 * Render
	 * ------------------------------------------------------------------ */
	return (
		<div className="bg-base-200 w-full min-w-0">
			{/* Model- / params-bar ---------------------------------------------- */}
			<AssistantContextBar onOptionsChange={setChatOptions} /* hand the aggregated options up */ />

			<DeleteConfirmationModal
				isOpen={showAbortModal}
				onClose={() => {
					setShowAbortModal(false);
				}}
				onConfirm={() => {
					setShowAbortModal(false);
					abortRef.current?.abort();
				}}
				title="Abort generation?"
				message="Partial answer that has already been received will stay in the chat. Do you want to stop the request?"
				confirmButtonText="Abort"
			/>

			<div className="overflow-x-hidden overflow-y-auto">
				<EditorArea
					ref={inputAreaRef}
					isBusy={isBusy}
					shortcutConfig={shortcutConfig}
					onSubmit={handleSubmitMessage}
					onRequestStop={() => {
						setShowAbortModal(true);
					}}
					editingMessageId={editingMessageId}
					cancelEditing={cancelEditing}
				/>
			</div>
		</div>
	);
});
