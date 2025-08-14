import type { ChangeEvent, KeyboardEvent } from 'react';
import React, { forwardRef, useImperativeHandle, useRef, useState } from 'react';

import { FiSend } from 'react-icons/fi';
import TextareaAutosize from 'react-textarea-autosize';

export interface ChatTextInputHandle {
	focus: () => void;
}

interface ChatTextInputProps {
	isBusy: boolean;
	onSubmit: (text: string) => void;
	setInputHeight: React.Dispatch<React.SetStateAction<number>>;
}

// Custom hook for handling form submission on Enter key press.
function useEnterSubmit(): {
	formRef: React.RefObject<HTMLFormElement | null>;
	onKeyDown: (event: KeyboardEvent<HTMLTextAreaElement>) => void;
} {
	const formRef = useRef<HTMLFormElement>(null);

	const handleKeyDown = (event: KeyboardEvent<HTMLTextAreaElement>) => {
		if (event.key === 'Enter' && !event.shiftKey && !event.nativeEvent.isComposing) {
			if (formRef.current) {
				formRef.current.requestSubmit();
			}
			event.preventDefault();
		}
	};

	return { formRef, onKeyDown: handleKeyDown };
}

const ChatTextInput = forwardRef<ChatTextInputHandle, ChatTextInputProps>(
	({ isBusy, onSubmit, setInputHeight }, ref) => {
		const [text, setText] = useState<string>('');
		const inputRef = useRef<HTMLTextAreaElement>(null);
		const isSubmittingRef = useRef<boolean>(false);
		const isSendButtonEnabled = text.trim().length > 0;

		const { formRef, onKeyDown } = useEnterSubmit();

		const handleHeightChange = (height: number) => {
			setInputHeight(height);
		};

		const handleTextChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
			setText(event.target.value);
		};

		const handleSubmit = (e?: React.FormEvent) => {
			if (e) e.preventDefault();
			if (text.trim().length === 0 || isSubmittingRef.current) return;

			isSubmittingRef.current = true;
			onSubmit(text.trim());
			setText('');
			isSubmittingRef.current = false;

			inputRef.current?.focus();
		};

		useImperativeHandle(ref, () => ({
			focus: () => inputRef.current?.focus(),
		}));

		return (
			<form
				ref={formRef}
				onSubmit={handleSubmit}
				className="flex items-center bg-base-100 rounded-2xl border border-base-300 focus-within:border-base-400 px-4 mx-2"
			>
				<TextareaAutosize
					ref={inputRef}
					value={text}
					onChange={handleTextChange}
					onKeyDown={onKeyDown}
					onHeightChange={handleHeightChange}
					placeholder="Type message..."
					className="flex-1 resize-none overflow-auto bg-transparent border-none outline-none min-h-[24px] p-2"
					minRows={2}
					maxRows={16}
					style={{ fontSize: '14px' }}
					spellCheck={false}
					disabled={isBusy}
				/>

				<button
					type="submit"
					className={`btn btn-md !bg-transparent border-none shadow-none px-1 ${!isSendButtonEnabled || isBusy ? 'btn-disabled' : ''}`}
					disabled={isBusy || !isSendButtonEnabled}
					aria-label="Send Message"
					title="Send Message"
				>
					<FiSend size={20} />
				</button>
			</form>
		);
	}
);

ChatTextInput.displayName = 'ChatTextInput';

export default ChatTextInput;
