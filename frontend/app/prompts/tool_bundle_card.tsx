import { useEffect, useState } from 'react';

import { FiCheck, FiChevronDown, FiChevronUp, FiEdit, FiPlus, FiTrash2, FiX } from 'react-icons/fi';

import { TOOL_INVOKE_CHAR } from '@/models/commands';
// You need to implement this

import { type Tool, type ToolBundle, ToolType } from '@/models/toolmodel';

import { toolStoreAPI } from '@/apis/baseapi';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';

import ToolModifyModal from './tool_modify';

interface ToolBundleCardProps {
	bundle: ToolBundle;
	tools: Tool[];
	onToolsChange: (bundleID: string, newTools: Tool[]) => void;
	onBundleDeleted: (bundle: ToolBundle) => void;
}

const ToolBundleCard: React.FC<ToolBundleCardProps> = ({ bundle, tools, onToolsChange, onBundleDeleted }) => {
	const [isExpanded, setIsExpanded] = useState(false);
	const [localTools, setLocalTools] = useState<Tool[]>(tools);
	const [isBundleEnabled, setIsBundleEnabled] = useState(bundle.isEnabled);

	useEffect(() => {
		setIsBundleEnabled(bundle.isEnabled);
	}, [bundle.isEnabled]);
	useEffect(() => {
		setLocalTools(tools);
	}, [tools]);

	const [isDeleteToolModalOpen, setIsDeleteToolModalOpen] = useState(false);
	const [toolToDelete, setToolToDelete] = useState<Tool | null>(null);

	const [isModifyModalOpen, setIsModifyModalOpen] = useState(false);
	const [toolToEdit, setToolToEdit] = useState<Tool | undefined>(undefined);

	const [showAlert, setShowAlert] = useState(false);
	const [alertMsg, setAlertMsg] = useState('');

	// Enable/disable bundle
	const toggleBundleEnable = async () => {
		try {
			const newVal = !isBundleEnabled;
			await toolStoreAPI.patchToolBundle(bundle.id, newVal);
			setIsBundleEnabled(newVal);
		} catch (err) {
			console.error('Toggle bundle enable failed:', err);
			setAlertMsg('Failed to toggle bundle enable state.');
			setShowAlert(true);
		}
	};

	// Enable/disable tool
	const patchToolEnable = async (tool: Tool) => {
		try {
			await toolStoreAPI.patchTool(bundle.id, tool.slug, tool.version, !tool.isEnabled);
			const updated: Tool = { ...tool, isEnabled: !tool.isEnabled };
			const newArr = localTools.map(t => (t.id === tool.id ? updated : t));
			setLocalTools(newArr);
			onToolsChange(bundle.id, newArr);
		} catch (err) {
			console.error('Toggle tool failed:', err);
			setAlertMsg('Failed to toggle tool.');
			setShowAlert(true);
		}
	};

	// Delete tool
	const requestDeleteTool = (tool: Tool) => {
		if (tool.isBuiltIn) {
			setAlertMsg('Cannot delete built-in tool.');
			setShowAlert(true);
			return;
		}
		setToolToDelete(tool);
		setIsDeleteToolModalOpen(true);
	};

	const confirmDeleteTool = async () => {
		if (!toolToDelete) return;
		try {
			await toolStoreAPI.deleteTool(bundle.id, toolToDelete.slug, toolToDelete.version);
			const newArr = localTools.filter(t => t.id !== toolToDelete.id);
			setLocalTools(newArr);
			onToolsChange(bundle.id, newArr);
		} catch (err) {
			console.error('Delete tool failed:', err);
			setAlertMsg('Failed to delete tool.');
			setShowAlert(true);
		} finally {
			setIsDeleteToolModalOpen(false);
			setToolToDelete(null);
		}
	};

	// Add/edit tool
	const openModifyModal = (tool?: Tool) => {
		if (tool?.isBuiltIn) {
			setAlertMsg('Built-in tools cannot be edited.');
			setShowAlert(true);
			return;
		}
		if (bundle.isBuiltIn) {
			setAlertMsg('Cannot add or edit tools in a built-in bundle.');
			setShowAlert(true);
			return;
		}
		setToolToEdit(tool);
		setIsModifyModalOpen(true);
	};

	const handleModifySubmit = async (partial: Partial<Tool>) => {
		try {
			if (toolToEdit) {
				// update existing
				await toolStoreAPI.putTool(
					bundle.id,
					toolToEdit.slug,
					toolToEdit.version,
					partial.displayName ?? toolToEdit.displayName,
					partial.isEnabled ?? toolToEdit.isEnabled,
					partial.argSchema ?? toolToEdit.argSchema,
					partial.outputSchema ?? toolToEdit.outputSchema,
					partial.type ?? toolToEdit.type,
					partial.goImpl ?? toolToEdit.goImpl,
					partial.httpImpl ?? toolToEdit.httpImpl,
					partial.description ?? toolToEdit.description,
					partial.tags ?? toolToEdit.tags
				);
			} else {
				// create
				const slug = partial.slug?.trim() ?? '';
				const display = partial.displayName?.trim() ?? '';
				await toolStoreAPI.putTool(
					bundle.id,
					slug,
					'1',
					display,
					partial.isEnabled ?? true,
					partial.argSchema ?? {},
					partial.outputSchema ?? {},
					partial.type ?? ToolType.HTTP,
					partial.goImpl,
					partial.httpImpl,
					partial.description,
					partial.tags
				);
			}
			// refresh local list
			const { toolListItems } = await toolStoreAPI.listTools([bundle.id], undefined, true);
			const toolPromises = toolListItems.map(li => toolStoreAPI.getTool(li.bundleID, li.toolSlug, li.toolVersion));
			const fresh = (await Promise.all(toolPromises)).filter((t): t is Tool => t !== undefined);
			setLocalTools(fresh);
			onToolsChange(bundle.id, fresh);
		} catch (err) {
			console.error('Toool save failed:', err);
			setAlertMsg('Failed to save tool.');
			setShowAlert(true);
		} finally {
			setIsModifyModalOpen(false);
			setToolToEdit(undefined);
		}
	};

	return (
		<div className="bg-base-100 rounded-2xl shadow-lg p-4 mb-8">
			{/* header */}
			<div className="grid grid-cols-12 gap-2 items-center">
				<div className="col-span-4 flex items-center gap-2">
					<h3 className="text-sm font-semibold">
						<span className="capitalize">{bundle.displayName || bundle.slug}</span>
						<span className="text-base-content/60 ml-1">({bundle.slug})</span>
					</h3>
				</div>
				<div className="col-span-1 flex items-center gap-2">
					<span className="text-xs uppercase tracking-wide text-base-content/60">
						{bundle.isBuiltIn ? 'Built-in' : 'Custom'}
					</span>
				</div>
				<div className="col-span-3 flex items-center gap-2">
					<label className="text-sm">Enabled</label>
					<input
						type="checkbox"
						className="toggle toggle-accent"
						checked={isBundleEnabled}
						onChange={toggleBundleEnable}
					/>
				</div>
				<div className="col-span-2 flex items-center text-sm">
					<span>Tools:&nbsp;{localTools.length}</span>
				</div>
				<div
					className="col-span-2 flex justify-end items-center cursor-pointer gap-1"
					onClick={() => {
						setIsExpanded(p => !p);
					}}
				>
					<label className="text-sm whitespace-nowrap">Tools</label>
					{isExpanded ? <FiChevronUp /> : <FiChevronDown />}
				</div>
			</div>

			{/* body - tools table */}
			{isExpanded && (
				<div className="mt-8 space-y-4">
					<div className="overflow-x-auto border border-base-content/10 rounded-2xl">
						<table className="table table-zebra w-full">
							<thead>
								<tr className="text-sm font-semibold bg-base-300">
									<th>Display Name</th>
									<th>Slug</th>
									<th className="text-center">Enabled</th>
									<th className="text-center">Version</th>
									<th className="text-center">Built-In</th>
									<th className="text-right pr-8">Actions</th>
								</tr>
							</thead>
							<tbody>
								{localTools.map(tool => (
									<tr key={tool.id} className="hover:bg-base-300">
										<td>{tool.displayName}</td>
										<td>
											{TOOL_INVOKE_CHAR}
											{tool.slug}
										</td>
										<td className="text-center align-middle">
											<input
												type="checkbox"
												className="toggle toggle-accent"
												checked={tool.isEnabled}
												onChange={() => patchToolEnable(tool)}
											/>
										</td>
										<td className="text-center">{tool.version}</td>
										<td className="text-center">
											{tool.isBuiltIn ? <FiCheck className="mx-auto" /> : <FiX className="mx-auto" />}
										</td>
										<td className="text-right">
											<div className="inline-flex gap-2">
												<button
													className="btn btn-sm btn-ghost rounded-2xl"
													onClick={() => {
														openModifyModal(tool);
													}}
													disabled={tool.isBuiltIn || bundle.isBuiltIn}
													title={tool.isBuiltIn || bundle.isBuiltIn ? 'Editing disabled for built-in items' : 'Edit'}
												>
													<FiEdit size={16} />
												</button>
												<button
													className="btn btn-sm btn-ghost rounded-2xl"
													onClick={() => {
														requestDeleteTool(tool);
													}}
													disabled={tool.isBuiltIn || bundle.isBuiltIn}
													title={tool.isBuiltIn || bundle.isBuiltIn ? 'Deleting disabled for built-in items' : 'Delete'}
												>
													<FiTrash2 size={16} />
												</button>
											</div>
										</td>
									</tr>
								))}
								{localTools.length === 0 && (
									<tr>
										<td colSpan={6} className="text-center py-3 text-sm">
											No tools in this bundle.
										</td>
									</tr>
								)}
							</tbody>
						</table>
					</div>
					{!bundle.isBuiltIn && (
						<div className="flex justify-between items-center">
							<button
								className="btn btn-md btn-ghost rounded-2xl flex items-center"
								onClick={() => {
									onBundleDeleted(bundle);
								}}
							>
								<FiTrash2 /> <span className="ml-1">Delete Bundle</span>
							</button>
							<button
								className="btn btn-md btn-ghost rounded-2xl flex items-center"
								onClick={() => {
									openModifyModal(undefined);
								}}
							>
								<FiPlus /> <span className="ml-1">Add Tool</span>
							</button>
						</div>
					)}
				</div>
			)}

			{/* dialogs / alerts */}
			<DeleteConfirmationModal
				isOpen={isDeleteToolModalOpen}
				onClose={() => {
					setIsDeleteToolModalOpen(false);
				}}
				onConfirm={confirmDeleteTool}
				title="Delete Tool"
				message={`Delete tool "${toolToDelete?.displayName ?? ''}"? This cannot be undone.`}
				confirmButtonText="Delete"
			/>

			<ToolModifyModal
				isOpen={isModifyModalOpen}
				onClose={() => {
					setIsModifyModalOpen(false);
					setToolToEdit(undefined);
				}}
				onSubmit={handleModifySubmit}
				initialData={
					toolToEdit
						? {
								tool: toolToEdit,
								bundleID: bundle.id,
								toolSlug: toolToEdit.slug,
							}
						: undefined
				}
				existingTools={localTools.map(t => ({
					tool: t,
					bundleID: bundle.id,
					toolSlug: t.slug,
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

export default ToolBundleCard;
