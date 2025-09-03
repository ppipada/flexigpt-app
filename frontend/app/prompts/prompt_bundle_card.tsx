import { type FC, useEffect, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp, FiEdit2, FiPlus, FiTrash2, FiX } from 'react-icons/fi';

import { PROMPT_TEMPLATE_INVOKE_CHAR } from '@/spec/command';
import type { PromptBundle, PromptTemplate } from '@/spec/prompt';

import { promptStoreAPI } from '@/apis/baseapi';
import { getAllPromptTemplates } from '@/apis/list_helper';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';

import AddEditPromptTemplate from '@/prompts/prompt_template_add_edit';

/* ---------- props ---------- */

interface PromptBundleCardProps {
	bundle: PromptBundle;
	templates: PromptTemplate[];

	onTemplatesChange: (bundleID: string, newTemplates: PromptTemplate[]) => void;
	onBundleDeleted: (bundle: PromptBundle) => void;
}

/* ---------- component ---------- */

const PromptBundleCard: FC<PromptBundleCardProps> = ({ bundle, templates, onTemplatesChange, onBundleDeleted }) => {
	/* local state */
	const [isExpanded, setIsExpanded] = useState(false);
	const [localTemplates, setLocalTemplates] = useState<PromptTemplate[]>(templates);

	const [isBundleEnabled, setIsBundleEnabled] = useState(bundle.isEnabled);

	useEffect(() => {
		setIsBundleEnabled(bundle.isEnabled);
	}, [bundle.isEnabled]);
	/* --------------------------------------------- */

	/* modals */
	const [isDeleteTemplateModalOpen, setIsDeleteTemplateModalOpen] = useState(false);
	const [templateToDelete, setTemplateToDelete] = useState<PromptTemplate | null>(null);

	const [isModifyModalOpen, setIsModifyModalOpen] = useState(false);
	const [templateToEdit, setTemplateToEdit] = useState<PromptTemplate | undefined>(undefined);

	const [showAlert, setShowAlert] = useState(false);
	const [alertMsg, setAlertMsg] = useState('');

	/* sync templates coming from parent */
	useEffect(() => {
		setLocalTemplates(templates);
	}, [templates]);

	/* ---------- bundle enable toggle ---------- */
	const toggleBundleEnable = async () => {
		try {
			const newVal = !isBundleEnabled;
			await promptStoreAPI.patchPromptBundle(bundle.id, newVal);
			/* update local UI state */
			setIsBundleEnabled(newVal);
		} catch (err) {
			console.error('Failed to toggle bundle:', err);
			setAlertMsg('Failed to toggle bundle enable state.');
			setShowAlert(true);
		}
	};

	/* ---------- template helpers ---------- */
	const patchTemplateEnable = async (tpl: PromptTemplate) => {
		try {
			await promptStoreAPI.patchPromptTemplate(bundle.id, tpl.slug, tpl.version, !tpl.isEnabled);
			const updated: PromptTemplate = { ...tpl, isEnabled: !tpl.isEnabled };
			const newArr = localTemplates.map(t => (t.id === tpl.id ? updated : t));
			setLocalTemplates(newArr);
			onTemplatesChange(bundle.id, newArr);
		} catch (err) {
			console.error('Toggle template failed:', err);
			setAlertMsg('Failed to toggle template.');
			setShowAlert(true);
		}
	};

	/* ----- delete template ----- */
	const requestDeleteTemplate = (tpl: PromptTemplate) => {
		if (tpl.isBuiltIn) {
			setAlertMsg('Cannot delete built-in template.');
			setShowAlert(true);
			return;
		}
		setTemplateToDelete(tpl);
		setIsDeleteTemplateModalOpen(true);
	};

	const confirmDeleteTemplate = async () => {
		if (!templateToDelete) return;
		try {
			await promptStoreAPI.deletePromptTemplate(bundle.id, templateToDelete.slug, templateToDelete.version);
			const newArr = localTemplates.filter(t => t.id !== templateToDelete.id);
			setLocalTemplates(newArr);
			onTemplatesChange(bundle.id, newArr);
		} catch (err) {
			console.error('Delete template failed:', err);
			setAlertMsg('Failed to delete template.');
			setShowAlert(true);
		} finally {
			setIsDeleteTemplateModalOpen(false);
			setTemplateToDelete(null);
		}
	};

	/* ----- add / edit template ----- */
	const openModifyModal = (tpl?: PromptTemplate) => {
		if (tpl?.isBuiltIn) {
			setAlertMsg('Built-in templates cannot be edited.');
			setShowAlert(true);
			return;
		}
		if (bundle.isBuiltIn) {
			setAlertMsg('Cannot add or edit templates in a built-in bundle.');
			setShowAlert(true);
			return;
		}
		setTemplateToEdit(tpl);
		setIsModifyModalOpen(true);
	};

	const handleModifySubmit = async (partial: Partial<PromptTemplate>) => {
		try {
			if (templateToEdit) {
				/* update existing */
				await promptStoreAPI.putPromptTemplate(
					bundle.id,
					templateToEdit.slug,
					partial.displayName ?? templateToEdit.displayName,
					partial.isEnabled ?? templateToEdit.isEnabled,
					partial.blocks ?? templateToEdit.blocks,
					templateToEdit.version, // overwrite same version for now
					partial.description ?? templateToEdit.description,
					partial.tags ?? templateToEdit.tags,
					partial.variables ?? templateToEdit.variables,
					partial.preProcessors ?? templateToEdit.preProcessors
				);
			} else {
				/* create */
				const slug = partial.slug?.trim() ?? '';
				const display = partial.displayName?.trim() ?? '';
				await promptStoreAPI.putPromptTemplate(
					bundle.id,
					slug,
					display,
					partial.isEnabled ?? true,
					partial.blocks ?? [],
					'1',
					partial.description,
					partial.tags,
					partial.variables,
					partial.preProcessors
				);
			}
			/* refresh local list */
			const promptTemplateListItems = await getAllPromptTemplates([bundle.id], undefined, true);
			const tplPromises = promptTemplateListItems.map(li =>
				promptStoreAPI.getPromptTemplate(li.bundleID, li.templateSlug, li.templateVersion)
			);
			const fresh = (await Promise.all(tplPromises)).filter((t): t is PromptTemplate => t !== undefined);
			setLocalTemplates(fresh);
			onTemplatesChange(bundle.id, fresh);
		} catch (err) {
			console.error('Save template failed:', err);
			setAlertMsg('Failed to save template.');
			setShowAlert(true);
		} finally {
			setIsModifyModalOpen(false);
			setTemplateToEdit(undefined);
		}
	};

	/* ---------- render ---------- */
	return (
		<div className="bg-base-100 mb-8 rounded-2xl p-4 shadow-lg">
			{/* header */}
			<div className="grid grid-cols-12 items-center gap-2">
				{/* label + name + slug */}
				<div className="col-span-4 flex items-center gap-2">
					<h3 className="text-sm font-semibold">
						<span className="capitalize">{bundle.displayName || bundle.slug}</span>
						<span className="text-base-content/60 ml-1">({bundle.slug})</span>
					</h3>
				</div>
				<div className="col-span-1 flex items-center gap-2">
					<span className="text-base-content/60 text-xs tracking-wide uppercase">
						{bundle.isBuiltIn ? 'Built-in' : 'Custom'}
					</span>
				</div>
				{/* enabled toggle */}
				<div className="col-span-3 flex items-center gap-2">
					<label className="text-sm">Enabled</label>
					<input
						type="checkbox"
						className="toggle toggle-accent"
						checked={isBundleEnabled}
						onChange={toggleBundleEnable}
					/>
				</div>

				{/* template count */}
				<div className="col-span-2 flex items-center text-sm">
					<span>Templates:&nbsp;{localTemplates.length}</span>
				</div>

				{/* chevron */}
				<div
					className="col-span-2 flex cursor-pointer items-center justify-end gap-1"
					onClick={() => {
						setIsExpanded(p => !p);
					}}
				>
					<label className="text-sm whitespace-nowrap">Templates</label>
					{isExpanded ? <FiChevronUp /> : <FiChevronDown />}
				</div>
			</div>

			{/* body - templates table */}
			{isExpanded && (
				<div className="mt-8 space-y-4">
					{/* table */}
					<div className="border-base-content/10 overflow-x-auto rounded-2xl border">
						<table className="table-zebra table w-full">
							<thead>
								<tr className="bg-base-300 text-sm font-semibold">
									<th>Display Name</th>
									<th>Slug</th>
									<th className="text-center">Enabled</th>
									<th className="text-center">Version</th>
									<th className="text-center">Built-In</th>
									<th className="text-center">Actions</th>
								</tr>
							</thead>
							<tbody>
								{localTemplates.map(tpl => (
									<tr key={tpl.id} className="hover:bg-base-300">
										<td>{tpl.displayName}</td>
										<td>
											{PROMPT_TEMPLATE_INVOKE_CHAR}
											{tpl.slug}
										</td>
										<td className="text-center align-middle">
											<input
												type="checkbox"
												className="toggle toggle-accent"
												checked={tpl.isEnabled}
												onChange={() => patchTemplateEnable(tpl)}
											/>
										</td>
										<td className="text-center">{tpl.version}</td>
										<td className="text-center">
											{tpl.isBuiltIn ? <FiCheck className="mx-auto" /> : <FiX className="mx-auto" />}
										</td>
										<td className="text-center">
											<div className="inline-flex gap-3">
												<button
													className="btn btn-sm btn-ghost rounded-2xl"
													onClick={() => {
														openModifyModal(tpl);
													}}
													disabled={tpl.isBuiltIn || bundle.isBuiltIn}
													title={tpl.isBuiltIn || bundle.isBuiltIn ? 'Editing disabled for built-in items' : 'Edit'}
												>
													<FiEdit2 size={16} />
												</button>

												<button
													className="btn btn-sm btn-ghost rounded-2xl"
													onClick={() => {
														requestDeleteTemplate(tpl);
													}}
													disabled={tpl.isBuiltIn || bundle.isBuiltIn}
													title={tpl.isBuiltIn || bundle.isBuiltIn ? 'Deleting disabled for built-in items' : 'Delete'}
												>
													<FiTrash2 size={16} />
												</button>
											</div>
										</td>
									</tr>
								))}
								{localTemplates.length === 0 && (
									<tr>
										<td colSpan={6} className="py-3 text-center text-sm">
											No templates in this bundle.
										</td>
									</tr>
								)}
							</tbody>
						</table>
					</div>

					{/* bottom-row actions (only for custom bundles) */}
					{!bundle.isBuiltIn && (
						<div className="flex items-center justify-between">
							<button
								className="btn btn-md btn-ghost flex items-center rounded-2xl"
								onClick={() => {
									onBundleDeleted(bundle);
								}}
							>
								<FiTrash2 /> <span className="ml-1">Delete Bundle</span>
							</button>

							<button
								className="btn btn-md btn-ghost flex items-center rounded-2xl"
								onClick={() => {
									openModifyModal(undefined);
								}}
							>
								<FiPlus /> <span className="ml-1">Add Template</span>
							</button>
						</div>
					)}
				</div>
			)}

			{/* dialogs / alerts */}
			<DeleteConfirmationModal
				isOpen={isDeleteTemplateModalOpen}
				onClose={() => {
					setIsDeleteTemplateModalOpen(false);
				}}
				onConfirm={confirmDeleteTemplate}
				title="Delete Prompt Template"
				message={`Delete template "${templateToDelete?.displayName ?? ''}"? This cannot be undone.`}
				confirmButtonText="Delete"
			/>

			<AddEditPromptTemplate
				isOpen={isModifyModalOpen}
				onClose={() => {
					setIsModifyModalOpen(false);
					setTemplateToEdit(undefined);
				}}
				onSubmit={handleModifySubmit}
				initialData={
					templateToEdit
						? {
								template: templateToEdit,
								bundleID: bundle.id,
								templateSlug: templateToEdit.slug,
							}
						: undefined
				}
				existingTemplates={localTemplates.map(t => ({
					template: t,
					bundleID: bundle.id,
					templateSlug: t.slug,
				}))}
			/>

			<ActionDeniedAlert
				isOpen={showAlert}
				onClose={() => {
					setShowAlert(false);
					setAlertMsg('');
				}}
				message={alertMsg}
			/>
		</div>
	);
};

export default PromptBundleCard;
