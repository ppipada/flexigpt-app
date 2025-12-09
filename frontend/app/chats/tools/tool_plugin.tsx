import { createSlatePlugin } from 'platejs';

import { KEY_TOOL_SELECTION } from '@/chats/tools/tool_editor_utils';
import { ToolSelectionElement } from '@/chats/tools/tool_plus_selection';

// type ToolPlusConfig = PluginConfig<typeof KEY_TOOL_PLUS_COMMAND, TriggerComboboxPluginOptions>;

// const ToolPlusInputPlugin = createSlatePlugin({
// 	key: KEY_TOOL_PLUS_INPUT,
// 	editOnly: true,
// 	node: { isElement: true, isInline: true, isVoid: true },
// });

// const ToolPlusPlugin = createTSlatePlugin<ToolPlusConfig>({
// 	key: KEY_TOOL_PLUS_COMMAND,
// 	editOnly: true,
// 	options: {
// 		trigger: '+',
// 		triggerPreviousCharPattern: /^\s?$/,
// 		createComboboxInput: () => ({
// 			children: [{ text: '' }],
// 			type: KEY_TOOL_PLUS_INPUT,
// 		}),
// 		triggerQuery: editor =>
// 			!editor.api.some({
// 				match: { type: editor.getType(KEYS.codeBlock) },
// 			}),
// 	},
// 	plugins: [ToolPlusInputPlugin],
// }).overrideEditor(withTriggerCombobox);

const ToolSelectionPlugin = createSlatePlugin({
	key: KEY_TOOL_SELECTION,
	node: { isElement: true, isInline: true, isVoid: true, isSelectable: false },
	editOnly: true,
});

export const ToolPlusKit = [
	// Keep only the hidden selection node that powers chips & data:
	ToolSelectionPlugin.withComponent(ToolSelectionElement),

	// To re-enable "+" inline menu, add back:
	// ToolPlusPlugin.configure({}),
	// ToolPlusInputPlugin.withComponent(ToolPlusInputElement),
];
