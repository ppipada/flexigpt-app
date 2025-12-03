import type { Dispatch, SetStateAction } from 'react';
import { useMemo, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp, FiEdit2, FiPlus, FiTrash, FiX } from 'react-icons/fi';

import {
	Select,
	SelectItem,
	SelectPopover,
	Tooltip,
	useSelectStore,
	useStoreState,
	useTooltipStore,
} from '@ariakit/react';

import { getUUIDv7 } from '@/lib/uuid_utils';

import { DeleteConfirmationModal } from '@/components/delete_confirmation_modal';

import { SystemPromptAddEditModal } from '@/chats/assitantcontexts/system_prompt_add_edit_modal';

export type SystemPromptItem = {
	id: string;
	title: string;
	prompt: string;
	locked?: boolean; // when true, cannot be deleted
};

function buildTitleFromPrompt(p: string): string {
	const s = (p || '').trim();
	if (!s) return '(empty)';
	return s.length > 64 ? s.slice(0, 64) : s;
}

export function createSystemPromptItem(prompt: string, opts?: { locked?: boolean }): SystemPromptItem {
	return {
		id: getUUIDv7(),
		title: buildTitleFromPrompt(prompt),
		prompt,
		locked: Boolean(opts?.locked),
	};
}

type SystemPromptDropdownProps = {
	prompts: SystemPromptItem[];
	selectedPromptId?: string;
	onSelect: (item: SystemPromptItem) => void;
	onAdd: (item: SystemPromptItem) => void;
	onEdit: (id: string, updatedPrompt: string) => void;
	onRemove: (id: string) => void;
	onClear: () => void;
	isOpen: boolean;
	setIsOpen: Dispatch<SetStateAction<boolean>>;
};

