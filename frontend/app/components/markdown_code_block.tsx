import type { FC } from 'react';
import { useMemo } from 'react';

import { GetSupportedLanguage } from '@/lib/markdown_utils';
import { highlighter } from '@/lib/syntax_highlighter';

import CopyButton from '@/components/copy_button';
import DownloadButton from '@/components/download_button';
import MermaidDiagram from '@/components/markdown_mermaid_diagram';

interface CodeProps {
	language: string;
	value: string;
	isStreaming: boolean;
}

const CodeBlock: FC<CodeProps> = ({ language, value, isStreaming }) => {
	const highlightedCode = useMemo(() => {
		if (!value.trim()) {
			return '';
		}

		try {
			return highlighter.codeToHtml(value, {
				lang: GetSupportedLanguage(language),
				theme: 'monokai',
			});
		} catch (error) {
			console.error('Error highlighting code:', error);
			// Fallback to plain code block
			return `<pre class="bg-gray-800 text-gray-100 p-4 rounded overflow-auto"><code>${value}</code></pre>`;
		}
	}, [value, language]);

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
					{highlightedCode ? (
						<div
							className="shiki-container"
							style={{
								fontSize: '14px',
								lineHeight: '1.5',
								overflow: 'auto',
							}}
							dangerouslySetInnerHTML={{ __html: highlightedCode }}
						/>
					) : (
						<pre className="bg-transparent text-gray-100 p-2 rounded overflow-auto text-sm">
							<code>{value}</code>
						</pre>
					)}
				</div>
			</div>
			{isMermaid && !isStreaming && <MermaidDiagram code={value} />}
		</>
	);
};

export default CodeBlock;
