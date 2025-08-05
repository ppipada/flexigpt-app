/* -------------------------------------------------------------------------- */
/* ChatInputField.tsx                                                         */
/* -------------------------------------------------------------------------- */
import type { ChangeEvent, KeyboardEvent } from 'react';
import React, { forwardRef, useCallback, useEffect, useImperativeHandle, useRef, useState } from 'react';

import { FiSend, FiSliders } from 'react-icons/fi';
/* ----------  react-textarea-autosize  ------------------------------------ */
import TextareaAutosize from 'react-textarea-autosize';

import { type ReasoningLevel, ReasoningType } from '@/spec/modelpreset';

import { useCloseDetails } from '@/hooks/use_close_details';

import { type ChatOption, DefaultChatOptions, getChatInputOptions } from '@/apis/chatoption_helper';

import AdvancedParamsModal from '@/chats/chat_input_field_advanced_params';
import DisablePreviousMessagesCheckbox from '@/chats/chat_input_field_disable_checkbox';
import ModelDropdown from '@/chats/chat_input_field_model_dropdown';
import { HybridReasoningCheckbox, ReasoningTokensDropdown } from '@/chats/chat_input_field_reasoning_hybrid';
import SingleReasoningDropdown from '@/chats/chat_input_field_reasoning_levels';
import TemperatureDropdown from '@/chats/chat_input_field_temperature';

interface ChatInputFieldProps {
	onSend: (message: string, options: ChatOption) => void;
	setInputHeight: React.Dispatch<React.SetStateAction<number>>;
}

export interface ChatInputFieldHandle {
	getChatOptions: () => ChatOption;
	focus: () => void;
}

// Custom hook for handling form submission on Enter key press.
function useEnterSubmit(): {
	formRef: React.RefObject<HTMLFormElement | null>;
	onKeyDown: (event: KeyboardEvent<HTMLTextAreaElement>) => void;
} {
	const formRef = useRef<HTMLFormElement>(null);

	const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>): void => {
		if (event.key === 'Enter' && !event.shiftKey && !event.nativeEvent.isComposing) {
			if (formRef.current) {
				formRef.current.requestSubmit();
			}
			event.preventDefault();
		}
	};

	return { formRef, onKeyDown: handleKeyDown };
}

