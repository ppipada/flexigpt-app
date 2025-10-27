import { IndentPlugin } from '@platejs/indent/react';
import { ListPlugin } from '@platejs/list/react';
import { KEYS } from 'platejs';

import { BlockList } from '@/components/editor/nodes/block_list';

export const ListKit = [
	IndentPlugin.configure({
		inject: {
			targetPlugins: [...KEYS.heading, KEYS.p, KEYS.blockquote, KEYS.codeBlock, KEYS.toggle, KEYS.img],
		},
	}),
	ListPlugin.configure({
		inject: {
			targetPlugins: [...KEYS.heading, KEYS.p, KEYS.blockquote, KEYS.codeBlock, KEYS.toggle],
		},
		render: {
			belowNodes: BlockList,
		},
	}),
];
