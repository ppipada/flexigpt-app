import type { ConversationMessage } from '@/spec/conversation';

import { CustomMDLanguage } from '@/lib/text_utils';

// Builds final content string: reasoning blocks (summary + thinking) + original content.
// - Uses `reasoningContents` if present.
// - Ignores redactedThinking.
export function buildEffectiveContentWithReasoning(message: ConversationMessage): string {
	const baseContent = message.uiContent;
	const reasoningContents = message.uiReasoningContents ?? [];

	if (reasoningContents.length === 0) {
		return baseContent;
	}

	const summaryParts: string[] = [];
	const thinkingParts: string[] = [];

	for (const rc of reasoningContents) {
		if (Array.isArray(rc.summary) && rc.summary.length > 0) {
			// rc.summary: string[]
			summaryParts.push(rc.summary.join('\n'));
		}

		if (Array.isArray(rc.thinking) && rc.thinking.length > 0) {
			// rc.thinking: string[]
			thinkingParts.push(rc.thinking.join('\n'));
		}

		// We intentionally ignore rc.redactedThinking and rc.encryptedContent for display.
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
