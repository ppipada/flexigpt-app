import { FiChevronUp, FiTool, FiX } from 'react-icons/fi';

import { Menu, MenuButton, MenuItem, useMenuStore } from '@ariakit/react';
import type { Path } from 'platejs';
import type { PlateEditor } from 'platejs/react';

import { removeToolByKey, toolIdentityKey, type ToolSelectionElementNode } from '@/chats/tools/tool_editor_utils';

interface ToolChoicesChipProps {
	editor: PlateEditor;
	// Entries from getToolNodesWithPath(editor); typed loosely here.
	toolEntries: Array<[ToolSelectionElementNode, Path]>;
}

/**
 * Aggregated "Tools" chip for attached tool choices.
 * - Shows a count of selected tools.
 * - Opens a dropdown listing each tool with an individual remove button.
 * - Has a "remove all" cross that clears all attached tools.
 */
export function ToolChoicesChip({ editor, toolEntries }: ToolChoicesChipProps) {
	const count = toolEntries.length;
	const menu = useMenuStore({ placement: 'bottom-start', focusLoop: true });

	if (count === 0) return null;

	const title = `Tools\n${count} tool${count === 1 ? '' : 's'} attached`;

	const handleRemoveSingle = (node: ToolSelectionElementNode) => {
		const key = toolIdentityKey(node.bundleID, node.bundleSlug, node.toolSlug, node.toolVersion);
		if (!key) return;
		removeToolByKey(editor, key);
	};

	const handleRemoveAll = () => {
		const seen = new Set<string>();

		for (const [node] of toolEntries) {
			const key = toolIdentityKey(node.bundleID, node.bundleSlug, node.toolSlug, node.toolVersion);
			if (!key || seen.has(key)) continue;
			seen.add(key);
			removeToolByKey(editor, key);
		}

		menu.hide();
	};

	return (
		<div
			className="bg-base-200 text-base-content flex shrink-0 items-center gap-1 rounded-2xl px-2 py-0"
			title={title}
			data-attachment-chip="tools-group"
		>
			<FiTool size={14} />
			<span className="max-w-24 truncate">Tools</span>
			<span className="text-base-content/60 whitespace-nowrap">{count}</span>

			<MenuButton
				store={menu}
				className="btn btn-ghost btn-xs px-0 py-0 shadow-none"
				aria-label="Show selected tools"
				title="Show selected tools"
			>
				<FiChevronUp size={14} />
			</MenuButton>

			{/* Remove all tool choices */}
			<button
				type="button"
				className="btn btn-ghost btn-xs text-error shrink-0 px-0 py-0 shadow-none"
				onClick={handleRemoveAll}
				title="Remove all tools"
				aria-label="Remove all tools"
			>
				<FiX size={14} />
			</button>

			<Menu
				store={menu}
				gutter={6}
				className="rounded-box bg-base-100 text-base-content border-base-300 z-50 max-h-72 min-w-65 overflow-y-auto border p-2 shadow-xl focus-visible:outline-none"
				autoFocusOnShow
			>
				<div className="text-base-content/70 mb-1 text-[11px] font-semibold">Tools</div>

				{toolEntries.map(([node]) => {
					const rawDisplay: string | undefined = node.toolSnapshot?.displayName ?? node.toolSlug;
					const display = rawDisplay && rawDisplay.length > 0 ? rawDisplay : 'Tool';
					const slug = `${node.bundleSlug ?? node.bundleID}/${node.toolSlug}@${node.toolVersion}`;
					const truncatedDisplay = display.length > 40 ? `${display.slice(0, 37)}…` : display;

					return (
						<MenuItem
							key={node.selectionID}
							store={menu}
							hideOnClick={false}
							className="data-active-item:bg-base-200 mb-1 rounded-xl last:mb-0"
						>
							<div
								className="flex items-center gap-2 px-2 py-1"
								title={`Tool choice: ${display} (${slug})`}
								data-attachment-chip="tool-choice"
								data-selection-id={node.selectionID}
							>
								<FiTool size={14} />
								<div className="min-w-0 flex-1">
									<div className="truncate text-xs font-medium">{truncatedDisplay}</div>
									<div className="text-base-content/70 truncate text-[11px]">{slug}</div>
								</div>
								<button
									type="button"
									className="btn btn-ghost btn-xs text-error shrink-0 px-1 py-0 shadow-none"
									onClick={() => {
										handleRemoveSingle(node);
									}}
									title="Remove tool choice"
									aria-label="Remove tool choice"
								>
									<FiX size={12} />
								</button>
							</div>
						</MenuItem>
					);
				})}
			</Menu>
		</div>
	);
}
