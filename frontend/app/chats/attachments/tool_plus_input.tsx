import React from 'react';

import { FiTool } from 'react-icons/fi';

import type { PlateEditor } from 'platejs/react';

import type { ToolListItem } from '@/spec/tool';

import { useTools } from '@/hooks/use_tool';

// Expecting toolStoreAPI to be exported from your base API layer.
// If your path differs, adjust the import accordingly.
import { toolStoreAPI } from '@/apis/baseapi';

import { SlashInputElement } from '@/components/editor/nodes/slash_node';

import { insertToolSelectionNode } from '@/chats/attachments/tool_editor_utils';

export function ToolPlusInputElement(props: Omit<Parameters<typeof SlashInputElement>[0], 'trigger' | 'groups'>) {
	const { data, loading } = useTools();

	const groups = React.useMemo(() => {
		if (loading) return [];
		return [
			{
				group: 'Tools',
				items: data.map(it => {
					const pretty = it.toolSlug.replace(/[-_]/g, ' ');
					return {
						slug: `${it.bundleSlug}/${it.toolSlug}`,
						displayName: pretty,
						icon: <FiTool />,
						data: it,
						focusEditor: false,
						onSelect: async (
							editor: PlateEditor,
							item: ToolListItem & { bundleID: string; bundleSlug: string; toolSlug: string; toolVersion: string }
						) => {
							try {
								const tool = await toolStoreAPI.getTool(item.bundleID, item.toolSlug, item.toolVersion);
								insertToolSelectionNode(
									editor,
									{
										bundleID: item.bundleID,
										bundleSlug: item.bundleSlug,
										toolSlug: item.toolSlug,
										toolVersion: item.toolVersion,
									},
									tool
								);
							} catch {
								insertToolSelectionNode(editor, {
									bundleID: item.bundleID,
									bundleSlug: item.bundleSlug,
									toolSlug: item.toolSlug,
									toolVersion: item.toolVersion,
								});
							}
						},
					};
				}),
			},
		];
	}, [data, loading]);

	return <SlashInputElement {...props} trigger="+" groups={groups} />;
}
