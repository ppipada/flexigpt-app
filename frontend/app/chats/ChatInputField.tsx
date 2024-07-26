import React, { useCallback, useEffect, useRef, useState, type RefObject } from 'react';
import { FiSend } from 'react-icons/fi';

interface ChatInputFieldProps {
	onSend: (message: string) => void;
	setInputHeight: (height: number) => void;
}

// Custom hook for handling form submission on Enter key press
function useEnterSubmit(): {
	formRef: RefObject<HTMLFormElement>;
	onKeyDown: (event: React.KeyboardEvent<HTMLTextAreaElement>) => void;
} {
	const formRef = useRef<HTMLFormElement>(null);

	const handleKeyDown = (event: React.KeyboardEvent<HTMLTextAreaElement>): void => {
		if (event.key === 'Enter' && !event.shiftKey && !event.nativeEvent.isComposing) {
			formRef.current?.requestSubmit();
			event.preventDefault();
		}
	};

	return { formRef, onKeyDown: handleKeyDown };
}

const ChatInputField: React.FC<ChatInputFieldProps> = ({ onSend, setInputHeight }) => {
	const [text, setText] = useState<string>('');
	const [isSendButtonEnabled, setIsSendButtonEnabled] = useState<boolean>(false);
	const inputRef = useRef<HTMLTextAreaElement>(null);

	const { formRef, onKeyDown } = useEnterSubmit();

	const autoResizeTextarea = useCallback(() => {
		if (inputRef.current) {
			inputRef.current.style.height = 'auto';
			inputRef.current.style.height = `${inputRef.current.scrollHeight}px`;
			setInputHeight(inputRef.current.scrollHeight);
		}
	}, [setInputHeight]);

	const handleTextChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
		const value = event.target.value;
		setText(value);
		setIsSendButtonEnabled(value.trim().length > 0);
		autoResizeTextarea();
	};

	const handleSubmit = () => {
		if (text.trim().length === 0) return;
		onSend(text.trim());
		setText('');
		setIsSendButtonEnabled(false);
		if (inputRef.current) {
			inputRef.current.style.height = 'auto';
			inputRef.current.style.height = `${inputRef.current.scrollHeight}px`;
			setInputHeight(inputRef.current.scrollHeight);
			inputRef.current?.focus();
		}
	};

	useEffect(() => {
		autoResizeTextarea();
	}, [text, autoResizeTextarea]);

	return (
		<div className="relative">
			<form
				ref={formRef}
				onSubmit={e => {
					e.preventDefault();
					handleSubmit();
				}}
				className="flex items-center bg-base-100 rounded-2xl border px-4 mx-2"
			>
				<textarea
					ref={inputRef}
					value={text}
					onChange={handleTextChange}
					onKeyDown={onKeyDown}
					placeholder="Type message..."
					className="flex-1 resize-none overflow-hidden bg-transparent border-none outline-none placeholder-gray-400 min-h-[24px] max-h-[240px] p-2"
					rows={1}
				/>
				<button
					type="submit"
					className={`btn btn-md !bg-transparent border-none shadow-none px-1 ${
						!isSendButtonEnabled ? 'btn-disabled' : ''
					}`}
					disabled={!isSendButtonEnabled}
					aria-label="Send message"
				>
					<FiSend size={24} />
				</button>
			</form>
		</div>
	);
};

export default ChatInputField;
