import React, { forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react';

import { FiSquare } from 'react-icons/fi';

import { ReasoningType } from '@/spec/modelpreset';

import { type ChatOption, DefaultChatOptions, getChatInputOptions } from '@/apis/chatoption_helper';

import DeleteConfirmationModal from '@/components/delete_confirmation';

import ChatTextInput, { type ChatTextInputHandle } from '@/chats/inputbox/text_input';
import ChatModelParamsBar from '@/chats/modelparams/modelparams_bar';

interface ChatInputFieldProps {
	onSend: (message: string, options: ChatOption) => void;
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
		const [selectedModel, setSelectedModel] = useState<ChatOption>(DefaultChatOptions);
		const [allOptions, setAllOptions] = useState<ChatOption[]>([DefaultChatOptions]);

		const [isHybridReasoningEnabled, setIsHybridReasoningEnabled] = useState<boolean>(true);
		const [disablePreviousMessages, setDisablePreviousMessages] = useState<boolean>(false);

		const [showAbortModal, setShowAbortModal] = useState(false);

		const inputAreaRef = useRef<ChatTextInputHandle>(null);

		useEffect(() => {
			if (!isBusy) setShowAbortModal(false);
		}, [isBusy]);

		// Load initial model options
		const loadInitialItems = useCallback(async () => {
			const r = await getChatInputOptions();
			setSelectedModel(r.default);

			// Initialize hybrid reasoning enabled state based on model.
			setIsHybridReasoningEnabled(r.default.reasoning?.type === ReasoningType.HybridWithTokens);
			setAllOptions(r.allOptions);
		}, []);

		useEffect(() => {
			loadInitialItems();
		}, [loadInitialItems]);

		// When model changes, update hybrid reasoning enabled state.
		useEffect(() => {
			setIsHybridReasoningEnabled(selectedModel.reasoning?.type === ReasoningType.HybridWithTokens);
		}, [selectedModel]);

		// Build ChatOption for submit
		const getFinalChatOptions = (): ChatOption => {
			const options = { ...selectedModel, disablePreviousMessages };

			// If it's a hybrid reasoning model but user disabled reasoning, remove it.
			if (selectedModel.reasoning?.type === ReasoningType.HybridWithTokens && !isHybridReasoningEnabled) {
				const modifiedOptions = { ...options };
				delete modifiedOptions.reasoning;
				return modifiedOptions;
			}

			return options;
		};

		const handleSubmitMessage = (text: string) => {
			onSend(text, getFinalChatOptions());
		};

		useImperativeHandle(ref, () => ({
			getChatOptions: () => getFinalChatOptions(),
			focus: () => {
				inputAreaRef.current?.focus();
			},
		}));

		return (
			<div className="flex-1">
				{isBusy && (
					<div className="flex items-center justify-center bg-base-200 mb-1 mx-8">
						<button
							className="btn btn-sm bg-neutral text-neutral-content shadow-none border-none gap-1 rounded-xl"
							onClick={() => {
								setShowAbortModal(true);
							}}
						>
							<FiSquare />
							Stop
						</button>
					</div>
				)}

				{/* Model params bar */}
				<ChatModelParamsBar
					selectedModel={selectedModel}
					setSelectedModel={setSelectedModel}
					allOptions={allOptions}
					disablePreviousMessages={disablePreviousMessages}
					setDisablePreviousMessages={setDisablePreviousMessages}
					isHybridReasoningEnabled={isHybridReasoningEnabled}
					setIsHybridReasoningEnabled={setIsHybridReasoningEnabled}
				/>

				{/* Abort confirmation dialog */}
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

				{/* Input area */}
				<ChatTextInput
					ref={inputAreaRef}
					isBusy={isBusy}
					onSubmit={handleSubmitMessage}
					setInputHeight={setInputHeight}
				/>
			</div>
		);
	}
);

ChatInputField.displayName = 'ChatInputField';

export default ChatInputField;