export function SystemPromptDropdown({
	prompts,
	selectedPromptId,
	onSelect,
	onAdd,
	onEdit,
	onRemove,
	onClear,
	isOpen,
	setIsOpen,
}: SystemPromptDropdownProps) {
	const [isAddOpen, setIsAddOpen] = useState(false);
	const [isEditOpen, setIsEditOpen] = useState(false);
	const [editingItemId, setEditingItemId] = useState<string | null>(null);
	const [isDeleteOpen, setIsDeleteOpen] = useState(false);
	const [itemPendingDelete, setItemPendingDelete] = useState<SystemPromptItem | null>(null);

	const editingItem = useMemo(() => prompts.find(p => p.id === editingItemId) || null, [prompts, editingItemId]);

	const select = useSelectStore({
		value: selectedPromptId ?? '',
		setValue: id => {
			if (typeof id !== 'string') return;

			if (!id) {
				onClear();
				return;
			}
			const item = prompts.find(p => p.id === id);
			if (!item) return;

			// Clicking the same item again toggles off
			if (selectedPromptId === id) {
				onClear();
			} else {
				onSelect(item);
			}
		},
		open: isOpen,
		setOpen: setIsOpen,
		placement: 'top-start',
		focusLoop: true,
	});

	const open = useStoreState(select, 'open');

	// Tooltip for full prompt text on each item.
	// - placement: 'top-start' -> above the item, left-aligned.
	// - portal: true on <Tooltip> so it escapes the popover box.
	const promptTooltip = useTooltipStore({ placement: 'left-end' });
	const tooltipAnchorEl = useStoreState(promptTooltip, 'anchorElement');
	const currentPromptText = tooltipAnchorEl?.dataset.prompt ?? '';

	const handleEditItem = (item: SystemPromptItem) => {
		setEditingItemId(item.id);
		setIsEditOpen(true);
		setIsOpen(false);
	};

	const handleRemoveItem = (item: SystemPromptItem) => {
		if (item.locked) return;
		setItemPendingDelete(item);
		setIsDeleteOpen(true);
		setIsOpen(false);
	};

	const handleConfirmDelete = () => {
		if (!itemPendingDelete) return;

		if (selectedPromptId === itemPendingDelete.id) {
			onClear();
		}
		onRemove(itemPendingDelete.id);

		setItemPendingDelete(null);
		setIsDeleteOpen(false);
	};

	const handleCancelDelete = () => {
		setItemPendingDelete(null);
		setIsDeleteOpen(false);
	};

	const hasPrompts = prompts.length > 0;

	return (
		<div className="flex w-full justify-center">
			<div className="relative w-full">
				<Select
					store={select}
					className="btn btn-xs text-neutral-custom w-full flex-1 items-center overflow-hidden border-none text-center text-nowrap shadow-none"
					title={selectedPromptId ? 'System Prompt (enabled)' : 'System Prompt (disabled)'}
				>
					<span className="min-w-0 truncate text-center text-xs font-normal">System Prompt?</span>
					{selectedPromptId ? (
						<FiCheck size={16} className="m-0 shrink-0 p-0" />
					) : (
						<FiX size={16} className="m-0 shrink-0 p-0" />
					)}
					{open ? (
						<FiChevronDown size={16} className="ml-2 shrink-0" />
					) : (
						<FiChevronUp size={16} className="ml-2 shrink-0" />
					)}
				</Select>

				<SelectPopover
					store={select}
					portal={false}
					gutter={4}
					autoFocusOnShow
					sameWidth
					className="border-base-300 bg-base-100 z-50 mt-1 max-h-72 w-full overflow-y-auto rounded-xl border p-1 text-xs shadow-lg outline-none"
				>
					{hasPrompts ? (
						prompts.map(item => (
							<SelectItem
								key={item.id}
								value={item.id}
								className="hover:bg-base-200 data-active-item:bg-base-300 m-0 flex cursor-pointer flex-col gap-1 rounded-md px-2 py-1 text-xs transition-colors outline-none"
								// Used by the tooltip store to get the full prompt text
								data-prompt={item.prompt}
								onFocus={e => {
									// Keyboard focus -> show tooltip above this item
									promptTooltip.setAnchorElement(e.currentTarget as HTMLElement);
									promptTooltip.show();
								}}
								onBlur={() => {
									// Leaving item -> hide tooltip
									promptTooltip.hide();
									promptTooltip.setAnchorElement(null);
								}}
								onMouseEnter={e => {
									// Mouse hover -> show tooltip above this item
									promptTooltip.setAnchorElement(e.currentTarget as HTMLElement);
									promptTooltip.show();
								}}
								onMouseLeave={() => {
									// Mouse leave -> hide tooltip
									promptTooltip.hide();
								}}
							>
								{/* Title (truncated in list) */}
								<div className="truncate text-start text-xs font-normal">{item.title}</div>

								{/* Action row */}
								<div className="flex items-center gap-1">
									{selectedPromptId === item.id && (
										<span className="bg-success/10 text-success inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium">
											<FiCheck className="h-3 w-3" />
											Current
										</span>
									)}

									<div className="ml-auto flex gap-1">
										<button
											type="button"
											className="btn btn-ghost btn-xs"
											title="Edit"
											onClick={e => {
												e.stopPropagation();
												e.preventDefault();
												handleEditItem(item);
											}}
										>
											<FiEdit2 size={12} />
										</button>
										<button
											type="button"
											className="btn btn-ghost btn-xs"
											title={item.locked ? 'Cannot delete default prompt' : 'Delete'}
											disabled={item.locked}
											onClick={e => {
												e.stopPropagation();
												e.preventDefault();
												handleRemoveItem(item);
											}}
										>
											<FiTrash size={12} />
										</button>
									</div>
								</div>
							</SelectItem>
						))
					) : (
						<div className="m-0 flex cursor-default items-center justify-between rounded-md px-2 py-1 text-xs opacity-70">
							<span>No saved prompts</span>
						</div>
					)}

					{/* Add / Clear row */}
					<div className="border-neutral/20 mt-1 border-t pt-1 text-xs">
						<div className="flex items-center justify-between p-1">
							<button
								type="button"
								className="btn btn-ghost btn-xs rounded-lg"
								onClick={() => {
									setIsOpen(false);
									setIsAddOpen(true);
								}}
							>
								<FiPlus size={14} className="mr-1" /> Add
							</button>

							<button
								type="button"
								className="btn btn-ghost btn-xs rounded-lg"
								onClick={() => {
									onClear();
									setIsOpen(false);
								}}
								title="Clear current system prompt"
								disabled={!selectedPromptId}
							>
								<FiX size={14} className="mr-1" /> Clear
							</button>
						</div>
					</div>
				</SelectPopover>

				{/* Tooltip showing full prompt text, above the currently hovered/focused item */}
				<Tooltip
					store={promptTooltip}
					portal
					className="rounded-box bg-base-100 text-base-content border-base-300 max-w-xl border p-2 text-xs whitespace-pre-wrap shadow-xl"
				>
					{currentPromptText}
				</Tooltip>

				{/* Add Modal */}
				<SystemPromptAddEditModal
					isOpen={isAddOpen}
					mode="add"
					initialValue=""
					promptsForCopy={prompts}
					onClose={() => {
						setIsAddOpen(false);
					}}
					onSave={value => {
						const newItem = createSystemPromptItem(value);
						onAdd(newItem);
					}}
				/>

				{/* Edit Modal */}
				<SystemPromptAddEditModal
					isOpen={isEditOpen}
					mode="edit"
					initialValue={editingItem?.prompt ?? ''}
					promptsForCopy={prompts}
					onClose={() => {
						setIsEditOpen(false);
						setEditingItemId(null);
					}}
					onSave={value => {
						if (!editingItem) return;
						onEdit(editingItem.id, value);
					}}
				/>

				{/* Delete confirmation */}
				<DeleteConfirmationModal
					isOpen={isDeleteOpen}
					onClose={handleCancelDelete}
					onConfirm={handleConfirmDelete}
					title="Remove saved prompt?"
					message={
						itemPendingDelete
							? `This will remove "${itemPendingDelete.title}" from your saved system prompts. This action cannot be undone.`
							: 'This will remove the saved system prompt. This action cannot be undone.'
					}
					confirmButtonText="Remove"
				/>
			</div>
		</div>
	);
}
