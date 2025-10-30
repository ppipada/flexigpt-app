import type { ChangeEvent, Dispatch, FormEvent, KeyboardEvent, RefObject, SetStateAction } from 'react';
import { forwardRef, useImperativeHandle, useRef, useState } from 'react';

import { FiSend } from 'react-icons/fi';

import TextareaAutosize from 'react-textarea-autosize';

export interface TextAreaHandle {
	focus: () => void;
}

interface TextAreaProps {
	isBusy: boolean;
	onSubmit: (text: string) => void;
	setInputHeight: Dispatch<SetStateAction<number>>;
}

// Custom hook for handling form submission on Enter key press.
function useEnterSubmit(): {
	formRef: RefObject<HTMLFormElement | null>;
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

export const TextArea = forwardRef<TextAreaHandle, TextAreaProps>(function TextArea(
	{ isBusy, onSubmit, setInputHeight },
	ref
) {
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

	const handleSubmit = (e?: FormEvent) => {
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
			className="bg-base-100 border-base-300 focus-within:border-base-400 mx-2 flex items-center rounded-2xl border px-4"
		>
			<TextareaAutosize
				ref={inputRef}
				value={text}
				onChange={handleTextChange}
				onKeyDown={onKeyDown}
				onHeightChange={handleHeightChange}
				placeholder="Type message..."
				className="min-h-6 flex-1 resize-none overflow-auto border-none bg-transparent p-2 outline-none"
				minRows={2}
				maxRows={16}
				style={{ fontSize: '14px' }}
				spellCheck={false}
				disabled={isBusy}
			/>

			<button
				type="submit"
				className={`btn btn-md border-none bg-transparent! px-1 shadow-none ${!isSendButtonEnabled || isBusy ? 'btn-disabled' : ''}`}
				disabled={isBusy || !isSendButtonEnabled}
				aria-label="Send Message"
				title="Send Message"
			>
				<FiSend size={20} />
			</button>
		</form>
	);
});
