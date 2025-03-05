import { ChangeEvent, FC, KeyboardEvent, useCallback, useEffect, useRef, useState } from 'react';
import { FiSend, FiX } from 'react-icons/fi';

interface EditBoxProps {
	editText: string;
	onTextChange: (e: ChangeEvent<HTMLTextAreaElement>) => void;
	onSubmit: (text: string) => void;
	onDiscard: () => void;
}

const MAX_HEIGHT = 420;

const EditBox: FC<EditBoxProps> = ({ editText, onTextChange, onSubmit, onDiscard }) => {
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
		<div className="flex flex-col w-full p-2">
			<div className="relative w-full rounded-2xl overflow-hidden">
				<textarea
					ref={textAreaRef}
					value={editText}
					onChange={handleTextChange}
					onKeyDown={handleKeyDown}
					className="resize-none bg-base-100 p-4 w-full"
					rows={4}
					style={{ fontSize: '14px' }}
					spellCheck="false"
				/>
			</div>
			<div className="flex justify-end mt-0 mr-2 space-x-2">
				<button
					onClick={onDiscard}
					className="btn btn-md bg-transparent rounded-xl border-none shadow-none px-2"
					aria-label="Discard changes"
					title="Discard"
				>
					<FiX size={24} />
				</button>
				<button
					onClick={handleSubmit}
					className={`btn btn-md bg-transparent rounded-xl border-none shadow-none px-2 ${!isSendButtonEnabled ? 'btn-disabled' : ''}`}
					disabled={!isSendButtonEnabled}
					aria-label="Send edited message"
					title="Send"
				>
					<FiSend size={24} />
				</button>
			</div>
		</div>
	);
};

export default EditBox;
