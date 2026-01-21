import { forwardRef, type RefObject, useEffect, useImperativeHandle, useRef, useState } from 'react';

import type { UIToolCall } from '@/spec/inference';
import type { ToolStoreChoice } from '@/spec/tool';

import type { ShortcutConfig } from '@/lib/keyboard_shortcuts';

import { DeleteConfirmationModal } from '@/components/delete_confirmation_modal';

import { type ChatOption, DefaultChatOptions } from '@/chats/assitantcontexts/chat_option_helper';
import { AssistantContextBar } from '@/chats/assitantcontexts/context_bar';
import { EditorArea, type EditorAreaHandle } from '@/chats/inputbox/input_editor';
import type { EditorExternalMessage, EditorSubmitPayload } from '@/chats/inputbox/input_editor_utils';

export interface InputBoxHandle {
	getChatOptions: () => ChatOption;
	focus: () => void;
	openTemplateMenu: () => void;
	openToolMenu: () => void;
	openAttachmentMenu: () => void;
	loadExternalMessage: (msg: EditorExternalMessage) => void;
	loadToolCalls: (toolCalls: UIToolCall[]) => void;
	setConversationToolsFromChoices: (tools: ToolStoreChoice[]) => void;
	setWebSearchFromChoices: (tools: ToolStoreChoice[]) => void;
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
	const [chatOptions, setChatOptions] = useState<ChatOption>(DefaultChatOptions);

	const [showAbortModal, setShowAbortModal] = useState(false);

	useEffect(() => {
		if (!isBusy) setShowAbortModal(false);
	}, [isBusy]);

	const inputAreaRef = useRef<EditorAreaHandle>(null);

	const handleSubmitMessage = (payload: EditorSubmitPayload) => {
		// Return the promise so <EditorArea /> can await it and surface
		// any synchronous errors from sendMessage (e.g. validation).
		return onSend(payload, chatOptions);
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
		setWebSearchFromChoices: choices => {
			inputAreaRef.current?.setWebSearchFromChoices(choices);
		},
	}));

	return (
		<div className="bg-base-200 w-full min-w-0">
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
					currentProviderSDKType={chatOptions.providerSDKType}
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
