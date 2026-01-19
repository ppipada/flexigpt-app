import {
	BoldPlugin,
	CodePlugin,
	HighlightPlugin,
	ItalicPlugin,
	KbdPlugin,
	StrikethroughPlugin,
	SubscriptPlugin,
	SuperscriptPlugin,
	UnderlinePlugin,
} from '@platejs/basic-nodes/react';

import { CodeLeaf } from '@/components/editor/nodes/code_node';
import { HighlightLeaf } from '@/components/editor/nodes/highlight_node';
import { KbdLeaf } from '@/components/editor/nodes/kbd_node';

export const BasicMarksKit = [
	BoldPlugin,
	CodePlugin.configure({
		node: { component: CodeLeaf },
		shortcuts: { toggle: { keys: 'mod+e' } },
	}),
	ItalicPlugin,
	UnderlinePlugin,
	StrikethroughPlugin.configure({
		shortcuts: { toggle: { keys: 'mod+shift+x' } },
	}),
	SubscriptPlugin,
	SuperscriptPlugin,
	HighlightPlugin.configure({
		node: { component: HighlightLeaf },
		shortcuts: { toggle: { keys: 'mod+shift+h' } },
	}),
	KbdPlugin.withComponent(KbdLeaf),
];
