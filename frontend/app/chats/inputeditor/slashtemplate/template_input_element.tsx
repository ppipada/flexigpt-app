import React from 'react';

import type { PlateEditor } from 'platejs/react';
import { FiZap } from 'react-icons/fi';

import { usePromptTemplates } from '@/hooks/use_template';

import { SlashInputElement } from '@/components/editor/nodes/slash_node';

import { KEY_TEMPLATE_SELECTION } from '@/chats/inputeditor/slashtemplate/template_selection_element';

function insertTemplateSelectionNode(
	editor: PlateEditor,
	bundleID: string,
	templateSlug: string,
	templateVersion: string
) {
	const node = {
		type: KEY_TEMPLATE_SELECTION,
		bundleID,
		templateSlug,
		templateVersion,
		variables: {},
		// void elements still need one empty text child in Slate
		children: [{ text: '' }],
	};

	editor.tf.withoutNormalizing(() => {
		// Insert the chip (inline+void)
		editor.tf.insertNodes(node, { select: true });

		// Move caret after the chip and add a trailing space so the user can keep typing
		editor.tf.collapse({ edge: 'end' });
		editor.tf.insertText(' ');
	});
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
					onSelect: (
						editor: PlateEditor,
						item: { bundleID: string; templateSlug: string; templateVersion: string }
					) => {
						insertTemplateSelectionNode(editor, item.bundleID, item.templateSlug, item.templateVersion);
					},
				})),
			},
		];
	}, [data, loading]);

	return <SlashInputElement {...props} trigger="/" groups={groups} />;
}
