import type { FC } from 'react';

import { useHighlight } from '@/hooks/use_highlight';

import CopyButton from '@/components/copy_button';
import DownloadButton from '@/components/download_button';
import MermaidDiagram from '@/components/markdown_mermaid_diagram';

interface CodeProps {
	language: string;
	value: string;
	isBusy: boolean;
}

const CodeBlock: FC<CodeProps> = ({ language, value, isBusy }) => {
	const html = useHighlight(value, language);
	const showFallback = !value.trim() || html === null || html === '';
	const isMermaid = language.toLowerCase() === 'mermaid';

	const fallback = (
		<pre className="text-code overflow-auto rounded bg-transparent p-2 text-sm">
			<code>{value}</code>
		</pre>
	);

	const fetchValue = async () => value;

	return (
		<>
			<div className="bg-code my-4 items-start overflow-hidden rounded-lg">
				<div className="bg-code-header flex items-center justify-between px-4">
					<span className="text-code text-sm capitalize">{language}</span>
					<div className="flex space-x-2">
						<DownloadButton
							language={language}
							valueFetcher={fetchValue}
							size={16}
							className="btn btn-sm text-code flex items-center border-none bg-transparent shadow-none"
						/>
						<CopyButton
							value={value}
							className="btn btn-sm text-code flex items-center border-none bg-transparent shadow-none"
							size={16}
						/>
					</div>
				</div>

				<div className="text-code p-1" style={{ fontSize: 14, lineHeight: 1.5 }}>
					{showFallback ? fallback : <div className="shiki-container" dangerouslySetInnerHTML={{ __html: html }} />}
				</div>
			</div>

			{isMermaid && !isBusy && <MermaidDiagram code={value} />}
		</>
	);
};

export default CodeBlock;
