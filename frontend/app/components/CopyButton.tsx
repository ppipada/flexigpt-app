import React, { FC, useState } from "react";
import { FiCheck, FiCopy } from "react-icons/fi";

export interface CopyButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  value?: string;
  size: number;
}

const CopyButton: FC<CopyButtonProps> = ({ value, size, ...buttonProps }) => {
  const [copied, setCopied] = useState<boolean>(false);

  const handleCopy = async () => {
    if (!value) {
      return;
    }
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch (error) {
      console.error("Failed to copy: ", error);
    }
  };

  return (
    <button
      aria-label="Copy to clipboard"
      onClick={handleCopy}
      {...buttonProps}
    >
      {copied ? <FiCheck size={size} /> : <FiCopy size={size} />}
    </button>
  );
};

export default CopyButton;
