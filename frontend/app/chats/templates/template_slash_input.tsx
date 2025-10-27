import { useMemo } from 'react';

import { FiZap } from 'react-icons/fi';

import type { PlateEditor } from 'platejs/react';

import { usePromptTemplates } from '@/hooks/use_template';

import { promptStoreAPI } from '@/apis/baseapi';

import { SlashInputElement } from '@/components/editor/nodes/slash_node';

import { insertTemplateSelectionNode } from '@/chats/templates/template_editor_utils';

export function TemplateSlashInputElement(props: Omit<Parameters<typeof SlashInputElement>[0], 'trigger' | 'groups'>) {
	const { data, loading } = usePromptTemplates();

	const groups = useMemo(() => {
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
