'use client';

import { PROMPT_TEMPLATE_INVOKE_CHAR } from '@/models/commands';
import { useEffect, useState } from 'react';
import { FiCheck, FiEdit, FiPlus, FiTrash2, FiX } from 'react-icons/fi';

interface PromptTemplate {
	id: string;
	name: string;
	command: string;
	hasTools: boolean;
	hasDocStore: boolean;
	tokenCount: number;
}

const fetchPromptTemplates = async (): Promise<PromptTemplate[]> => {
	await new Promise(resolve => setTimeout(resolve, 10));

	return [
		{ id: '1', name: 'General Query', command: 'gen', hasTools: true, hasDocStore: true, tokenCount: 150 },
		{ id: '2', name: 'Summarization', command: 'summary', hasTools: false, hasDocStore: true, tokenCount: 200 },
		{ id: '3', name: 'Code Explanation', command: 'codeexplain', hasTools: true, hasDocStore: false, tokenCount: 180 },
	];
};

const PromptTemplates: React.FC = () => {
	const [templates, setTemplates] = useState<PromptTemplate[]>([]);
	const [loading, setLoading] = useState(true);

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
				<button className="btn btn-ghost rounded-2xl text-sm">
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
							<tr key={template.id} className="hover border-none shadow-none">
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
									<button className="btn btn-sm btn-ghost rounded-2xl" aria-label="Edit Template">
										<FiEdit />
									</button>
									<button className="btn btn-sm btn-ghost rounded-2xl" aria-label="Delete Template">
										<FiTrash2 />
									</button>
								</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</div>
	);
};

export default PromptTemplates;
