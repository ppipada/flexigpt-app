/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import type { ConversationMessage } from '@/spec/conversation';
import { ReasoningContentType } from '@/spec/modelpreset';

import { CustomMDLanguage } from '@/lib/text_utils';

// Builds final content string: reasoning blocks (summary + thinking) + original content.
// - Uses `reasoningContents` if present.
// - Ignores redactedThinking.
export function buildEffectiveContentWithReasoning(message: ConversationMessage): string {
	const baseContent = message.content ?? '';
	const reasoningContents = message.reasoningContents;

	if (!reasoningContents || reasoningContents.length === 0) {
		return baseContent;
	}

	const summaryParts: string[] = [];
	const thinkingParts: string[] = [];

	for (const rc of reasoningContents) {
		if (rc.type === ReasoningContentType.ReasoningOpenAIResponses && rc.contentOpenAIResponses) {
			const { summary, content } = rc.contentOpenAIResponses;

			if (Array.isArray(summary) && summary.length > 0) {
				// summary is string[]
				summaryParts.push(summary.join('\n'));
			}

			if (Array.isArray(content) && content.length > 0) {
				// content is string[]
				thinkingParts.push(content.join('\n'));
			}
		} else if (rc.type === ReasoningContentType.ReasoningAnthropicMessages && rc.contentAnthropicMessages) {
			const { thinking } = rc.contentAnthropicMessages;

			if (thinking) {
				// Ignore redactedThinking on purpose.
				thinkingParts.push(thinking);
			}
		}
	}

	// If we still have nothing, just return the base content.
	if (!summaryParts.length && !thinkingParts.length) {
		return baseContent;
	}

	let reasoningText = '';

	if (summaryParts.length) {
		const summaryText = summaryParts.join('\n\n');
		reasoningText += `\n~~~${CustomMDLanguage.ThinkingSummary}\n${summaryText}\n~~~\n`;
	}

	if (thinkingParts.length) {
		const thinkingText = thinkingParts.join('\n\n');
		reasoningText += `\n~~~${CustomMDLanguage.Thinking}\n${thinkingText}\n~~~\n`;
	}

	// If the message has no visible content, just return the reasoning blocks.
	if (!baseContent.trim()) {
		return reasoningText.trimStart();
	}

	// Otherwise: reasoning (summary + thinking) followed by the normal content.
	return `${reasoningText}\n${baseContent}\n`;
}
