import { memo, useMemo } from 'react';

// <-- adjust path
import { useDebounced } from '@/lib/debounce';

import EnhancedMarkdown from '@/components/markdown_enhanced';

/* ------------------------------------------------------------------ */
/*  props                                                             */
/* ------------------------------------------------------------------ */
interface ChatMessageContentProps {
	content: string; // final text
	streamedText?: string; // partial text while streaming
	isStreaming?: boolean;
	align: string;
	renderAsMarkdown?: boolean;
}

/* ------------------------------------------------------------------ */
/*  component                                                         */
/* ------------------------------------------------------------------ */
const ChatMessageContent = ({
	content,
	streamedText = '',
	isStreaming = false,
	align,
	renderAsMarkdown = true,
}: ChatMessageContentProps) => {
	/* ----------------------- pick the live text ---------------------- */
	const liveText = isStreaming ? streamedText : content;
	const textToRender = useDebounced(liveText, 250); // max ~4×/sec

	/* ---------------------- plain-text branch ------------------------ */
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

	/* ---------------------- markdown branch -------------------------- */
	return (
		<div className="bg-base-100 px-4 py-2">
			<EnhancedMarkdown text={textToRender} align={align} isStreaming={isStreaming} />
		</div>
	);
};

/* ------------------------------------------------------------------ */
/*  memo                                                              */
/* ------------------------------------------------------------------ */
function areEqual(prev: ChatMessageContentProps, next: ChatMessageContentProps) {
	return (
		prev.content === next.content &&
		prev.streamedText === next.streamedText &&
		prev.isStreaming === next.isStreaming &&
		prev.align === next.align &&
		prev.renderAsMarkdown === next.renderAsMarkdown
	);
}

export default memo(ChatMessageContent, areEqual);
