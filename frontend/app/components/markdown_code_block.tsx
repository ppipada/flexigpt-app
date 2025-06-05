import type { FC } from 'react';

import { useHighlight } from '@/lib/highlight_hook';

import CopyButton from '@/components/copy_button';
import DownloadButton from '@/components/download_button';
import MermaidDiagram from '@/components/markdown_mermaid_diagram';

interface CodeProps {
	language: string;
	value: string;
	isStreaming: boolean;
}

const CodeBlock: FC<CodeProps> = ({ language, value, isStreaming }) => {
	const html = useHighlight(value, language);
	const showFallback = !value.trim() || html === null || html === '';
	const isMermaid = language.toLowerCase() === 'mermaid';

	const fallback = (
		<pre className="bg-transparent text-gray-100 p-2 rounded overflow-auto text-sm">
			<code>{value}</code>
		</pre>
	);

	const fetchValue = async () => value;

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

				<div className="p-1" style={{ fontSize: 14, lineHeight: 1.5 }}>
					{showFallback ? fallback : <div className="shiki-container" dangerouslySetInnerHTML={{ __html: html }} />}
				</div>
			</div>

			{isMermaid && !isStreaming && <MermaidDiagram code={value} />}
		</>
	);
};

export default CodeBlock;
