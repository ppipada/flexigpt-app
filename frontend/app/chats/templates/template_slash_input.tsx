import React from 'react';

import { FiZap } from 'react-icons/fi';

import { KEYS } from 'platejs';
import type { PlateEditor } from 'platejs/react';

import type { PromptTemplate } from '@/spec/prompt';

import { usePromptTemplates } from '@/hooks/use_template';

import { promptStoreAPI } from '@/apis/baseapi';

import { SlashInputElement } from '@/components/editor/nodes/slash_node';

import { KEY_TEMPLATE_SELECTION } from '@/chats/templates/template_editor_utils';
import { buildInitialToolStates } from '@/chats/templates/template_processing';

function insertTemplateSelectionNode(
	editor: PlateEditor,
	bundleID: string,
	templateSlug: string,
	templateVersion: string,
	template?: PromptTemplate
) {
	const selectionID = `tpl:${bundleID}/${templateSlug}@${templateVersion}:${Date.now().toString(36)}${Math.random()
		.toString(36)
		.slice(2, 8)}`;
	const nnode = {
		type: KEY_TEMPLATE_SELECTION,
		bundleID,
		templateSlug,
		templateVersion,
		selectionID,
		variables: {} as Record<string, unknown>,
		// Snapshot full template for downstream sync "get" to have the full context.
		templateSnapshot: template,
		// Local overrides
		overrides: {} as {
			displayName?: string;
			description?: string;
			tags?: string[];
			blocks?: PromptTemplate['blocks'];
			variables?: PromptTemplate['variables'];
			preProcessors?: PromptTemplate['preProcessors'];
		},
		// Each preprocessor call state
		toolStates: buildInitialToolStates(template),
		// void elements still need one empty text child in Slate
		children: [{ text: '' }],
	};

	editor.tf.withoutNormalizing(() => {
		// Insert the chip (inline+void)
		editor.tf.insertNodes([nnode, { type: KEYS.p, text: '\n' }], { select: true });
		// Move caret after the chip and add a trailing space so the user can keep typing
		editor.tf.collapse({ edge: 'end' });
		editor.tf.select(undefined, { edge: 'end' }); // Select end of block above
	});
	editor.tf.focus();
}

export function TemplateSlashInputElement(props: Omit<Parameters<typeof SlashInputElement>[0], 'trigger' | 'groups'>) {
	const { data, loading } = usePromptTemplates();

	const groups = React.useMemo(() => {
		if (loading) return [];
		return [
			{
				group: 'Templates',
				items: data.map(t => ({
					slug: `${t.bundleID}/${t.templateSlug}`,
					displayName: t.templateSlug.replace(/[-_]/g, ' '),
					icon: <FiZap />,
					data: t,
					focusEditor: false,
					onSelect: async (
						editor: PlateEditor,
						item: { bundleID: string; templateSlug: string; templateVersion: string }
					) => {
						try {
							const tmpl = await promptStoreAPI.getPromptTemplate(
								item.bundleID,
								item.templateSlug,
								item.templateVersion
							);
							insertTemplateSelectionNode(editor, item.bundleID, item.templateSlug, item.templateVersion, tmpl);
						} catch {
							// Fallback insert without snapshot if API fails.
							insertTemplateSelectionNode(editor, item.bundleID, item.templateSlug, item.templateVersion);
						}
					},
				})),
			},
		];
	}, [data, loading]);

	return <SlashInputElement {...props} trigger="/" groups={groups} />;
}
