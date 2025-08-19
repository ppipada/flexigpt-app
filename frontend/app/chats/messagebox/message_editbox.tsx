import type { ChangeEvent, FC, KeyboardEvent } from 'react';
import { useCallback, useEffect, useRef, useState } from 'react';

import { FiSend, FiX } from 'react-icons/fi';

interface MessageEditBoxProps {
	editText: string;
	onTextChange: (e: ChangeEvent<HTMLTextAreaElement>) => void;
	onSubmit: (text: string) => void;
	onDiscard: () => void;
}

const MAX_HEIGHT = 420;

const MessageEditBox: FC<MessageEditBoxProps> = ({ editText, onTextChange, onSubmit, onDiscard }) => {
	const [isSendButtonEnabled, setIsSendButtonEnabled] = useState<boolean>(editText.trim().length > 0);
	const textAreaRef = useRef<HTMLTextAreaElement>(null);

	const autoResizeTextarea = useCallback(() => {
		if (textAreaRef.current) {
			textAreaRef.current.style.height = 'auto';
			textAreaRef.current.style.height = `${Math.min(textAreaRef.current.scrollHeight, MAX_HEIGHT)}px`;
		}
	}, []);

	useEffect(() => {
		autoResizeTextarea();
	}, [editText, autoResizeTextarea]);

	// Add this useEffect to focus and scroll into the textarea when the component mounts
	useEffect(() => {
		if (textAreaRef.current) {
			textAreaRef.current.focus();
			textAreaRef.current.scrollIntoView({ behavior: 'smooth', block: 'center' });
		}
	}, []);

	const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
		if (e.key === 'Enter' && !e.shiftKey) {
			onSubmit(editText);
			e.preventDefault();
		}
	};

	const handleTextChange = (e: ChangeEvent<HTMLTextAreaElement>) => {
		onTextChange(e);
		setIsSendButtonEnabled(e.target.value.trim().length > 0);
		autoResizeTextarea();
	};

	const handleSubmit = () => {
		onSubmit(editText);
	};

	return (
		<div className="flex w-full flex-col p-2">
			<div className="relative w-full overflow-hidden rounded-2xl">
				<textarea
					ref={textAreaRef}
					value={editText}
					onChange={handleTextChange}
					onKeyDown={handleKeyDown}
					className="bg-base-100 w-full resize-none p-4"
					rows={4}
					style={{ fontSize: '14px' }}
					spellCheck="false"
				/>
			</div>
			<div className="mt-0 mr-2 flex justify-end space-x-2">
				<button
					onClick={onDiscard}
					className="btn btn-md rounded-xl border-none bg-transparent px-2 shadow-none"
					aria-label="Discard Changes"
					title="Discard Changes"
				>
					<FiX size={20} />
				</button>
				<button
					onClick={handleSubmit}
					className={`btn btn-md rounded-xl border-none bg-transparent px-2 shadow-none ${!isSendButtonEnabled ? 'btn-disabled' : ''}`}
					disabled={!isSendButtonEnabled}
					aria-label="Send Edited Message"
					title="Send Edited Message"
				>
					<FiSend size={20} />
				</button>
			</div>
		</div>
	);
};

export default MessageEditBox;
