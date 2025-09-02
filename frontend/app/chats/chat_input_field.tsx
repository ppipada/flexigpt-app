import React, { forwardRef, useEffect, useImperativeHandle, useRef, useState } from 'react';

import { FiSquare } from 'react-icons/fi';

import { type ChatOption, DefaultChatOptions } from '@/apis/chatoption_helper';

import DeleteConfirmationModal from '@/components/delete_confirmation';

import AssistantContextBar from '@/chats/assitantcontext/assistantcontext_bar';
import CommandTipsBar from '@/chats/command_tips_bar';
import EditorTextInput, { type EditorTextInputHandle } from '@/chats/inputeditor/editor_text_input';

interface ChatInputFieldProps {
	onSend: (message: string, options: ChatOption) => Promise<void>;
	setInputHeight: React.Dispatch<React.SetStateAction<number>>;
	isBusy: boolean;
	abortRef: React.RefObject<AbortController | null>;
}

export interface ChatInputFieldHandle {
	getChatOptions: () => ChatOption;
	focus: () => void;
}

const ChatInputField = forwardRef<ChatInputFieldHandle, ChatInputFieldProps>(
	({ onSend, setInputHeight, isBusy, abortRef }, ref) => {
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
		 * <ChatTextInput /> ref utilities
		 * ------------------------------------------------------------------ */
		// const inputAreaRef = useRef<ChatTextInputHandle>(null);
		const inputAreaRef = useRef<EditorTextInputHandle>(null);

		/* ------------------------------------------------------------------
		 * Send-message
		 * ------------------------------------------------------------------ */
		const handleSubmitMessage = async (text: string) => {
			onSend(text, chatOptions);
		};

		/* ------------------------------------------------------------------
		 * Expose imperative API
		 * ------------------------------------------------------------------ */
		useImperativeHandle(ref, () => ({
			getChatOptions: () => chatOptions,
			focus: () => {
				inputAreaRef.current?.focus();
			},
		}));

		/* ------------------------------------------------------------------
		 * Render
		 * ------------------------------------------------------------------ */
		return (
			<div className="bg-base-200 flex-1">
				{/* Busy / abort banner ------------------------------------------------ */}
				{isBusy && (
					<div className="bg-base-200 mx-8 mb-1 flex items-center justify-center">
						<button
							className="btn btn-sm bg-neutral text-neutral-content gap-1 rounded-xl border-none shadow-none"
							onClick={() => {
								setShowAbortModal(true);
							}}
						>
							<FiSquare />
							Stop
						</button>
					</div>
				)}

				{/* Model- / params-bar ---------------------------------------------- */}
				<AssistantContextBar onOptionsChange={setChatOptions} /* hand the aggregated options up */ />

				{/* Abort confirmation dialog ---------------------------------------- */}
				{showAbortModal && (
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
				)}
				<div className="flex-1 overflow-auto">
					{/* Chat text-input --------------------------------------------------- */}
					{/* <ChatTextInput
					ref={inputAreaRef}
					isBusy={isBusy}
					onSubmit={handleSubmitMessage}
					setInputHeight={setInputHeight}
				/> */}
					<EditorTextInput
						ref={inputAreaRef}
						isBusy={isBusy}
						onSubmit={handleSubmitMessage}
						setInputHeight={setInputHeight}
					/>
				</div>
				{/* Neutral tips bar under the editor */}
				<div className="mx-4 my-0">
					<CommandTipsBar />
				</div>
			</div>
		);
	}
);

ChatInputField.displayName = 'ChatInputField';

export default ChatInputField;
