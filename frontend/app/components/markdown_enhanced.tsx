import type { ReactNode } from 'react';
import { memo, useMemo } from 'react';

import 'katex/dist/katex.min.css';
import Markdown from 'react-markdown';
import rehypeKatex from 'rehype-katex';
import rehypeRaw from 'rehype-raw';
import rehypeSanitize, { defaultSchema } from 'rehype-sanitize';
import remarkGemoji from 'remark-gemoji';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import supersub from 'remark-supersub';

import { SanitizeLaTeX } from '@/lib/markdown_utils';
import { CustomMDLanguage } from '@/lib/text_utils';

import { backendAPI } from '@/apis/baseapi';

import CodeBlock from '@/components/markdown_code_block';
import { MdErrorBoundary } from '@/components/markdown_error_boundary';
import ThinkingFence from '@/components/thinking_fence';

const strictSchema = {
	...defaultSchema, // keep ancestors / clobber / prefix logic
	attributes: {
		...defaultSchema.attributes,
		// The `language-*` regex is allowed by default.
		code: [['className', /^language-./, /^math-./]],
		input: defaultSchema.attributes?.input.filter(a => a !== 'value' && a !== 'checked'),
	},
};

interface EnhancedMarkdownProps {
	text: string;
	align?: string;
	isBusy?: boolean;
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

const EnhancedMarkdown = ({ text, align = 'left', isBusy = false }: EnhancedMarkdownProps) => {
	const processedText = useMemo(() => {
		// During a stream skip LaTeX sanitisation for speed.
		return isBusy ? text : SanitizeLaTeX(text);
	}, [text, isBusy]);

	const components = useMemo(
		() => ({
			h1: ({ children }: PComponentProps) => <h1 className="my-2 text-xl font-bold">{children}</h1>,
			h2: ({ children }: PComponentProps) => <h2 className="my-2 text-lg font-bold">{children}</h2>,
			h3: ({ children }: PComponentProps) => <h3 className="my-2 text-base font-bold">{children}</h3>,
			ul: ({ children }: PComponentProps) => <ul className="list-disc py-1 pl-2">{children}</ul>,
			ol: ({ children }: PComponentProps) => <ol className="list-decimal py-1 pl-2">{children}</ol>,
			li: ({ children }: PComponentProps) => <li className="py-1">{children}</li>,
			table: ({ children }: PComponentProps) => <table className="w-full table-auto">{children}</table>,
			thead: ({ children }: PComponentProps) => <thead className="bg-base-300">{children}</thead>,
			tbody: ({ children }: PComponentProps) => <tbody>{children}</tbody>,
			tr: ({ children }: PComponentProps) => <tr className="border-t">{children}</tr>,
			th: ({ children }: PComponentProps) => <th className="px-4 py-2 text-left">{children}</th>,
			td: ({ children }: PComponentProps) => <td className="px-4 py-2">{children}</td>,

			p: ({ className, children }: PComponentProps) => (
				<p className={`${className ?? ''} my-2 ${align} break-words`} style={{ lineHeight: '1.5', fontSize: '14px' }}>
					{children}
				</p>
			),

			blockquote: ({ children }: PComponentProps) => (
				<blockquote className="border-neutral/20 border-l-4 pl-4 italic">{children}</blockquote>
			),

			a: ({ href, children }: { href?: string; children?: ReactNode }) => (
				<a
					href={href}
					target={href?.startsWith('http') ? '_blank' : undefined}
					rel={href?.startsWith('http') ? 'noopener noreferrer' : undefined}
					className="cursor-pointer text-blue-600 underline hover:text-blue-800"
					onClick={e => {
						e.preventDefault();
						if (href) backendAPI.openurl(href);
					}}
				>
					{children}
				</a>
			),

			code: ({ inline, className, children, ...props }: CodeComponentProps) => {
				if (inline || !className) {
					return (
						<code
							{...props}
							className={`bg-base-200 inline text-wrap break-words whitespace-pre-wrap ${className ?? ''}`}
						>
							{children}
						</code>
					);
				}

				const match = /lang-(\w+)/.exec(className || '') || /language-(\w+)/.exec(className || '');
				const language = match && match[1] ? match[1] : 'text';
				// eslint-disable-next-line @typescript-eslint/no-base-to-string
				const value = String(children).replace(/\n$/, '');
				if (language === (CustomMDLanguage.ThinkingSummary as string)) {
					return <ThinkingFence detailsSummaryText="Thinking Summary" text={value} isBusy={isBusy} />;
				} else if (language === (CustomMDLanguage.Thinking as string)) {
					return <ThinkingFence detailsSummaryText="Thinking" text={value} isBusy={isBusy} />;
				} else {
					return <CodeBlock language={language} value={value} isBusy={isBusy} />;
				}
			},
		}),
		[align, isBusy]
	);

	return (
		<MdErrorBoundary source={processedText}>
			<Markdown
				remarkPlugins={[supersub, remarkGemoji, remarkMath, remarkGfm]}
				rehypePlugins={[rehypeRaw, [rehypeSanitize, { ...strictSchema }], rehypeKatex]}
				components={components}
				skipHtml={false}
			>
				{processedText}
			</Markdown>
		</MdErrorBoundary>
	);
};

export default memo(EnhancedMarkdown);
