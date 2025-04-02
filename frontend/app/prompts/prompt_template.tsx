import DeleteConfirmationModal from '@/components/delete_confirmation';
import { PROMPT_TEMPLATE_INVOKE_CHAR } from '@/models/commands';
import type { PromptTemplate } from '@/models/promptmodel';
import ModifyPromptTemplate from '@/prompts/prompt_template_modify';
import { useEffect, useState } from 'react';
import { FiCheck, FiEdit, FiPlus, FiTrash2, FiX } from 'react-icons/fi';

const fetchPromptTemplates = async (): Promise<PromptTemplate[]> => {
	await new Promise(resolve => setTimeout(resolve, 10));

	return [
		{
			id: '1',
			name: 'General Query',
			command: 'gen',
			template: 'give me a answer',
			hasTools: true,
			hasDocStore: true,
			tokenCount: 150,
		},
		{
			id: '2',
			name: 'Summarization',
			command: 'summary',
			template: 'summarize the input',
			hasTools: false,
			hasDocStore: true,
			tokenCount: 200,
		},
		{
			id: '3',
			name: 'Code Explanation',
			command: 'codeexplain',
			template: 'explain the code',
			hasTools: true,
			hasDocStore: false,
			tokenCount: 180,
		},
	];
};

const PromptTemplates: React.FC = () => {
	const [templates, setTemplates] = useState<PromptTemplate[]>([]);
	const [loading, setLoading] = useState(true);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [templateToDelete, setTemplateToDelete] = useState<PromptTemplate | null>(null);

	const [isModifyModalOpen, setIsModifyModalOpen] = useState(false);
	const [templateToEdit, setTemplateToEdit] = useState<Partial<PromptTemplate> | undefined>(undefined);

	useEffect(() => {
		const loadTemplates = async () => {
			try {
				const data = await fetchPromptTemplates();
				setTemplates(data);
			} catch (error) {
				console.error('Failed to fetch prompt templates:', error);
			} finally {
				setLoading(false);
			}
		};

		loadTemplates();
	}, []);

	// Function to open the delete confirmation modal
	const openDeletePromptTemplateModal = (template: PromptTemplate) => {
		setTemplateToDelete(template);
		setIsDeleteModalOpen(true);
	};

	// Function to close the delete confirmation modal
	const closeDeletePromptTemplateModal = () => {
		setIsDeleteModalOpen(false);
		setTemplateToDelete(null);
	};

	// Function to handle the template deletion
	const handleDeleteTemplate = () => {
		if (templateToDelete) {
			// Perform the actual deletion here
			setTemplates(prevTemplates => prevTemplates.filter(template => template.id !== templateToDelete.id));
		}
		closeDeletePromptTemplateModal();
	};

	const openModifyPromptTemplateModal = (template?: PromptTemplate) => {
		setTemplateToEdit(template);
		setIsModifyModalOpen(true);
	};

	const closeModifyPromptTemplateModal = () => {
		setIsModifyModalOpen(false);
		setTemplateToEdit(undefined);
	};

	const handleModifyTemplate = (templateData: Partial<PromptTemplate>) => {
		if (templateToEdit?.id) {
			// Edit existing template
			setTemplates(prevTemplates =>
				prevTemplates.map(t => (t.id === templateToEdit.id ? { ...t, ...templateData } : t))
			);
		} else {
			// Add new template
			const newTemplate: PromptTemplate = {
				...(templateData as PromptTemplate),
			};
			newTemplate.id = Date.now().toString();
			setTemplates(prevTemplates => [...prevTemplates, newTemplate]);
		}
		closeModifyPromptTemplateModal();
	};

	if (loading) {
		return (
			<div className="flex justify-center items-center h-screen">
				<span className="loading loading-spinner loading-lg"></span>
			</div>
		);
	}

	return (
		<div className="mx-auto p-4">
			{/* Heading Row */}
			<div className="text-right items-center mb-2">
				<button
					className="btn btn-ghost rounded-2xl text-sm"
					onClick={() => {
						openModifyPromptTemplateModal();
					}}
				>
					<FiPlus size={18} /> Add Prompt Templates
				</button>
			</div>

			{/* Prompt Templates Table */}
			<div className="overflow-x-auto">
				<table className="table table-zebra w-full">
					<thead>
						<tr className="font-semibold text-sm px-4 py-0 m-0 bg-base-300">
							<th className="rounded-tl-2xl">Template Name</th>
							<th>Use Command</th>
							<th>Tools</th>
							<th>Document Stores</th>
							<th>Token Count</th>
							<th className="text-right rounded-tr-2xl pr-8">Actions</th>
						</tr>
					</thead>
					<tbody>
						{templates.map((template, index) => (
							<tr key={template.id} className="hover:bg-base-300 border-none shadow-none">
								<td className={index === templates.length - 1 ? 'rounded-bl-2xl' : ''}>{template.name}</td>
								<td>
									{PROMPT_TEMPLATE_INVOKE_CHAR}
									{template.command}
								</td>
								<td>
									{template.hasTools ? (
										<FiCheck className="text-success" size={24} />
									) : (
										<FiX className="text-error" size={24} />
									)}
								</td>
								<td>
									{template.hasDocStore ? (
										<FiCheck className="text-success" size={24} />
									) : (
										<FiX className="text-error" size={24} />
									)}
								</td>
								<td>{template.tokenCount}</td>
								<td className={index === templates.length - 1 ? 'rounded-br-2xl text-right' : 'text-right'}>
									<button
										className="btn btn-sm btn-ghost rounded-2xl"
										aria-label="Edit Template"
										onClick={() => {
											openModifyPromptTemplateModal(template);
										}}
									>
										<FiEdit />
									</button>
									<button
										className="btn btn-sm btn-ghost rounded-2xl"
										aria-label="Delete Template"
										onClick={() => {
											openDeletePromptTemplateModal(template);
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

			{/* Delete Confirmation Modal */}
			<DeleteConfirmationModal
				isOpen={isDeleteModalOpen}
				onClose={closeDeletePromptTemplateModal}
				onConfirm={handleDeleteTemplate}
				title="Delete Prompt Template"
				message={`Are you sure you want to delete the prompt template "${templateToDelete?.name || ''}"? This action cannot be undone.`}
				confirmButtonText="Delete"
			/>

			{/* Modify Prompt Template Modal */}
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
