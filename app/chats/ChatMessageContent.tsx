import { FC, ReactNode, memo } from "react";
import ReactMarkdown from "react-markdown";
import SyntaxHighlighter from "react-syntax-highlighter";
import { monokaiSublime } from "react-syntax-highlighter/dist/esm/styles/hljs";
import CopyButton from "../components/CopyButton";
import DownloadButton from "../components/DownloadButton";

export const MemoizedMarkdown = memo(
  ReactMarkdown,
  (prevProps, nextProps) =>
    prevProps.children === nextProps.children &&
    prevProps.className === nextProps.className
);

export interface ChatMessageContentProps {
  content: string;
  align: string;
}

interface CodeProps {
  language: string;
  value: string;
}

const CodeBlock: FC<CodeProps> = memo(({ language, value }) => {
  return (
    <div className="rounded-md bg-gray-800 my-2 items-start overflow-hidden">
      <div className="flex justify-between items-center bg-gray-700 px-4">
        <span className="text-white">{language}</span>
        <div className="flex space-x-2">
          <DownloadButton
            language={language}
            value={value}
            size={16}
            className="btn btn-sm bg-transparent text-white border-none flex items-center shadow-none"
          />
          <CopyButton
            value={value}
            className="btn btn-sm bg-transparent text-white border-none flex items-center shadow-none"
            size={16}
          />
        </div>
      </div>
      <div className="p-1">
        <SyntaxHighlighter
          language={language}
          style={monokaiSublime}
          showLineNumbers
          customStyle={{
            background: "transparent",
            padding: "0.5em",
            borderRadius: "0.25rem",
          }}
        >
          {value}
        </SyntaxHighlighter>
      </div>
    </div>
  );
});

interface CodeComponentProps {
  inline?: boolean;
  className?: string;
  children?: ReactNode;
}

interface PComponentProps {
  children?: ReactNode;
}

export function ChatMessageContent({
  content,
  align,
}: ChatMessageContentProps) {
  const components = {
    p({ children }: PComponentProps) {
      return (
        <p className={`my-2 ${align}`} style={{ lineHeight: "1.5" }}>
          {children}
        </p>
      );
    },
    code: ({ inline, className, children, ...props }: CodeComponentProps) => {
      if (inline || !className) {
        return (
          <code className="bg-base-200 p-1 rounded" {...props}>
            {children}
          </code>
        );
      }
      const match =
        /lang-(\w+)/.exec(className || "") ||
        /language-(\w+)/.exec(className || "");
      const language = match && match[1] ? match[1] : "text";

      return (
        <CodeBlock
          language={language}
          value={String(children).replace(/\n$/, "")}
          {...props}
        />
      );
    },
  };

  return (
    <div className="bg-base-100 rounded-lg shadow-lg p-2">
      <MemoizedMarkdown components={components}>{content}</MemoizedMarkdown>
    </div>
  );
}
