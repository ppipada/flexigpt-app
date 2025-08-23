import { ListPlugin } from '@platejs/list/react';
import { KEYS } from 'platejs';

import { BlockList } from '@/components/editor/nodes/block_list';
import { IndentKit } from '@/components/editor/plugins/indent_kit';

export const ListKit = [
	...IndentKit,
	ListPlugin.configure({
		inject: {
			targetPlugins: [...KEYS.heading, KEYS.p, KEYS.blockquote, KEYS.toggle],
		},
		render: {
			belowNodes: BlockList,
		},
	}),
];
