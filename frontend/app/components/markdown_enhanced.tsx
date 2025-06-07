import type { ReactNode } from 'react';
import { memo, useMemo } from 'react';

import 'katex/dist/katex.min.css';
import Markdown from 'react-markdown';
import rehypeKatex from 'rehype-katex';
import remarkGemoji from 'remark-gemoji';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import supersub from 'remark-supersub';

import { backendAPI } from '@/apis/baseapi';

import { SanitizeLaTeX } from '@/lib/markdown_utils';

import CodeBlock from '@/components/markdown_code_block';

const remarkPlugins = [remarkGemoji, supersub, remarkMath, remarkGfm];
const remarkPluginsStreaming = [remarkGemoji, supersub, remarkGfm];
const rehypePlugins = [rehypeKatex];

interface EnhancedMarkdownProps {
	text: string;
	align?: string;
	isStreaming?: boolean;
}

interface CodeComponentProps {
	inline?: boolean;
	className?: string;
	children?: ReactNode;
}

interface PComponentProps {
	className?: string;
	children?: ReactNode;
}

const EnhancedMarkdown = ({ text, align = 'left', isStreaming = false }: EnhancedMarkdownProps) => {
	const processedText = useMemo(() => {
		// During a stream skip LaTeX sanitisation for speed.
		return isStreaming ? text : SanitizeLaTeX(text);
	}, [text, isStreaming]);

	const components = useMemo(
		() => ({
			h1: ({ children }: PComponentProps) => <h1 className="text-xl font-bold my-2">{children}</h1>,
			h2: ({ children }: PComponentProps) => <h2 className="text-lg font-bold my-2">{children}</h2>,
			h3: ({ children }: PComponentProps) => <h3 className="text-base font-bold my-2">{children}</h3>,

			p: ({ className, children }: PComponentProps) => (
				<p className={`${className ?? ''} my-2 ${align} break-words`} style={{ lineHeight: '1.5', fontSize: '14px' }}>
					{children}
				</p>
			),

			code: ({ inline, className, children, ...props }: CodeComponentProps) => {
				if (inline || !className) {
					return (
						<code
							{...props}
							className={`bg-base-200 inline text-wrap whitespace-pre-wrap break-words ${className ?? ''}`}
						>
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
						isStreaming={isStreaming}
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
						if (href) backendAPI.openurl(href);
					}}
				>
					{children}
				</a>
			),

			blockquote: ({ children }: PComponentProps) => (
				<blockquote className="border-l-4 border-neutral/20 pl-4 italic">{children}</blockquote>
			),
		}),
		[align, isStreaming]
	);

	return (
		<Markdown
			remarkPlugins={isStreaming ? remarkPluginsStreaming : remarkPlugins}
			rehypePlugins={isStreaming ? [] : rehypePlugins}
			components={components}
		>
			{processedText}
		</Markdown>
	);
};

export default memo(EnhancedMarkdown);
