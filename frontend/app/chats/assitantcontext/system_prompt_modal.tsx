import React, { forwardRef, useState } from 'react';

import { createPortal } from 'react-dom';
import { FiCheck, FiChevronDown, FiChevronUp, FiPlus, FiX } from 'react-icons/fi';

import { getUUIDv7 } from '@/lib/uuid_utils';

export type SystemPromptItem = {
	id: string;
	title: string;
	prompt: string;
};

export function buildTitleFromPrompt(p: string): string {
	const s = (p || '').trim();
	if (!s) return '(empty)';
	return s.length > 24 ? `${s.slice(0, 24)}…` : s;
}

export function createSystemPromptItem(prompt: string): SystemPromptItem {
	return {
		id: getUUIDv7(),
		title: buildTitleFromPrompt(prompt),
		prompt,
	};
}

type SystemPromptDropdownProps = {
	prompts: SystemPromptItem[];
	selectedPromptId?: string;
	onSelect: (item: SystemPromptItem) => void;
	onAdd: (item: SystemPromptItem) => void;
	isOpen: boolean;
	setIsOpen: React.Dispatch<React.SetStateAction<boolean>>;
};

const AddPromptModal: React.FC<{
	isOpen: boolean;
	onClose: () => void;
	onSave: (item: SystemPromptItem) => void;
}> = ({ isOpen, onClose, onSave }) => {
	const [value, setValue] = useState('');

	if (!isOpen) return null;

	const save = (e: React.FormEvent) => {
		e.preventDefault();
		const v = value.trim();
		if (!v) return;

		onSave(createSystemPromptItem(v));
		onClose();
	};

	return createPortal(
		<dialog className="modal modal-open">
			<div className="modal-box max-w-xl max-h-[80vh] overflow-auto rounded-2xl">
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">Add System Prompt</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close">
						<FiX size={12} />
					</button>
				</div>

				<form onSubmit={save} className="space-y-4">
					<div>
						<textarea
							className="textarea textarea-bordered w-full rounded-xl h-32"
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

const SystemPromptDropdown = forwardRef<HTMLDetailsElement, SystemPromptDropdownProps>(
	({ prompts, selectedPromptId, onSelect, onAdd, isOpen, setIsOpen }, detailsRef) => {
		const [isAddOpen, setIsAddOpen] = useState(false);

		function closeDropdown() {
			if (detailsRef && typeof detailsRef !== 'function' && detailsRef.current) {
				detailsRef.current.open = false;
			}
			setIsOpen(false);
		}

		return (
			<div className="flex w-full justify-center">
				<details
					ref={detailsRef}
					className="dropdown dropdown-top dropdown-end"
					open={isOpen}
					onToggle={e => {
						setIsOpen((e.currentTarget as HTMLDetailsElement).open);
					}}
				>
					<summary
						className="btn btn-xs text-nowrap text-neutral-custom shadow-none border-none overflow-hidden"
						title="System Prompt"
					>
						<div className="flex items-center">
							<span className="text-xs font-normal">System Prompt</span>
							{isOpen ? (
								<FiChevronDown size={16} className="ml-1 md:ml-2" />
							) : (
								<FiChevronUp size={16} className="ml-1 md:ml-2" />
							)}
						</div>
					</summary>

					<ul className="dropdown-content menu bg-base-100 rounded-xl w-full p-2">
						{prompts.length > 0 ? (
							prompts.map(item => (
								<li
									key={item.id}
									className="cursor-pointer text-xs"
									onClick={() => {
										onSelect(item);
										closeDropdown();
									}}
								>
									<a className="flex justify-between items-center p-1 m-0" title={item.prompt}>
										<span>{item.title}</span>
										{selectedPromptId === item.id && <FiCheck />}
									</a>
								</li>
							))
						) : (
							<li className="text-xs opacity-70">
								<a className="flex justify-between items-center p-1 m-0 cursor-default">
									<span>No saved prompts</span>
								</a>
							</li>
						)}

						<li className="text-xs">
							<hr className="p-0 my-2 border-0 border-t border-neutral/20" />
							<button
								type="button"
								className="btn btn-ghost btn-xs w-full rounded-lg justify-center"
								onClick={() => {
									closeDropdown();
									setIsAddOpen(true);
								}}
							>
								<FiPlus size={14} className="mr-1" /> Add
							</button>
						</li>
					</ul>
				</details>

				<AddPromptModal
					isOpen={isAddOpen}
					onClose={() => {
						setIsAddOpen(false);
					}}
					onSave={item => {
						onAdd(item);
					}}
				/>
			</div>
		);
	}
);

SystemPromptDropdown.displayName = 'SystemPromptDropdown';
export default SystemPromptDropdown;
