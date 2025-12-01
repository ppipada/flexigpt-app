import { type Dispatch, forwardRef, type SetStateAction, useMemo, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp, FiEdit2, FiPlus, FiTrash, FiX } from 'react-icons/fi';

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

export const SystemPromptDropdown = forwardRef<HTMLDetailsElement, SystemPromptDropdownProps>(
	function SystemPromptDropdown(
		{ prompts, selectedPromptId, onSelect, onAdd, onEdit, onRemove, onClear, isOpen, setIsOpen },
		detailsRef
	) {
		const [isAddOpen, setIsAddOpen] = useState(false);
		const [isEditOpen, setIsEditOpen] = useState(false);
		const [editingItemId, setEditingItemId] = useState<string | null>(null);
		const [isDeleteOpen, setIsDeleteOpen] = useState(false);
		const [itemPendingDelete, setItemPendingDelete] = useState<SystemPromptItem | null>(null);

		const editingItem = useMemo(() => prompts.find(p => p.id === editingItemId) || null, [prompts, editingItemId]);

		function closeDropdown() {
			if (detailsRef && typeof detailsRef !== 'function' && detailsRef.current) {
				detailsRef.current.open = false;
			}
			setIsOpen(false);
		}

		const handleSelectItem = (item: SystemPromptItem) => {
			if (selectedPromptId === item.id) {
				// Toggle off if clicking the selected one
				onClear();
			} else {
				onSelect(item);
			}
			closeDropdown();
		};

		const handleEditItem = (item: SystemPromptItem) => {
			setEditingItemId(item.id);
			setIsEditOpen(true);
			closeDropdown();
		};

		const handleRemoveItem = (item: SystemPromptItem) => {
			if (item.locked) return;
			setItemPendingDelete(item);
			setIsDeleteOpen(true);
			closeDropdown();
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

		return (
			<div className="flex w-full justify-center">
				<details
					ref={detailsRef}
					className="dropdown dropdown-top dropdown-end w-full justify-center"
					open={isOpen}
					onToggle={e => {
						setIsOpen((e.currentTarget as HTMLDetailsElement).open);
					}}
				>
					<summary
						className="btn btn-xs text-neutral-custom w-full flex-1 items-center overflow-hidden border-none text-center text-nowrap shadow-none"
						title={selectedPromptId ? 'System Prompt (enabled)' : 'System Prompt (disabled)'}
					>
						<span className="min-w-0 truncate text-center text-xs font-normal">System Prompt?</span>
						{selectedPromptId ? (
							<FiCheck size={16} className="m-0 shrink-0 p-0" />
						) : (
							<FiX size={16} className="m-0 shrink-0 p-0" />
						)}{' '}
						<span className="text-xs font-normal"></span>
						{isOpen ? (
							<FiChevronDown size={16} className="ml-2 shrink-0" />
						) : (
							<FiChevronUp size={16} className="ml-2 shrink-0" />
						)}
					</summary>

					<ul className="dropdown-content menu bg-base-100 w-full rounded-xl p-1">
						{prompts.length > 0 ? (
							prompts.map(item => (
								<li key={item.id} className="w-full p-1">
									{/* ---------- title (now really truncates) --------------------- */}
									<div className="pointer-events-none w-full px-0 py-1 hover:bg-inherit hover:shadow-none">
										<span className="truncate text-start text-xs font-normal">{item.title}</span>
									</div>
									{/* ---------- action row --------------------------------------- */}
									<div className="flex items-center gap-1 p-0 hover:bg-inherit hover:shadow-none">
										{/* chip / select button (left) */}
										{selectedPromptId === item.id ? (
											<span className="bg-success/10 text-success inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-medium">
												<FiCheck className="h-3 w-3" />
												Current
											</span>
										) : (
											<button
												type="button"
												className="btn btn-ghost btn-xs px-2"
												onClick={() => {
													handleSelectItem(item);
												}}
												title="Use this prompt"
											>
												Use
											</button>
										)}

										{/* edit + delete (right) */}
										<div className="ml-auto flex gap-1">
											<button
												type="button"
												className="btn btn-ghost btn-xs"
												title="Edit"
												onClick={e => {
													e.stopPropagation();
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
													handleRemoveItem(item);
												}}
											>
												<FiTrash size={12} />
											</button>
										</div>
									</div>
								</li>
							))
						) : (
							<li className="text-xs opacity-70">
								<div className="m-0 flex cursor-default items-center justify-between p-1">
									<span>No saved prompts</span>
								</div>
							</li>
						)}

						<li className="text-xs">
							<hr className="border-neutral/20 my-1 border-0 border-t p-0" />
							<div className="flex items-center justify-between p-1 hover:bg-inherit hover:shadow-none">
								<button
									type="button"
									className="btn btn-ghost btn-xs rounded-lg"
									onClick={() => {
										closeDropdown();
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
										closeDropdown();
									}}
									title="Clear current system prompt"
									disabled={!selectedPromptId}
								>
									<FiX size={14} className="mr-1" /> Clear
								</button>
							</div>
						</li>
					</ul>
				</details>

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
		);
	}
);
