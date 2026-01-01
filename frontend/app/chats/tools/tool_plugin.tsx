import { createSlatePlugin } from 'platejs';

import { KEY_TOOL_SELECTION } from '@/chats/tools/tool_editor_utils';
import { ToolSelectionElement } from '@/chats/tools/tool_plus_selection';

const ToolSelectionPlugin = createSlatePlugin({
	key: KEY_TOOL_SELECTION,
	node: { isElement: true, isInline: true, isVoid: true, isSelectable: false },
	editOnly: true,
});

export const ToolPlusKit = [
	// Keep only the hidden selection node that powers chips & data:
	ToolSelectionPlugin.withComponent(ToolSelectionElement),
];
