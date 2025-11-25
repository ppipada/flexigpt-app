import { createSlatePlugin } from 'platejs';

import { TemplateSelectionElement } from '@/chats/templates/template_slash_selection';
import { KEY_TEMPLATE_SELECTION, KEY_TEMPLATE_VARIABLE } from '@/chats/templates/template_spec';
import { TemplateVariableElement } from '@/chats/templates/template_variables_inline';

// type TemplateSlashConfig = PluginConfig<typeof KEY_TEMPLATE_SLASH_COMMAND, TriggerComboboxPluginOptions>;

// const TemplateSlashInputPlugin = createSlatePlugin({
// 	key: KEY_TEMPLATE_SLASH_INPUT,
// 	editOnly: true,
// 	node: { isElement: true, isInline: true, isVoid: true },
// });

// const TemplateSlashPlugin = createTSlatePlugin<TemplateSlashConfig>({
// 	key: KEY_TEMPLATE_SLASH_COMMAND,
// 	editOnly: true,
// 	options: {
// 		trigger: '/',
// 		triggerPreviousCharPattern: /^\s?$/,
// 		createComboboxInput: () => ({
// 			children: [{ text: '' }],
// 			type: KEY_TEMPLATE_SLASH_INPUT,
// 		}),
// 		triggerQuery: editor =>
// 			!editor.api.some({
// 				match: { type: editor.getType(KEYS.codeBlock) },
// 			}),
// 	},
// 	plugins: [TemplateSlashInputPlugin],
// }).overrideEditor(withTriggerCombobox);

const TemplateSelectionPlugin = createSlatePlugin({
	key: KEY_TEMPLATE_SELECTION,
	// this is the chip “schema”
	node: { isElement: true, isInline: true, isVoid: true },
});

// Plugin
const TemplateVariablePlugin = createSlatePlugin({
	key: KEY_TEMPLATE_VARIABLE,
	node: { isElement: true, isInline: true, isVoid: true },
});

export const TemplateSlashKit = [
	// Data-carrier for selections:
	TemplateSelectionPlugin.withComponent(TemplateSelectionElement),
	// Data-carrier for variables:
	TemplateVariablePlugin.withComponent(TemplateVariableElement),

	// (Optional) If we ever want the inline "/" menu back, re-add:
	// TemplateSlashPlugin.configure({}),
	// TemplateSlashInputPlugin.withComponent(TemplateSlashInputElement),
];
