import emojiMartData from '@emoji-mart/data';
import { EmojiInputPlugin, EmojiPlugin } from '@platejs/emoji/react';

import { EmojiInputElement } from '@/components/editor/nodes/emoji_node';

export const EmojiKit = [
	EmojiPlugin.configure({
		options: { data: emojiMartData as any },
	}),
	EmojiInputPlugin.withComponent(EmojiInputElement),
];
