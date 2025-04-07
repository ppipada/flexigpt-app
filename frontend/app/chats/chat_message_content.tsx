import { backendAPI } from '@/apis/baseapi';
import CodeBlock from '@/chats/chat_message_content_codeblock';
import 'katex/dist/katex.min.css';
import type { ReactNode } from 'react';
import { useMemo } from 'react';
import Markdown from 'react-markdown';
import rehypeKatex from 'rehype-katex';
import remarkGemoji from 'remark-gemoji';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import supersub from 'remark-supersub';

// LaTeX processing function
const containsLatexRegex = /\\\(.*?\\\)|\\\[.*?\\\]|\$.*?\$|\\begin\{equation\}.*?\\end\{equation\}/;
const inlineLatex = new RegExp(/\\\((.+?)\\\)/, 'g');
const blockLatex = new RegExp(/\\\[(.*?[^\\])\\\]/, 'gs');

const processLaTeX = (content: string) => {
	let processedContent = content.replace(/(\$)(?=\s?\d)/g, '\\$');

	if (!containsLatexRegex.test(processedContent)) {
		return processedContent;
	}

	processedContent = processedContent
		.replace(inlineLatex, (match: string, equation: string) => `$${equation}$`)
		.replace(blockLatex, (match: string, equation: string) => `$$${equation}$$`);

	return processedContent;
};

interface CodeComponentProps {
	inline?: boolean;
	className?: string;
	children?: ReactNode;
}

interface PComponentProps {
	className?: string;
	children?: ReactNode;
}

interface ChatMessageContentProps {
	content: string;
	align: string;
	streamedMessage: string;
	renderAsMarkdown?: boolean;
}

// const MemoizedMarkdown = memo(
// 	Markdown,
// 	(prevProps, nextProps) => prevProps.children === nextProps.children && prevProps.className === nextProps.className
// );

function ChatMessageMarkdownContent({ content, align, streamedMessage }: ChatMessageContentProps) {
	// Process the content to handle LaTeX expressions
	const processedContent = processLaTeX(content);

	const components = useMemo(
		() => ({
			h1: ({ children }: PComponentProps) => <h1 className="text-xl font-bold my-2">{children}</h1>,
			h2: ({ children }: PComponentProps) => <h2 className="text-lg font-bold my-2">{children}</h2>,
			h3: ({ children }: PComponentProps) => <h3 className="text-base font-bold my-2">{children}</h3>,
			p: ({ className, children }: PComponentProps) => {
				const newClassName = `${className || ''} my-2 ${align} break-words`;
				return (
					<p className={newClassName} style={{ lineHeight: '1.5', fontSize: '14px' }}>
						{children}
					</p>
				);
			},
			code: ({ inline, className, children, ...props }: CodeComponentProps) => {
				if (inline || !className) {
					const newClassName = `bg-base-200 inline text-wrap whitespace-pre-wrap break-words ${className}`;
					return (
						<code className={newClassName} {...props}>
							{children}
						</code>
					);
				}
				const match = /lang-(\w+)/.exec(className || '') || /language-(\w+)/.exec(className || '');
				const language = match && match[1] ? match[1] : 'text';

				return (
					<CodeBlock
						language={language}
						// eslint-disable-next-line @typescript-eslint/no-base-to-string
						value={String(children).replace(/\n$/, '')}
						streamedMessage={streamedMessage}
						{...props}
					/>
				);
			},
			ul: ({ children }: PComponentProps) => (
				<span>
					<ul className="list-disc py-1">{children}</ul>
				</span>
			),
			ol: ({ children }: PComponentProps) => (
				<span>
					<ol className="list-decimal py-1">{children}</ol>
				</span>
			),
			li: ({ children }: PComponentProps) => (
				<span>
					<li className="ml-4 py-1">{children}</li>
				</span>
			),
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
					className="underline text-blue-600 hover:text-blue-800 cursor-pointer"
					onClick={e => {
						e.preventDefault();
						if (href) {
							backendAPI.openurl(href);
						}
					}}
				>
					{children}
				</a>
			),
			blockquote: ({ children }: PComponentProps) => (
				<blockquote className="border-l-4 border-gray-300 pl-4 italic">{children}</blockquote>
			),
		}),
		[align, streamedMessage]
	);

	return (
		<div className="bg-base-100 px-4 py-2">
			<Markdown
				remarkPlugins={[remarkGemoji, supersub, remarkGfm, remarkMath]}
				rehypePlugins={[rehypeKatex]}
				components={components}
			>
				{processedContent}
			</Markdown>
		</div>
	);
}

export default function ChatMessageContent({
	content,
	align,
	streamedMessage,
	renderAsMarkdown = true,
}: ChatMessageContentProps) {
	if (renderAsMarkdown) {
		return <ChatMessageMarkdownContent content={content} align={align} streamedMessage={streamedMessage} />;
	}

	// Memoize the plain text content to prevent unnecessary re-renders
	const plainTextContent = useMemo(() => {
		return content.split('\n').map((line, index) => (
			<p
				key={index}
				className={`${align} break-words`}
				style={{ whiteSpace: 'pre-wrap', lineHeight: '1.5', fontSize: '14px' }}
			>
				{line || '\u00A0' /* Use non-breaking space for empty lines */}
			</p>
		));
	}, [content, align]);

	return <div className="bg-base-100 px-4 py-2">{plainTextContent}</div>;
}
