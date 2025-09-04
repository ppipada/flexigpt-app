import { type TriggerComboboxPluginOptions, withTriggerCombobox } from '@platejs/combobox';
import { createSlatePlugin, createTSlatePlugin, KEYS, type PluginConfig } from 'platejs';

import { KEY_TOOL_PLUS_COMMAND, KEY_TOOL_PLUS_INPUT, KEY_TOOL_SELECTION } from '@/chats/attachments/tool_editor_utils';
import { ToolPlusInputElement } from '@/chats/attachments/tool_plus_input';
import { ToolSelectionElement } from '@/chats/attachments/tool_plus_selection';

type ToolPlusConfig = PluginConfig<typeof KEY_TOOL_PLUS_COMMAND, TriggerComboboxPluginOptions>;

const ToolPlusInputPlugin = createSlatePlugin({
	key: KEY_TOOL_PLUS_INPUT,
	editOnly: true,
	node: { isElement: true, isInline: true, isVoid: true },
});

const ToolPlusPlugin = createTSlatePlugin<ToolPlusConfig>({
	key: KEY_TOOL_PLUS_COMMAND,
	editOnly: true,
	options: {
		trigger: '+',
		triggerPreviousCharPattern: /^\s?$/,
		createComboboxInput: () => ({
			children: [{ text: '' }],
			type: KEY_TOOL_PLUS_INPUT,
		}),
		triggerQuery: editor =>
			!editor.api.some({
				match: { type: editor.getType(KEYS.codeBlock) },
			}),
	},
	plugins: [ToolPlusInputPlugin],
}).overrideEditor(withTriggerCombobox);

const ToolSelectionPlugin = createSlatePlugin({
	key: KEY_TOOL_SELECTION,
	node: { isElement: true, isInline: true, isVoid: true },
	editOnly: true,
});

export const ToolPlusKit = [
	ToolSelectionPlugin.withComponent(ToolSelectionElement),
	ToolPlusPlugin.configure({}),
	ToolPlusInputPlugin.withComponent(ToolPlusInputElement),
];
