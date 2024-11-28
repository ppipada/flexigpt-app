'use client';

import DeleteConfirmationModal from '@/components/delete_confirmation';
import { TOOL_INVOKE_CHAR } from '@/models/commands';
import { Tool } from '@/models/promptmodel';
import { useEffect, useState } from 'react';
import { FiEdit, FiPlus, FiTrash2 } from 'react-icons/fi';

const fetchTools = async (): Promise<Tool[]> => {
	await new Promise(resolve => setTimeout(resolve, 100));

	return [
		{ id: '1', name: 'Web Search', command: 'search', tokenCount: 50 },
		{ id: '2', name: 'Calculator', command: 'calc', tokenCount: 30 },
		{ id: '3', name: 'Weather', command: 'weather', tokenCount: 40 },
	];
};

const Tools: React.FC = () => {
	const [tools, setTools] = useState<Tool[]>([]);
	const [loading, setLoading] = useState(true);
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
	const [toolToDelete, setToolToDelete] = useState<Tool | null>(null);

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

	// Function to open the delete confirmation modal
	const openDeleteToolModal = (tool: Tool) => {
		setToolToDelete(tool);
		setIsDeleteModalOpen(true);
	};

	// Function to close the delete confirmation modal
	const closeDeleteToolModal = () => {
		setIsDeleteModalOpen(false);
		setToolToDelete(null);
	};

	// Function to handle the tool deletion
	const handleDeleteTool = () => {
		if (toolToDelete) {
			// Perform the actual deletion here
			setTools(prevTools => prevTools.filter(tool => tool.id !== toolToDelete.id));
		}
		closeDeleteToolModal();
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
			<div className="text-right items-center mb-2">
				<button className="btn btn-ghost rounded-2xl text-sm">
					<FiPlus size={18} /> Add Tool
				</button>
			</div>
			{/* Tools Table */}
			<div className="overflow-x-auto">
				<table className="table table-zebra w-full">
					<thead>
						<tr className="font-semibold text-sm px-4 py-0 m-0 bg-base-300">
							<th className="rounded-tl-2xl">Tool Name</th>
							<th>Use Command</th>
							<th>Token Count</th>
							<th className="text-right rounded-tr-2xl pr-8">Actions</th>
						</tr>
					</thead>
					<tbody>
						{tools.map((tool, index) => (
							<tr key={tool.id} className="hover border-none shadow-none">
								<td className={index === tools.length - 1 ? 'rounded-bl-2xl' : ''}>{tool.name}</td>
								<td>
									{TOOL_INVOKE_CHAR}
									{tool.command}
								</td>
								<td>{tool.tokenCount}</td>
								<td className={index === tools.length - 1 ? 'rounded-br-2xl text-right' : 'text-right'}>
									<button className="btn btn-sm btn-ghost rounded-2xl" aria-label="Edit Tool">
										<FiEdit />
									</button>
									<button
										className="btn btn-sm btn-ghost rounded-2xl"
										aria-label="Delete Tool"
										onClick={() => openDeleteToolModal(tool)}
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
				message={`Are you sure you want to delete the tool "${toolToDelete?.name}"? This action cannot be undone.`}
				confirmButtonText="Delete"
			/>
		</div>
	);
};

export default Tools;
