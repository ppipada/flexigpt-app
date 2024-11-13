import CopyButton from '@/components/copy_button';
import DownloadButton from '@/components/download_button';
import 'katex/dist/katex.min.css';
import mermaid from 'mermaid';
import { FC, ReactNode, memo, useEffect, useRef } from 'react';
import Markdown from 'react-markdown';
import SyntaxHighlighter from 'react-syntax-highlighter';
import { monokaiSublime } from 'react-syntax-highlighter/dist/esm/styles/hljs';
import rehypeKatex from 'rehype-katex';
import rehypeReact from 'rehype-react';
import remarkGemoji from 'remark-gemoji';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import remarkParse from 'remark-parse';
import remarkRehype from 'remark-rehype';
import supersub from 'remark-supersub';
import { v4 as uuidv4 } from 'uuid';

// LaTeX processing function
const containsLatexRegex = /\\\(.*?\\\)|\\\[.*?\\\]|\$.*?\$|\\begin\{equation\}.*?\\end\{equation\}/;
const inlineLatex = new RegExp(/\\\((.+?)\\\)/, 'g');
const blockLatex = new RegExp(/\\\[(.*?[^\\])\\\]/, 'gs');

export const processLaTeX = (content: string) => {
	let processedContent = content.replace(/(\$)(?=\s?\d)/g, '\\$');

	if (!containsLatexRegex.test(processedContent)) {
		return processedContent;
	}

	processedContent = processedContent
		.replace(inlineLatex, (match: string, equation: string) => `$${equation}$`)
		.replace(blockLatex, (match: string, equation: string) => `$$${equation}$$`);

	return processedContent;
};

export const MemoizedMarkdown = memo(
	Markdown,
	(prevProps, nextProps) => prevProps.children === nextProps.children && prevProps.className === nextProps.className
);

interface MermaidDiagramProps {
	code: string;
}

const MermaidDiagram: FC<MermaidDiagramProps> = ({ code }) => {
	const containerRef = useRef<HTMLDivElement | null>(null);
	const uniqueId = useRef(`mermaid-${uuidv4()}`); // Generate a unique ID

	useEffect(() => {
		if (containerRef.current) {
			mermaid.initialize({
				startOnLoad: true,
				theme: 'default',
			});
			mermaid
				.render(uniqueId.current, code)
				.then(renderResult => {
					if (containerRef.current) {
						containerRef.current.innerHTML = renderResult.svg;
					}
				})
				.catch(error => {
					console.error('Mermaid rendering failed:', error);
				});
		}
	}, [code]);

	const fetchDiagramAsBlob = async (): Promise<Blob> => {
		if (containerRef.current) {
			const svg = containerRef.current.querySelector('svg');
			if (svg) {
				const svgData = new XMLSerializer().serializeToString(svg);

				// Convert SVG to PNG using a canvas
				return new Promise<Blob>((resolve, reject) => {
					const canvas = document.createElement('canvas');
					const ctx = canvas.getContext('2d');
					const img = new Image();

					img.onload = () => {
						// Set the desired resolution scale factor
						const scaleFactor = 2; // Adjust this value for higher/lower resolution

						// Set canvas dimensions based on the scale factor
						canvas.width = img.width * scaleFactor;
						canvas.height = img.height * scaleFactor;

						if (ctx) {
							// Scale the drawing context
							ctx.scale(scaleFactor, scaleFactor);

							// Fill the background with white
							ctx.fillStyle = 'white';
							ctx.fillRect(0, 0, canvas.width, canvas.height);

							// Draw the image onto the canvas
							ctx.drawImage(img, 0, 0);
						}

						// Convert the canvas to a Blob
						canvas.toBlob(
							blob => {
								if (blob) {
									resolve(blob);
								} else {
									reject(new Error('Canvas is empty'));
								}
							},
							'image/png',
							1.0
						);
					};

					img.onerror = err => {
						reject(err);
					};

					// Set the source of the image to the SVG data URL
					img.src = 'data:image/svg+xml;base64,' + btoa(unescape(encodeURIComponent(svgData)));
				});
			} else {
				throw new Error('SVG element not found in container');
			}
		} else {
			throw new Error('Container not found');
		}
	};

	return (
		<div className="rounded-md my-4 items-start overflow-hidden" style={{ backgroundColor: '#E5E9F0' }}>
			<div className="flex justify-between items-center bg-gray-700 px-4">
				<span className="text-white">Mermaid Diagram</span>
				<DownloadButton
					valueFetcher={fetchDiagramAsBlob}
					size={16}
					fileprefix="diagram"
					isBinary={true} // Important for binary content
					language="mermaid"
					className="btn btn-sm bg-transparent text-white border-none flex items-center shadow-none"
				/>
			</div>
			<div className="flex justify-center text-center p-1" ref={containerRef}></div>
		</div>
	);
};

interface CodeProps {
	language: string;
	value: string;
	isStreaming: boolean;
}

const CodeBlock: FC<CodeProps> = memo(({ language, value, isStreaming }) => {
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
			{isMermaid && !isStreaming && <MermaidDiagram code={value} />}
		</>
	);
});
CodeBlock.displayName = 'CodeBlock';

interface CodeComponentProps {
	inline?: boolean;
	className?: string;
	children?: ReactNode;
}

interface PComponentProps {
	className?: string;
	children?: ReactNode;
}

export interface ChatMessageContentProps {
	content: string;
	align: string;
	isStreaming: boolean;
}

export function ChatMessageContent({ content, align, isStreaming }: ChatMessageContentProps) {
	// Process the content to handle LaTeX expressions
	const processedContent = processLaTeX(content);

	const components = {
		h1: ({ children }: PComponentProps) => <h1 className="text-xl font-bold my-2">{children}</h1>,
		h2: ({ children }: PComponentProps) => <h2 className="text-lg font-bold my-2">{children}</h2>,
		h3: ({ children }: PComponentProps) => <h3 className="text-base font-bold my-2">{children}</h3>,
		p: ({ className, children }: PComponentProps) => {
			const newClassName = `${className} my-2 ${align}`;
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
					value={String(children).replace(/\n$/, '')}
					isStreaming={isStreaming}
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
				className="underline text-blue-600 hover:text-blue-800"
			>
				{children}
			</a>
		),
		blockquote: ({ children }: PComponentProps) => (
			<blockquote className="border-l-4 border-gray-300 pl-4 italic">{children}</blockquote>
		),
	};

	return (
		<div className="bg-base-100 rounded-2xl shadow-lg px-4 py-2">
			<MemoizedMarkdown
				remarkPlugins={[remarkParse, remarkGemoji, supersub, remarkGfm, remarkMath, remarkRehype]}
				rehypePlugins={[rehypeKatex, rehypeReact]}
				components={components}
			>
				{processedContent}
			</MemoizedMarkdown>
		</div>
	);
}