const ChatInputField = forwardRef<ChatInputFieldHandle, ChatInputFieldProps>(({ onSend, setInputHeight }, ref) => {
	const [text, setText] = useState<string>('');
	const isSendButtonEnabled = text.trim().length > 0;

	/* The ref now points to the autosizing textarea */
	const inputRef = useRef<HTMLTextAreaElement>(null);

	const isSubmittingRef = useRef<boolean>(false);
	const [selectedModel, setSelectedModel] = useState<ChatOption>(DefaultChatOptions);
	const [isAdvancedModalOpen, setIsAdvancedModalOpen] = useState<boolean>(false);
	const [isHybridReasoningEnabled, setIsHybridReasoningEnabled] = useState<boolean>(true);
	const [disablePreviousMessages, setDisablePreviousMessages] = useState<boolean>(false);
	const [isModelDropdownOpen, setIsModelDropdownOpen] = useState<boolean>(false);
	const [isSecondaryDropdownOpen, setIsSecondaryDropdownOpen] = useState<boolean>(false);
	const [allOptions, setAllOptions] = useState<ChatOption[]>([DefaultChatOptions]);

	// Refs for the dropdown details elements.
	const modelDetailsRef = useRef<HTMLDetailsElement>(null);
	const secondaryDetailsRef = useRef<HTMLDetailsElement>(null);

	// Close logic for model dropdown.
	useCloseDetails({
		detailsRef: modelDetailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsModelDropdownOpen(false);
		},
	});

	// Close logic for secondary dropdown (temperature or reasoning).
	useCloseDetails({
		detailsRef: secondaryDetailsRef,
		events: ['mousedown'],
		onClose: () => {
			setIsSecondaryDropdownOpen(false);
		},
	});

	// Load initial model options.
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

	// Enter key submission logic.
	const { formRef, onKeyDown } = useEnterSubmit();

	/* ------------- Update parent with actual textarea height ------------- */
	const handleHeightChange = useCallback(
		(height: number) => {
			setInputHeight(height);
		},
		[setInputHeight]
	);

	/* -------------------------- onChange -------------------------------- */
	const handleTextChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
		setText(event.target.value);
	};

	/* ------------------ build ChatOption for submit -------------------- */
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

	const handleSubmit = (e?: React.FormEvent) => {
		if (e) e.preventDefault();
		if (text.trim().length === 0 || isSubmittingRef.current) return;

		isSubmittingRef.current = true;

		onSend(text.trim(), getFinalChatOptions());

		setText('');
		isSubmittingRef.current = false;

		inputRef.current?.focus();
	};

	// Expose the function to get current chat options + focus.
	useImperativeHandle(ref, () => ({
		getChatOptions: () => getFinalChatOptions(),
		focus: () => {
			inputRef.current?.focus();
		},
	}));

	// Clamps temperature to [0, 1].
	const setTemperature = (temp: number) => {
		const clampedTemp = Math.max(0, Math.min(1, temp));
		setSelectedModel(prev => ({ ...prev, temperature: clampedTemp }));
	};

	// Set reasoning level for SingleWithLevels type.
	const setReasoningLevel = (newLevel: ReasoningLevel) => {
		setSelectedModel(prev => ({
			...prev,
			reasoning: {
				type: ReasoningType.SingleWithLevels,
				level: newLevel,
				tokens: 1024,
			},
		}));
	};

	// Set tokens for HybridWithTokens type.
	const setHybridTokens = (tokens: number) => {
		setSelectedModel(prev => {
			if (!prev.reasoning || prev.reasoning.type !== ReasoningType.HybridWithTokens) {
				return prev;
			}
			return {
				...prev,
				reasoning: { ...prev.reasoning, tokens },
			};
		});
	};

	// Handle saving the advanced parameters from the modal.
	// We simply merge these changes back into selectedModel.
	const handleSaveAdvancedParams = (updatedModel: ChatOption) => {
		setSelectedModel(updatedModel);
		setIsAdvancedModalOpen(false);
	};

	const secondaryProps = {
		isOpen: isSecondaryDropdownOpen,
		setIsOpen: setIsSecondaryDropdownOpen,
		detailsRef: secondaryDetailsRef,
	};

	return (
		<div className="relative">
			{/* ---------- First row : model & options ------------------------ */}
			<div className="flex items-center justify-between bg-base-200 mb-1 mx-8">
				{/* Model dropdown */}
				<div className="w-1/3">
					<ModelDropdown
						selectedModel={selectedModel}
						setSelectedModel={setSelectedModel}
						allOptions={allOptions}
						isOpen={isModelDropdownOpen}
						setIsOpen={setIsModelDropdownOpen}
						detailsRef={modelDetailsRef}
					/>
				</div>

				{/* Middle section (reasoning / temperature  */}
				<div className="flex items-center justify-between w-2/3">
					{selectedModel.reasoning?.type === ReasoningType.HybridWithTokens && (
						<HybridReasoningCheckbox
							isReasoningEnabled={isHybridReasoningEnabled}
							setIsReasoningEnabled={setIsHybridReasoningEnabled}
						/>
					)}

					{selectedModel.reasoning?.type === ReasoningType.HybridWithTokens ? (
						isHybridReasoningEnabled ? (
							/* ─── checkbox ON → tokens ─── */
							<ReasoningTokensDropdown
								tokens={selectedModel.reasoning.tokens}
								setTokens={setHybridTokens}
								{...secondaryProps}
							/>
						) : (
							/* ─── checkbox OFF → temperature ─── */
							<TemperatureDropdown
								temperature={selectedModel.temperature ?? 0.1}
								setTemperature={setTemperature}
								{...secondaryProps}
							/>
						)
					) : selectedModel.reasoning?.type === ReasoningType.SingleWithLevels ? (
						/* single-level reasoning */
						<SingleReasoningDropdown
							reasoningLevel={selectedModel.reasoning.level}
							setReasoningLevel={setReasoningLevel}
							{...secondaryProps}
						/>
					) : (
						/* model has no reasoning field at all */
						<TemperatureDropdown
							temperature={selectedModel.temperature ?? 0.1}
							setTemperature={setTemperature}
							{...secondaryProps}
						/>
					)}

					<DisablePreviousMessagesCheckbox
						disablePreviousMessages={disablePreviousMessages}
						setDisablePreviousMessages={setDisablePreviousMessages}
					/>

					{/* Advanced params modal trigger */}
					<button
						type="button"
						className="btn btn-sm btn-ghost mx-2 text-neutral/60"
						onClick={() => {
							setIsAdvancedModalOpen(true);
						}}
						title="Set Advanced Params"
					>
						<FiSliders size={16} />
					</button>
				</div>
			</div>

			{/* ------------------ Advanced params modal ---------------------- */}
			{isAdvancedModalOpen && (
				<AdvancedParamsModal
					isOpen={isAdvancedModalOpen}
					onClose={() => {
						setIsAdvancedModalOpen(false);
					}}
					currentModel={selectedModel}
					onSave={handleSaveAdvancedParams}
				/>
			)}

			{/* -------------------- Main input form -------------------------- */}
			<form
				ref={formRef}
				onSubmit={handleSubmit}
				className="flex items-center bg-base-100 rounded-2xl border border-base-300 focus-within:border-base-400 px-4 mx-2"
			>
				{/* -----------------  react-textarea-autosize  ---------------- */}
				<TextareaAutosize
					ref={inputRef}
					value={text}
					onChange={handleTextChange}
					onKeyDown={onKeyDown}
					onHeightChange={handleHeightChange}
					placeholder="Type message..."
					className="flex-1 resize-none overflow-auto bg-transparent border-none outline-none text-neutral
                       min-h-[24px] p-2"
					minRows={2}
					maxRows={16}
					style={{
						fontSize: '14px',
					}}
					spellCheck={false}
				/>

				{/* --------------------- Send button ------------------------- */}
				<button
					type="submit"
					className={`btn btn-md !bg-transparent border-none shadow-none px-1 ${
						!isSendButtonEnabled ? 'btn-disabled' : ''
					}`}
					disabled={!isSendButtonEnabled}
					aria-label="Send Message"
					title="Send Message"
				>
					<FiSend size={24} />
				</button>
			</form>
		</div>
	);
});

ChatInputField.displayName = 'ChatInputField';

export default ChatInputField;
