import { useEffect, useState } from 'react';

import { FiEdit, FiPlus, FiTrash2 } from 'react-icons/fi';

import { TOOL_INVOKE_CHAR } from '@/models/commands';
import type { Tool } from '@/models/toolmodel';

import DeleteConfirmationModal from '@/components/delete_confirmation';

import ModifyTool from '@/prompts/tool_modify';

const fetchTools = async (): Promise<Tool[]> => {
	await new Promise(resolve => setTimeout(resolve, 100));

	return [
		{
			id: '1',
			name: 'Web Search',
			command: 'search',
			schema: 'search schema',
			inFunc: 'def searchFunc',
			tokenCount: 50,
		},
		{ id: '2', name: 'Calculator', command: 'calc', schema: 'calc schema', inFunc: 'def calcFunc', tokenCount: 30 },
		{
			id: '3',
			name: 'Weather',
			command: 'weather',
			schema: 'weather schema',
			inFunc: 'def weatherFunc',
			tokenCount: 40,
		},
	];
};

const Tools: React.FC = () => {
	const [tools, setTools] = useState<Tool[]>([]);
	const [loading, setLoading] = useState(true);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [toolToDelete, setToolToDelete] = useState<Tool | null>(null);
	const [isModifyModalOpen, setIsModifyModalOpen] = useState(false);
	const [toolToModify, setToolToModify] = useState<Tool | null>(null);

	useEffect(() => {
		const loadTools = async () => {
			try {
				const data = await fetchTools();
				setTools(data);
			} catch (error) {
				console.error('Failed to fetch tools:', error);
			} finally {
				setLoading(false);
			}
		};

		loadTools();
	}, []);

	const openDeleteToolModal = (tool: Tool) => {
		setToolToDelete(tool);
		setIsDeleteModalOpen(true);
	};

	const closeDeleteToolModal = () => {
		setIsDeleteModalOpen(false);
		setToolToDelete(null);
	};

	const handleDeleteTool = () => {
		if (toolToDelete) {
			setTools(prevTools => prevTools.filter(tool => tool.id !== toolToDelete.id));
		}
		closeDeleteToolModal();
	};

	const openModifyToolModal = (tool?: Tool) => {
		setToolToModify(tool || null);
		setIsModifyModalOpen(true);
	};

	const closeModifyToolModal = () => {
		setIsModifyModalOpen(false);
		setToolToModify(null);
	};

	const handleModifyTool = (toolData: Partial<Tool>) => {
		if (toolToModify) {
			// Edit existing tool
			setTools(prevTools => prevTools.map(tool => (tool.id === toolToModify.id ? { ...tool, ...toolData } : tool)));
		} else {
			// Add new tool
			const newTool: Tool = {
				...toolData,
				id: Date.now().toString(),
				tokenCount: 0,
			} as Tool;
			setTools(prevTools => [...prevTools, newTool]);
		}
		closeModifyToolModal();
	};

	if (loading) {
		return (
			<div className="flex justify-center items-center h-screen">
				<span className="loading loading-spinner loading-sm"></span>
			</div>
		);
	}

	return (
		<div className="mx-auto p-4">
			<div className="text-right items-center mb-2">
				<button
					className="btn btn-ghost rounded-2xl text-sm"
					onClick={() => {
						openModifyToolModal();
					}}
				>
					<FiPlus size={18} /> Add Tool
				</button>
			</div>
			{/* Tools Table */}
			<div className="overflow-x-auto border border-base-content/10 rounded-2xl">
				<table className="table table-zebra w-full">
					<thead>
						<tr className="font-semibold text-sm px-4 py-0 m-0 bg-base-300">
							<th>Tool Name</th>
							<th>Slug</th>
							<th className="text-center">Token Count</th>
							<th className="text-right pr-8">Actions</th>
						</tr>
					</thead>
					<tbody>
						{tools.map(tool => (
							<tr key={tool.id} className="hover:bg-base-300">
								<td>{tool.name}</td>
								<td>
									{TOOL_INVOKE_CHAR}
									{tool.command}
								</td>
								<td className="text-center">{tool.tokenCount}</td>
								<td className="text-right">
									<button
										className="btn btn-sm btn-ghost rounded-2xl"
										aria-label="Edit Tool"
										title="Edit Tool"
										onClick={() => {
											openModifyToolModal(tool);
										}}
									>
										<FiEdit />
									</button>
									<button
										className="btn btn-sm btn-ghost rounded-2xl"
										aria-label="Delete Tool"
										title="Delete Tool"
										onClick={() => {
											openDeleteToolModal(tool);
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
				onClose={closeDeleteToolModal}
				onConfirm={handleDeleteTool}
				title="Delete Tool"
				message={`Are you sure you want to delete the tool "${toolToDelete?.name || ''}"? This action cannot be undone.`}
				confirmButtonText="Delete"
			/>

			{/* Modify Tool Modal */}
			<ModifyTool
				isOpen={isModifyModalOpen}
				onClose={closeModifyToolModal}
				onSubmit={handleModifyTool}
				initialData={toolToModify || undefined}
				existingTools={tools}
			/>
		</div>
	);
};

export default Tools;
