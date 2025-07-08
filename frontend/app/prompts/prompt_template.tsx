import { useEffect, useState } from 'react';

import { FiCheck, FiEdit, FiPlus, FiTrash2, FiX } from 'react-icons/fi';

import { PROMPT_TEMPLATE_INVOKE_CHAR } from '@/models/commands';
import { PromptRoleEnum, type PromptTemplate } from '@/models/promptmodel';

import { promptStoreAPI } from '@/apis/baseapi';

import { getUUIDv7 } from '@/lib/uuid_utils';

import DeleteConfirmationModal from '@/components/delete_confirmation';

import ModifyPromptTemplate from '@/prompts/prompt_template_modify';

/* ---------- Local types ---------- */

interface TemplateItem {
	template: PromptTemplate; // full template returned by `getPromptTemplate`
	bundleID: string; // needed for mutations
	templateSlug: string; // shortcut so we don’t re-derive it
}

const DEFAULT_BUNDLE_ID = 'default-bundle'; // change if you expose bundles in the UI

/* ---------- Component ---------- */

const PromptTemplates: React.FC = () => {
	/* --- state --- */
	const [templates, setTemplates] = useState<TemplateItem[]>([]);
	const [loading, setLoading] = useState(true);

	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [templateToDelete, setTemplateToDelete] = useState<TemplateItem | null>(null);

	const [isModifyModalOpen, setIsModifyModalOpen] = useState(false);
	const [templateToEdit, setTemplateToEdit] = useState<TemplateItem | undefined>(undefined);

	/* --- helpers --- */
	const fetchTemplates = async () => {
		setLoading(true);
		try {
			const { promptTemplateListItems } = await promptStoreAPI.listPromptTemplates();
			const detailPromises = promptTemplateListItems.map(async item => {
				try {
					const t = await promptStoreAPI.getPromptTemplate(
						item.bundleID,
						item.templateSlug,
						item.templateVersion // latest per list-item
					);
					if (t) {
						return { template: t, bundleID: item.bundleID, templateSlug: item.templateSlug };
					}
				} catch {
					/* ignore single fetch error – keep the UI responsive */
				}
				return undefined;
			});

			const result = (await Promise.all(detailPromises)).filter((v): v is TemplateItem => v !== undefined);
			setTemplates(result);
		} catch (err) {
			console.error('Failed to fetch templates:', err);
		} finally {
			setLoading(false);
		}
	};

	/* --- lifecycle --- */
	useEffect(() => {
		fetchTemplates();
	}, []);

	/* --- delete --- */
	const openDeletePromptTemplateModal = (item: TemplateItem) => {
		setTemplateToDelete(item);
		setIsDeleteModalOpen(true);
	};

	const closeDeletePromptTemplateModal = () => {
		setIsDeleteModalOpen(false);
		setTemplateToDelete(null);
	};

	const handleDeleteTemplate = async () => {
		if (!templateToDelete) return;

		try {
			await promptStoreAPI.deletePromptTemplate(
				templateToDelete.bundleID,
				templateToDelete.templateSlug,
				templateToDelete.template.version
			);
			setTemplates(prev => prev.filter(t => t.template.id !== templateToDelete.template.id));
		} catch (err) {
			console.error('Delete failed:', err);
		} finally {
			closeDeletePromptTemplateModal();
		}
	};

	/* --- add / edit --- */
	const openModifyPromptTemplateModal = (item?: TemplateItem) => {
		setTemplateToEdit(item);
		setIsModifyModalOpen(true);
	};

	const closeModifyPromptTemplateModal = () => {
		setIsModifyModalOpen(false);
		setTemplateToEdit(undefined);
	};

	const handleModifyTemplate = async (partial: Partial<PromptTemplate>) => {
		try {
			if (templateToEdit) {
				/* ---- UPDATE ---- */
				await promptStoreAPI.putPromptTemplate(
					templateToEdit.bundleID,
					templateToEdit.templateSlug,
					partial.displayName ?? templateToEdit.template.displayName,
					partial.isEnabled ?? templateToEdit.template.isEnabled,
					partial.blocks ?? templateToEdit.template.blocks,
					templateToEdit.template.version,
					partial.description ?? templateToEdit.template.description,
					partial.tags ?? templateToEdit.template.tags,
					partial.variables ?? templateToEdit.template.variables,
					partial.preProcessors ?? templateToEdit.template.preProcessors
				);
			} else {
				/* ---- CREATE ---- */
				const slug = partial.slug ? partial.slug.trim() : '';
				const displayName = partial.displayName ? partial.displayName.trim() : '';
				await promptStoreAPI.putPromptTemplate(
					DEFAULT_BUNDLE_ID,
					slug,
					displayName,
					partial.isEnabled ?? true,
					partial.blocks ?? [
						{
							id: getUUIDv7(),
							role: PromptRoleEnum.User,
							content: partial.description ?? '',
						},
					],
					'1', // first version
					partial.description,
					partial.tags,
					partial.variables,
					partial.preProcessors
				);
			}
			await fetchTemplates();
		} catch (err) {
			console.error('Create / update failed:', err);
		} finally {
			closeModifyPromptTemplateModal();
		}
	};

	/* ---------- render ---------- */
	if (loading) {
		return (
			<div className="flex justify-center items-center h-screen">
				<span className="loading loading-spinner loading-sm" />
			</div>
		);
	}

	return (
		<div className="mx-auto p-4">
			{/* Toolbar */}
			<div className="text-right mb-2">
				<button
					className="btn btn-ghost rounded-2xl text-sm"
					onClick={() => {
						openModifyPromptTemplateModal();
					}}
				>
					<FiPlus size={18} /> Add Prompt Template
				</button>
			</div>

			{/* Table */}
			<div className="overflow-x-auto">
				<table className="table table-zebra w-full">
					<thead>
						<tr className="font-semibold text-sm bg-base-300">
							<th className="rounded-tl-2xl">Display Name</th>
							<th>Slug</th>
							<th>Enabled</th>
							<th>Version</th>
							<th>Modified</th>
							<th className="text-right rounded-tr-2xl pr-8">Actions</th>
						</tr>
					</thead>
					<tbody>
						{templates.map((item, idx) => (
							<tr key={item.template.id} className="hover:bg-base-300 border-none shadow-none">
								<td className={idx === templates.length - 1 ? 'rounded-bl-2xl' : ''}>{item.template.displayName}</td>
								<td>
									{PROMPT_TEMPLATE_INVOKE_CHAR}
									{item.template.slug}
								</td>
								<td>
									{item.template.isEnabled ? (
										<FiCheck className="text-success" size={24} />
									) : (
										<FiX className="text-error" size={24} />
									)}
								</td>
								<td>{item.template.version}</td>
								<td>{new Date(item.template.modifiedAt).toLocaleString()}</td>
								<td className={idx === templates.length - 1 ? 'rounded-br-2xl text-right' : 'text-right'}>
									<button
										className="btn btn-sm btn-ghost rounded-2xl"
										aria-label="Edit"
										title="Edit"
										onClick={() => {
											openModifyPromptTemplateModal(item);
										}}
									>
										<FiEdit />
									</button>
									<button
										className="btn btn-sm btn-ghost rounded-2xl"
										aria-label="Delete"
										title="Delete"
										onClick={() => {
											openDeletePromptTemplateModal(item);
										}}
									>
										<FiTrash2 />
									</button>
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>

			{/* --- Delete modal --- */}
			<DeleteConfirmationModal
				isOpen={isDeleteModalOpen}
				onClose={closeDeletePromptTemplateModal}
				onConfirm={handleDeleteTemplate}
				title="Delete Prompt Template"
				message={`Are you sure you want to delete "${templateToDelete?.template.displayName ?? ''}"? This action cannot be undone.`}
				confirmButtonText="Delete"
			/>

			{/* --- Modify / create modal --- */}
			<ModifyPromptTemplate
				isOpen={isModifyModalOpen}
				onClose={closeModifyPromptTemplateModal}
				onSubmit={handleModifyTemplate}
				initialData={templateToEdit}
				existingTemplates={templates}
			/>
		</div>
	);
};

export default PromptTemplates;
