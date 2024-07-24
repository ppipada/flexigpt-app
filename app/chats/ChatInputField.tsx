import React, { useRef, useState } from "react";
import { FiSend } from "react-icons/fi";
import { useEnterSubmit } from "../lib/hooks/use-enter-submit";

interface ChatInputFieldProps {
  onSend: (message: string) => void;
}

const ChatInputField: React.FC<ChatInputFieldProps> = ({ onSend }) => {
  const [text, setText] = useState<string>("");
  const [isSendButtonEnabled, setIsSendButtonEnabled] =
    useState<boolean>(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const { formRef, onKeyDown } = useEnterSubmit();

  const handleTextChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const value = event.target.value;
    setText(value);
    setIsSendButtonEnabled(value.trim().length > 0);
  };

  const handleSubmit = () => {
    if (text.trim().length === 0) return;
    onSend(text.trim());
    setText("");
    setIsSendButtonEnabled(false);
    inputRef.current?.focus();
  };

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
        <input
          ref={inputRef}
          type="text"
          value={text}
          onChange={handleTextChange}
          onKeyDown={onKeyDown}
          placeholder="Type message..."
          className="input input-bordered flex-1"
        />
        <button
          type="submit"
          className={`btn !bg-transparent shadow-none ml-1 ${
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
