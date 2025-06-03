import { type FC } from 'react';

import 'katex/dist/katex.min.css';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { monokaiSublime } from 'react-syntax-highlighter/dist/cjs/styles/hljs';

import CopyButton from '@/components/copy_button';
import DownloadButton from '@/components/download_button';
import { MermaidDiagram } from '@/components/mermaid_diagram';

interface CodeProps {
	language: string;
	value: string;
	isStreaming: boolean;
}

const CodeBlock: FC<CodeProps> = ({ language, value, isStreaming }) => {
	const fetchValue = async (): Promise<string> => {
		return value;
	};
	const isMermaid = language.toLowerCase() === 'mermaid';
	return (
		<>
			<div className="rounded-lg bg-gray-800 my-4 items-start overflow-hidden">
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
						wrapLines={true}
						lineNumberStyle={{ userSelect: 'none', pointerEvents: 'none', WebkitUserSelect: 'none' }}
						customStyle={{
							background: 'transparent',
							padding: '0.5em',
							borderRadius: '0.25rem',
							fontSize: '14px',
							// overflow: 'auto',
							// position: 'relative',
							// userSelect: 'text',
						}}
					>
						{value}
					</SyntaxHighlighter>
				</div>
			</div>
			{isMermaid && !isStreaming && <MermaidDiagram code={value} />}
		</>
	);
};

export default CodeBlock;
