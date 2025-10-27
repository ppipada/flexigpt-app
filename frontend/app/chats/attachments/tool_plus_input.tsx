import { useMemo } from 'react';

import { FiTool } from 'react-icons/fi';

import { type PlateEditor, useEditorRef } from 'platejs/react';

import type { ToolListItem } from '@/spec/tool';

import { useTools } from '@/hooks/use_tool';

// Expecting toolStoreAPI to be exported from your base API layer.
// If your path differs, adjust the import accordingly.
import { toolStoreAPI } from '@/apis/baseapi';

import { SlashInputElement } from '@/components/editor/nodes/slash_node';

import { getToolNodesWithPath, insertToolSelectionNode, toolIdentityKey } from '@/chats/attachments/tool_editor_utils';

export function ToolPlusInputElement(props: Omit<Parameters<typeof SlashInputElement>[0], 'trigger' | 'groups'>) {
	const { data, loading } = useTools();
	const editor = useEditorRef() as PlateEditor;
	const attachedKeys = useMemo(() => {
		const keys = new Set<string>();
		for (const [node] of getToolNodesWithPath(editor, false)) {
			const n = node;
			keys.add(toolIdentityKey(n.bundleID, n.bundleSlug, n.toolSlug, n.toolVersion));
		}
		return keys;
	}, [editor, editor.children]);
	const groups = useMemo(() => {
		if (loading) return [];
		// Filter out tools already attached in the current document.
		const filtered = data.filter(it => {
			// Some lists might not provide bundleID; toolIdentityKey handles fallback to slug.
			return !attachedKeys.has(toolIdentityKey(it.bundleID, it.bundleSlug, it.toolSlug, it.toolVersion));
		});
		return [
			{
				group: 'Tools',
				items: filtered.map(it => {
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
	}, [data, loading, attachedKeys]);

	return <SlashInputElement {...props} trigger="+" groups={groups} />;
}
