import { useEffect, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import type { Tool, ToolBundle } from '@/spec/tool';

// Your IToolStoreAPI instance
import { getUUIDv7 } from '@/lib/uuid_utils';

import { toolStoreAPI } from '@/apis/baseapi';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';

import AddToolBundleModal from '@/prompts/tool_bundle_add';
import ToolBundleCard from '@/prompts/tool_bundle_card';

interface BundleData {
	bundle: ToolBundle;
	tools: Tool[];
}

const ToolsPage: React.FC = () => {
	const [bundles, setBundles] = useState<BundleData[]>([]);
	const [loading, setLoading] = useState(true);

	const [showAlert, setShowAlert] = useState(false);
	const [alertMsg, setAlertMsg] = useState('');

	const [bundleToDelete, setBundleToDelete] = useState<ToolBundle | null>(null);
	const [isAddModalOpen, setIsAddModalOpen] = useState(false);

	// Fetch all bundles and their tools
	const fetchAll = async () => {
		setLoading(true);
		try {
			const { toolBundles } = await toolStoreAPI.listToolBundles(undefined, true);
			const bundleResults: BundleData[] = await Promise.all(
				toolBundles.map(async b => {
					try {
						const { toolListItems } = await toolStoreAPI.listTools([b.id], undefined, true);
						const toolPromises = toolListItems.map(itm =>
							toolStoreAPI.getTool(itm.bundleID, itm.toolSlug, itm.toolVersion)
						);
						const tools = (await Promise.all(toolPromises)).filter((t): t is Tool => t !== undefined);
						return { bundle: b, tools };
					} catch {
						return { bundle: b, tools: [] };
					}
				})
			);
			setBundles(bundleResults);
		} catch (err) {
			console.error('Load tool bundles failed:', err);
			setAlertMsg('Failed to load tool bundles. Please try again.');
			setShowAlert(true);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		fetchAll();
	}, []);

	const onToolsChange = (bundleID: string, newTools: Tool[]) => {
		setBundles(prev => prev.map(bd => (bd.bundle.id === bundleID ? { ...bd, tools: newTools } : bd)));
	};

	const handleBundleDelete = async () => {
		if (!bundleToDelete) return;
		try {
			await toolStoreAPI.deleteToolBundle(bundleToDelete.id);
			setBundles(prev => prev.filter(bd => bd.bundle.id !== bundleToDelete.id));
		} catch (err) {
			console.error('Delete tool bundle failed:', err);
			setAlertMsg('Failed to delete tool bundle.');
			setShowAlert(true);
		} finally {
			setBundleToDelete(null);
		}
	};

	const handleAddBundle = async (slug: string, display: string, description?: string) => {
		try {
			const id = getUUIDv7();
			await toolStoreAPI.putToolBundle(id, slug, display, true, description);
			setIsAddModalOpen(false);
			await fetchAll();
		} catch (err) {
			console.error('Add tool bundle failed:', err);
			setAlertMsg('Failed to add tool bundle.');
			setShowAlert(true);
		}
	};

	return (
		<div>
			{/* header */}
			<div className="fixed top-12 flex w-full justify-center">
				<div className="flex w-10/12 items-center justify-between p-0 lg:w-2/3">
					<h1 className="flex-grow text-center text-xl font-semibold">Tool Bundles</h1>
					<button
						className="btn btn-ghost flex items-center rounded-2xl"
						onClick={() => {
							setIsAddModalOpen(true);
						}}
					>
						<FiPlus /> <span className="ml-1">Add Bundle</span>
					</button>
				</div>
			</div>

			{/* body */}
			<div
				className="mt-12 flex w-full grow flex-col items-center overflow-y-auto"
				style={{ maxHeight: `calc(100vh - 144px)` }}
			>
				<div className="flex w-5/6 flex-col space-y-4 xl:w-2/3">
					{loading && <p className="mt-8 text-center text-sm">Loading bundles…</p>}
					{!loading && bundles.length === 0 && (
						<p className="mt-8 text-center text-sm">No tool bundles configured yet.</p>
					)}

					{bundles.map(bd => (
						<ToolBundleCard
							key={bd.bundle.id}
							bundle={bd.bundle}
							tools={bd.tools}
							onToolsChange={onToolsChange}
							onBundleDeleted={b => {
								setBundleToDelete(b);
							}}
						/>
					))}
				</div>
			</div>

			{/* delete bundle modal */}
			<DeleteConfirmationModal
				isOpen={bundleToDelete !== null}
				onClose={() => {
					setBundleToDelete(null);
				}}
				onConfirm={handleBundleDelete}
				title="Delete Tool Bundle"
				message={`Delete bundle "${bundleToDelete?.displayName ?? ''}" and all its tools?`}
				confirmButtonText="Delete"
			/>

			{/* add bundle modal */}
			<AddToolBundleModal
				isOpen={isAddModalOpen}
				onClose={() => {
					setIsAddModalOpen(false);
				}}
				onSubmit={handleAddBundle}
				existingSlugs={bundles.map(b => b.bundle.slug)}
			/>

			{/* alert */}
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

export default ToolsPage;
