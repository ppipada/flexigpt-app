import { memo, useMemo } from 'react';

// <-- adjust path
import { useDebounced } from '@/lib/debounce';

import EnhancedMarkdown from '@/components/markdown_enhanced';

interface ChatMessageContentProps {
	// Final text
	content: string;
	// Partial text while streaming.
	streamedText?: string;
	isStreaming?: boolean;
	isPending?: boolean;
	align: string;
	renderAsMarkdown?: boolean;
}

const ChatMessageContent = ({
	content,
	streamedText = '',
	isStreaming = false,
	isPending = false,
	align,
	renderAsMarkdown = true,
}: ChatMessageContentProps) => {
	const liveText = isStreaming ? streamedText : content;
	// Max ~4×/sec.
	const textToRender = useDebounced(liveText, 250);

	if (isPending && textToRender.trim() === '') {
		return (
			<div className="bg-base-100 px-4 py-2 flex items-center">
				Thinking
				<span className="ml-4 loading loading-dots loading-md" />
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

		return <div className="bg-base-100 px-4 py-2">{plainText}</div>;
	}
	return (
		<div className="bg-base-100 px-4 py-2">
			<EnhancedMarkdown text={textToRender} align={align} isStreaming={isStreaming} />
		</div>
	);
};

function areEqual(prev: ChatMessageContentProps, next: ChatMessageContentProps) {
	return (
		prev.content === next.content &&
		prev.streamedText === next.streamedText &&
		prev.isStreaming === next.isStreaming &&
		prev.isPending === next.isPending &&
		prev.align === next.align &&
		prev.renderAsMarkdown === next.renderAsMarkdown
	);
}

export default memo(ChatMessageContent, areEqual);
