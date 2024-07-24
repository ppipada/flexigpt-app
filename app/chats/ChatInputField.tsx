import React, { useEffect, useRef, useState, type RefObject } from "react";
import { FiSend } from "react-icons/fi";

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

  const handleKeyDown = (
    event: React.KeyboardEvent<HTMLTextAreaElement>
  ): void => {
    if (
      event.key === "Enter" &&
      !event.shiftKey &&
      !event.nativeEvent.isComposing
    ) {
      formRef.current?.requestSubmit();
      event.preventDefault();
    }
  };

  return { formRef, onKeyDown: handleKeyDown };
}

const ChatInputField: React.FC<ChatInputFieldProps> = ({
  onSend,
  setInputHeight,
}) => {
  const [text, setText] = useState<string>("");
  const [isSendButtonEnabled, setIsSendButtonEnabled] =
    useState<boolean>(false);
  const inputRef = useRef<HTMLTextAreaElement>(null);

  const { formRef, onKeyDown } = useEnterSubmit();

  const handleTextChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = event.target.value;
    setText(value);
    setIsSendButtonEnabled(value.trim().length > 0);
    autoResizeTextarea();
  };

  const handleSubmit = () => {
    if (text.trim().length === 0) return;
    onSend(text.trim());
    setText("");
    setIsSendButtonEnabled(false);
    if (inputRef.current) {
      inputRef.current.style.height = "auto";
      inputRef.current.style.height = `${inputRef.current.scrollHeight}px`;
      setInputHeight(inputRef.current.scrollHeight);
      inputRef.current?.focus();
    }
  };

  const autoResizeTextarea = () => {
    if (inputRef.current) {
      inputRef.current.style.height = "auto";
      inputRef.current.style.height = `${inputRef.current.scrollHeight}px`;
      setInputHeight(inputRef.current.scrollHeight);
    }
  };

  useEffect(() => {
    autoResizeTextarea();
  }, [text]);

  return (
    <div className="relative w-full">
      <form
        ref={formRef}
        onSubmit={(e) => {
          e.preventDefault();
          handleSubmit();
        }}
        className="flex items-center w-full"
      >
        <textarea
          ref={inputRef}
          value={text}
          onChange={handleTextChange}
          onKeyDown={onKeyDown}
          placeholder="Type message..."
          className="textarea textarea-bordered flex-1 resize-none overflow-hidden min-h-[24px] max-h-[240px]"
          rows={1}
        />
        <button
          type="submit"
          className={`btn btn-md !bg-transparent border-none shadow-none m-0 p-2 ${
            !isSendButtonEnabled ? "btn-disabled" : ""
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
