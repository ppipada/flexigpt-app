import { useMemo } from 'react';

import { FiChevronUp, FiEdit2, FiGlobe, FiX } from 'react-icons/fi';

import { Menu, MenuButton, MenuItem, useMenuStore } from '@ariakit/react';

import type { ToolListItem, UIToolUserArgsStatus } from '@/spec/tool';

type SelectedIdentity = { bundleID: string; toolSlug: string; toolVersion: string };

export function WebSearchBottomBarChip({
	eligibleTools,
	enabled,
	selected,
	canEdit,
	argsStatus,
	onEnabledChange,
	onSelectTool,
	onEditOptions,
}: {
	eligibleTools: ToolListItem[];
	enabled: boolean;
	selected?: SelectedIdentity;
	canEdit: boolean;
	argsStatus?: UIToolUserArgsStatus;
	onEnabledChange: (enabled: boolean) => void;
	onSelectTool: (tool: ToolListItem) => void;
	onEditOptions: () => void;
}) {
	if (eligibleTools.length === 0) return null;

	const hasDropdown = eligibleTools.length > 1;
	const menu = useMenuStore({ placement: 'top-end', focusLoop: true });

	const selectedLabel = useMemo(() => {
		if (!selected) return '';
		const hit = eligibleTools.find(
			t =>
				t.bundleID === selected.bundleID && t.toolSlug === selected.toolSlug && t.toolVersion === selected.toolVersion
		);
		return hit?.toolDefinition.displayName ?? hit?.toolSlug ?? '';
	}, [eligibleTools, selected]);

	const isArgsBad = !!(enabled && argsStatus?.hasSchema && !argsStatus.isSatisfied);

	const container =
		(enabled
			? 'bg-info/10 text-neutral-custom border-info/50 hover:bg-info/15'
			: 'bg-base-200 text-neutral-custom border-base-300/40 hover:bg-base-300/80') +
		' flex items-center gap-1 rounded-2xl border px-2 py-0';

	const titleLines: string[] = [];
	titleLines.push('Web search');
	if (selectedLabel) titleLines.push(`Tool: ${selectedLabel}`);
	titleLines.push(enabled ? 'Status: Enabled' : 'Status: Disabled');
	if (isArgsBad) titleLines.push('Options: Missing required fields');

	return (
		<div className={container} title={titleLines.join('\n')} data-bottom-bar-websearch>
			{/* Main toggle */}
			<button
				type="button"
				className="flex items-center gap-2"
				onClick={() => {
					onEnabledChange(!enabled);
				}}
				aria-label={enabled ? 'Web search' : 'Enable web search'}
			>
				<FiGlobe size={14} />
				<span className="max-w-28 truncate">{enabled ? 'Web search' : 'Enable web search'}</span>
				{selectedLabel ? <span className="max-w-36 truncate text-xs opacity-70">{selectedLabel}</span> : null}
			</button>

			{/* Args warning badge (lightweight) */}
			{isArgsBad ? <span className="badge badge-warning badge-xs ml-1">Options</span> : null}

			{/* Edit options (only when enabled) */}
			{enabled && canEdit ? (
				<button
					type="button"
					className="btn btn-ghost btn-xs p-0 shadow-none"
					onClick={e => {
						e.preventDefault();
						e.stopPropagation();
						onEditOptions();
					}}
					title="Edit web-search options"
					aria-label="Edit web-search options"
				>
					<FiEdit2 size={12} />
				</button>
			) : null}

			{enabled ? (
				<button
					type="button"
					className="btn btn-ghost btn-xs p-0 shadow-none"
					onClick={e => {
						e.preventDefault();
						e.stopPropagation();
						onEnabledChange(!enabled);
					}}
					title="Disable Web Search"
					aria-label="Disable Web Search"
				>
					<FiX size={12} />
				</button>
			) : null}

			{/* Dropdown only when multiple eligible tools */}
			{hasDropdown ? (
				<>
					<MenuButton
						store={menu}
						className="btn btn-ghost btn-xs p-0 shadow-none"
						aria-label="Choose web-search tool"
						title="Choose web-search tool"
					>
						<FiChevronUp size={14} />
					</MenuButton>

					<Menu
						store={menu}
						gutter={6}
						className="rounded-box bg-base-100 text-base-content border-base-300 z-50 max-h-72 min-w-72 overflow-y-auto border p-2 shadow-xl"
						autoFocusOnShow
						portal
					>
						<div className="text-base-content/70 mb-2 text-[11px] font-semibold">Web search tools</div>
						{eligibleTools.map(t => {
							const isSelected =
								!!selected &&
								selected.bundleID === t.bundleID &&
								selected.toolSlug === t.toolSlug &&
								selected.toolVersion === t.toolVersion;
							return (
								<MenuItem
									key={`${t.bundleID}-${t.toolSlug}-${t.toolVersion}`}
									className="data-active-item:bg-base-200 mb-1 rounded-xl last:mb-0"
									onClick={() => {
										onSelectTool(t);
										menu.hide();
									}}
								>
									<div className="flex items-center gap-2 px-2 py-1">
										<FiGlobe size={14} />
										<div className="min-w-0 flex-1">
											<div className="truncate text-xs font-medium">
												{t.toolDefinition.displayName || t.toolSlug}
												{isSelected ? ' (selected)' : ''}
											</div>
											<div className="text-base-content/70 truncate text-[11px]">
												{
													// eslint-disable-next-line @typescript-eslint/no-unnecessary-condition
													t.bundleSlug ?? t.bundleID
												}
												/{t.toolSlug}@{t.toolVersion}
											</div>
										</div>
									</div>
								</MenuItem>
							);
						})}
					</Menu>
				</>
			) : null}
		</div>
	);
}
