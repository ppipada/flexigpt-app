import { MermaidDiagram } from '@/chats/chat_message_content_mermaid';
import CopyButton from '@/components/copy_button';
import DownloadButton from '@/components/download_button';
import 'katex/dist/katex.min.css';
import { FC } from 'react';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { monokaiSublime } from 'react-syntax-highlighter/dist/esm/styles/hljs';

interface CodeProps {
	language: string;
	value: string;
	streamedMessage: string;
}

export const CodeBlock: FC<CodeProps> = ({ language, value, streamedMessage }) => {
	const fetchValue = async (): Promise<string> => {
		return value;
	};
	const isMermaid = language.toLowerCase() === 'mermaid';
	return (
		<>
			<div className="rounded-md bg-gray-800 my-4 items-start overflow-hidden">
				<div className="flex justify-between items-center bg-gray-700 px-4">
					<span className="text-white">{language}</span>
					<div className="flex space-x-2">
						<DownloadButton
							language={language}
							valueFetcher={fetchValue}
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
							background: 'transparent',
							padding: '0.5em',
							borderRadius: '0.25rem',
							fontSize: '14px',
						}}
					>
						{value}
					</SyntaxHighlighter>
				</div>
			</div>
			{isMermaid && streamedMessage == '' && <MermaidDiagram code={value} />}
		</>
	);
};
CodeBlock.displayName = 'CodeBlock';
