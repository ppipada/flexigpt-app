import { type TriggerComboboxPluginOptions, withTriggerCombobox } from '@platejs/combobox';
import { createSlatePlugin, createTSlatePlugin, KEYS, type PluginConfig } from 'platejs';

import { TemplateSlashInputElement } from '@/chats/templates/template_slash_input';
import { TemplateSelectionElement } from '@/chats/templates/template_slash_selection';
import {
	KEY_TEMPLATE_SELECTION,
	KEY_TEMPLATE_SLASH_COMMAND,
	KEY_TEMPLATE_SLASH_INPUT,
	KEY_TEMPLATE_VARIABLE,
} from '@/chats/templates/template_spec';
import { TemplateVariableElement } from '@/chats/templates/template_variables_inline';

type TemplateSlashConfig = PluginConfig<typeof KEY_TEMPLATE_SLASH_COMMAND, TriggerComboboxPluginOptions>;

const TemplateSlashInputPlugin = createSlatePlugin({
	key: KEY_TEMPLATE_SLASH_INPUT,
	editOnly: true,
	node: { isElement: true, isInline: true, isVoid: true },
});

const TemplateSlashPlugin = createTSlatePlugin<TemplateSlashConfig>({
	key: KEY_TEMPLATE_SLASH_COMMAND,
	editOnly: true,
	options: {
		trigger: '/',
		triggerPreviousCharPattern: /^\s?$/,
		createComboboxInput: () => ({
			children: [{ text: '' }],
			type: KEY_TEMPLATE_SLASH_INPUT,
		}),
		triggerQuery: editor =>
			!editor.api.some({
				match: { type: editor.getType(KEYS.codeBlock) },
			}),
	},
	plugins: [TemplateSlashInputPlugin],
}).overrideEditor(withTriggerCombobox);

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
	TemplateSelectionPlugin.withComponent(TemplateSelectionElement),
	TemplateSlashPlugin.configure({}),
	TemplateSlashInputPlugin.withComponent(TemplateSlashInputElement),
	TemplateVariablePlugin.withComponent(TemplateVariableElement),
];
