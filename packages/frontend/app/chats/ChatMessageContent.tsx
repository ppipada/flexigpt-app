import CopyButton from '@/components/CopyButton';
import DownloadButton from '@/components/DownloadButton';
import { FC, ReactNode, memo } from 'react';
import Markdown from 'react-markdown';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { monokaiSublime } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import rehypeReact from 'rehype-react';
import remarkGfm from 'remark-gfm';
import remarkParse from 'remark-parse';
import remarkRehype from 'remark-rehype';

export const MemoizedMarkdown = memo(
	Markdown,
	(prevProps, nextProps) => prevProps.children === nextProps.children && prevProps.className === nextProps.className
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
	const fetchValue = async (): Promise<string> => {
		return value;
	};
	return (
		<div className="rounded-md bg-gray-800 my-2 items-start overflow-hidden">
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
	);
});
CodeBlock.displayName = 'CodeBlock';

interface CodeComponentProps {
	inline?: boolean;
	className?: string;
	children?: ReactNode;
}

interface PComponentProps {
	children?: ReactNode;
}

export function ChatMessageContent({ content, align }: ChatMessageContentProps) {
	const components = {
		h1: ({ children }: PComponentProps) => <h1 className="text-2xl font-bold my-2">{children}</h1>,
		h2: ({ children }: PComponentProps) => <h2 className="text-xl font-bold my-2">{children}</h2>,
		h3: ({ children }: PComponentProps) => <h3 className="text-lg font-bold my-2">{children}</h3>,
		p: ({ children }: PComponentProps) => (
			<p className={`my-1 ${align}`} style={{ lineHeight: '1.5', fontSize: '14px' }}>
				{children}
			</p>
		),
		code: ({ inline, className, children, ...props }: CodeComponentProps) => {
			if (inline || !className) {
				return (
					<code className="bg-base-200 whitespace-nowrap inline" {...props}>
						{children}
					</code>
				);
			}
			const match = /lang-(\w+)/.exec(className || '') || /language-(\w+)/.exec(className || '');
			const language = match && match[1] ? match[1] : 'text';

			return <CodeBlock language={language} value={String(children).replace(/\n$/, '')} {...props} />;
		},
		ul: ({ children }: PComponentProps) => <ul className="list-disc list-inside">{children}</ul>,
		ol: ({ children }: PComponentProps) => <ol className="list-decimal list-inside">{children}</ol>,
		li: ({ children }: PComponentProps) => <li className="ml-4">{children}</li>,
		table: ({ children }: PComponentProps) => <table className="table-auto w-full">{children}</table>,
		thead: ({ children }: PComponentProps) => <thead className="bg-base-300">{children}</thead>,
		tbody: ({ children }: PComponentProps) => <tbody>{children}</tbody>,
		tr: ({ children }: PComponentProps) => <tr className="border-t">{children}</tr>,
		th: ({ children }: PComponentProps) => <th className="px-4 py-2 text-left">{children}</th>,
		td: ({ children }: PComponentProps) => <td className="px-4 py-2">{children}</td>,
		a: ({ href, children }: { href?: string; children?: ReactNode }) => (
			<a
				href={href}
				target={href?.startsWith('http') ? '_blank' : undefined}
				rel={href?.startsWith('http') ? 'noopener noreferrer' : undefined}
				className="underline text-blue-600 hover:text-blue-800"
			>
				{children}
			</a>
		),

		blockquote: ({ children }: PComponentProps) => (
			<blockquote className="border-l-4 border-gray-300 pl-4 italic">{children}</blockquote>
		),
		del: ({ children }: PComponentProps) => <del className="line-through">{children}</del>,
	};

	return (
		<div className="bg-base-100 rounded-2xl shadow-lg px-4 py-2">
			<MemoizedMarkdown
				remarkPlugins={[remarkParse, remarkGfm, remarkRehype]}
				rehypePlugins={[rehypeReact]}
				components={components}
			>
				{content}
			</MemoizedMarkdown>
		</div>
	);
}
