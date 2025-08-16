import {
	BoldPlugin,
	ItalicPlugin,
	KbdPlugin,
	StrikethroughPlugin,
	SubscriptPlugin,
	SuperscriptPlugin,
	UnderlinePlugin,
} from '@platejs/basic-nodes/react';

import { KbdLeaf } from '@/components/editor/nodes/kbd_node';

export const BasicMarksKit = [
	BoldPlugin,
	ItalicPlugin,
	UnderlinePlugin,
	StrikethroughPlugin.configure({
		shortcuts: { toggle: { keys: 'mod+shift+x' } },
	}),
	SubscriptPlugin.configure({
		shortcuts: { toggle: { keys: 'mod+comma' } },
	}),
	SuperscriptPlugin.configure({
		shortcuts: { toggle: { keys: 'mod+period' } },
	}),
	KbdPlugin.withComponent(KbdLeaf),
];
