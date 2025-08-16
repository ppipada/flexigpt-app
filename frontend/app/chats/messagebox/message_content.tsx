import { memo, useMemo } from 'react';

import { useDebounce } from '@/hooks/use_debounce';

import EnhancedMarkdown from '@/components/markdown_enhanced';

interface MessageContentProps {
	messageID: string;
	// Final text
	content: string;
	// Partial text while streaming.
	streamedText?: string;
	isStreaming?: boolean;
	isPending?: boolean;
	align: string;
	renderAsMarkdown?: boolean;
}

const MessageContent = ({
	messageID,
	content,
	streamedText = '',
	isStreaming = false,
	isPending = false,
	align,
	renderAsMarkdown = true,
}: MessageContentProps) => {
	const liveText = isStreaming ? streamedText : content;
	// Max ~4×/sec.
	const textToRender = useDebounce(liveText, 128);

	if (isPending && textToRender.trim() === '') {
		return (
			<div className="px-4 py-2 flex items-center">
				Thinking
				<span className="ml-4 loading loading-dots loading-sm" />
			</div>
		);
	}

	if (!renderAsMarkdown) {
		const plainText = useMemo(
			() =>
				textToRender.split('\n').map((line, idx) => (
					<p
						key={idx}
						className={`${align} break-words`}
						style={{ whiteSpace: 'pre-wrap', lineHeight: '1.5', fontSize: '14px' }}
					>
						{line || '\u00A0' /* Use non-breaking space for empty lines */}
					</p>
				)),
			[textToRender, align]
		);

		return <div className="px-4 py-2">{plainText}</div>;
	}
	return (
		<div className="px-4 py-2">
			<EnhancedMarkdown
				key={`${messageID}:${isStreaming ? 'live' : 'done'}`}
				text={textToRender}
				align={align}
				isStreaming={isStreaming}
			/>
		</div>
	);
};

function areEqual(prev: MessageContentProps, next: MessageContentProps) {
	return (
		prev.content === next.content &&
		prev.streamedText === next.streamedText &&
		prev.isStreaming === next.isStreaming &&
		prev.isPending === next.isPending &&
		prev.align === next.align &&
		prev.renderAsMarkdown === next.renderAsMarkdown
	);
}

export default memo(MessageContent, areEqual);
