import { createSlatePlugin } from 'platejs';

import { TemplateSelectionElement } from '@/chats/templates/template_slash_selection';
import { KEY_TEMPLATE_SELECTION, KEY_TEMPLATE_VARIABLE } from '@/chats/templates/template_spec';
import { TemplateVariableElement } from '@/chats/templates/template_variables_inline';

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
];
