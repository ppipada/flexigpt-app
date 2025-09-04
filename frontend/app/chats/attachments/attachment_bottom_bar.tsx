import { FiTool, FiX } from 'react-icons/fi';

import { type PlateEditor, useEditorRef } from 'platejs/react';

import { getToolNodesWithPath, removeToolAtPath } from '@/chats/attachments/tool_editor_utils';

/**
  Bottom bar for rendering attached items (Tools/Files/Docs).
  - Only renders if there is at least one attachment.
  - Single row, horizontally scrollable chips.
  - Uses DaisyUI/Tailwind to match the look-and-feel of your template toolbar.
*/
export function AttachmentBottomBar() {
	const editor = useEditorRef() as PlateEditor;

	// Recompute on each render; the parent EditorArea re-renders on doc changes.
	const toolEntries = getToolNodesWithPath(editor);

	// Hide bar completely when there are no attachments
	if (toolEntries.length === 0) return null;

	return (
		<div
			className="border-base-300 bg-base-100/95 supports-[backdrop-filter]:bg-base-100/60 w-full border-t backdrop-blur"
			data-attachments-bottom-bar
			aria-label="Attachment"
		>
			<div className="flex items-center gap-2 px-2 py-1 font-mono text-xs">
				{/* Chips scroller */}
				<div className="no-scrollbar flex min-w-0 flex-1 items-center gap-2 overflow-x-auto py-0">
					{toolEntries.map(([node, path]) => {
						const n = node;
						const display = n.toolSnapshot?.displayName ?? n.toolSlug;
						const slug = `${n.bundleSlug ?? n.bundleID}/${n.toolSlug}@${n.toolVersion}`;

						return (
							<div
								key={n.selectionID}
								className="bg-base-200 hover:bg-base-300/80 text-base-content flex shrink-0 items-center gap-2 rounded-2xl px-2 py-0 text-xs"
								title={`Tool - ${display} (${slug})`}
								data-attachment-chip="tool"
								data-selection-id={n.selectionID}
							>
								<FiTool />
								<span className="truncate">{display.length > 36 ? display.slice(0, 36) + '...' : display}</span>
								<button
									type="button"
									className="btn btn-ghost btn-xs text-error shrink-0 px-1 py-0 shadow-none"
									onClick={() => {
										removeToolAtPath(editor, path);
									}}
									title="Remove tool"
									aria-label="Remove tool"
								>
									<FiX />
								</button>
							</div>
						);
					})}
				</div>
			</div>
		</div>
	);
}
