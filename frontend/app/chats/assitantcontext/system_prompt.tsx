import React, { forwardRef, useEffect, useMemo, useState } from 'react';

import { createPortal } from 'react-dom';
import { FiCheck, FiChevronDown, FiChevronUp, FiCopy, FiEdit2, FiPlus, FiTrash, FiX } from 'react-icons/fi';

import { getUUIDv7 } from '@/lib/uuid_utils';

import Dropdown from '@/components/dropdown';

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
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
};

/* -------------------------- Shared Prompt Modal -------------------------- */

const PromptModal: React.FC<{
	isOpen: boolean;
	mode: 'add' | 'edit';
	initialValue?: string;
	promptsForCopy?: SystemPromptItem[];
	onClose: () => void;
	onSave: (value: string) => void;
}> = ({ isOpen, mode, initialValue = '', promptsForCopy = [], onClose, onSave }) => {
	const [value, setValue] = useState<string>(initialValue);
	const [copyFromId, setCopyFromId] = useState<string>('');

	useEffect(() => {
		if (isOpen) {
			setValue(initialValue);
			setCopyFromId('');
		}
	}, [isOpen, initialValue]);

	const save = (e: React.FormEvent) => {
		e.preventDefault();
		const v = value.trim();
		if (!v) return;
		onSave(v);
		onClose();
	};

	const handleCopyFrom = (id: string) => {
		setCopyFromId(id);
		const found = promptsForCopy.find(p => p.id === id);
		if (found) {
			setValue(found.prompt);
		}
	};

	// Build dropdown items for "Copy Existing"
	const copyDropdownItems = useMemo(() => {
		const map: Record<string, { isEnabled: boolean }> = {};
		for (const p of promptsForCopy) {
			map[p.id] = { isEnabled: true };
		}
		return map;
	}, [promptsForCopy]);

	const getCopyDisplayName = (id: string) => {
		const found = promptsForCopy.find(p => p.id === id);
		return found?.title ?? id;
	};

	if (!isOpen) return null;
	return createPortal(
		<dialog className="modal modal-open">
			<div className="modal-box max-h-[80vh] max-w-3xl overflow-auto rounded-2xl">
				<div className="mb-4 flex items-center justify-between">
					<h3 className="text-lg font-bold">{mode === 'add' ? 'Add System Prompt' : 'Edit System Prompt'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close">
						<FiX size={12} />
					</button>
				</div>

				<form onSubmit={save} className="space-y-4">
					{mode === 'add' && (
						<div className="grid grid-cols-12 items-center gap-1">
							<label className="col-span-2 text-sm opacity-70">Copy Existing:</label>
							<div className="col-span-9">
								<Dropdown<string>
									dropdownItems={copyDropdownItems}
									selectedKey={copyFromId}
									onChange={key => {
										handleCopyFrom(key);
									}}
									filterDisabled={false}
									title="Select a saved prompt to copy"
									getDisplayName={getCopyDisplayName}
									maxMenuHeight={260}
								/>
							</div>
							<button
								type="button"
								className="btn btn-ghost btn-xs col-span-1 p-4"
								title="Copy again"
								onClick={() => {
									if (copyFromId) handleCopyFrom(copyFromId);
								}}
								disabled={!copyFromId}
							>
								<FiCopy size={14} />
							</button>
						</div>
					)}

					<div>
						<textarea
							className="textarea textarea-bordered h-40 w-full rounded-xl"
							value={value}
							onChange={e => {
								setValue(e.target.value);
							}}
							placeholder="Enter system prompt instructions here..."
							spellCheck="false"
						/>
					</div>

					<div className="modal-action">
						<button type="button" className="btn rounded-xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-primary rounded-xl" disabled={!value.trim()}>
							Save
						</button>
					</div>
				</form>
			</div>
		</dialog>,
		document.body
	);
};

/* ------------------------ Main Dropdown Component ------------------------ */

const SystemPromptDropdown = forwardRef<HTMLDetailsElement, SystemPromptDropdownProps>(
	({ prompts, selectedPromptId, onSelect, onAdd, onEdit, onRemove, onClear, isOpen, setIsOpen }, detailsRef) => {
		const [isAddOpen, setIsAddOpen] = useState(false);
		const [isEditOpen, setIsEditOpen] = useState(false);
		const [editingItemId, setEditingItemId] = useState<string | null>(null);

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
			const confirm = window.confirm('Remove this saved prompt?');
			if (!confirm) return;

			// If removing the currently selected item, clear selection
			if (selectedPromptId === item.id) {
				onClear();
			}
			onRemove(item.id);
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
						<span className="min-w-0 truncate text-center text-xs font-normal">System Prompt:</span>
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
				<PromptModal
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
				<PromptModal
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
			</div>
		);
	}
);

SystemPromptDropdown.displayName = 'SystemPromptDropdown';
export default SystemPromptDropdown;
