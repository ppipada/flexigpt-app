import { useEffect, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import type { Tool, ToolBundle } from '@/spec/tool';

import { getUUIDv7 } from '@/lib/uuid_utils';

import { toolStoreAPI } from '@/apis/baseapi';
import { getAllToolBundles, getAllTools } from '@/apis/list_helper';

import { ActionDeniedAlertModal } from '@/components/action_denied_modal';
import { DeleteConfirmationModal } from '@/components/delete_confirmation_modal';
import { Loader } from '@/components/loader';
import { PageFrame } from '@/components/page_frame';

import { AddToolBundleModal } from '@/tools/tool_bundle_add_modal';
import { ToolBundleCard } from '@/tools/tool_bundle_card';

interface BundleData {
	bundle: ToolBundle;
	tools: Tool[];
}

// eslint-disable-next-line no-restricted-exports
export default function ToolsPage() {
	const [bundles, setBundles] = useState<BundleData[]>([]);
	const [loading, setLoading] = useState(true);

	const [showAlert, setShowAlert] = useState(false);
	const [alertMsg, setAlertMsg] = useState('');

	const [bundleToDelete, setBundleToDelete] = useState<ToolBundle | null>(null);
	const [isAddModalOpen, setIsAddModalOpen] = useState(false);

	const fetchAll = async () => {
		setLoading(true);
		try {
			const toolBundles = await getAllToolBundles(undefined, true);
			const bundleResults: BundleData[] = await Promise.all(
				toolBundles.map(async b => {
					try {
						const toolListItems = await getAllTools([b.id], undefined, true);
						const tools = toolListItems.map(itm => itm.toolDefinition);
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

	const onBundleEnableChange = (bundleID: string, enabled: boolean) => {
		setBundles(prev =>
			prev.map(bd => (bd.bundle.id === bundleID ? { ...bd, bundle: { ...bd.bundle, isEnabled: enabled } } : bd))
		);
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

	if (loading) {
		return <Loader text="Loading tool bundles…" />;
	}

	return (
		<PageFrame>
			<div className="flex h-full w-full flex-col items-center">
				<div className="fixed mt-8 flex w-10/12 items-center p-2 lg:w-2/3">
					<h1 className="flex grow items-center justify-center text-xl font-semibold">Tool Bundles</h1>
					<button
						className="btn btn-ghost flex items-center rounded-2xl"
						onClick={() => {
							setIsAddModalOpen(true);
						}}
					>
						<FiPlus size={20} /> <span className="ml-1">Add Bundle</span>
					</button>
				</div>

				<div
					className="mt-24 flex w-full grow flex-col items-center overflow-y-auto"
					style={{ maxHeight: `calc(100vh - 128px)` }}
				>
					<div className="flex w-5/6 flex-col space-y-4 xl:w-2/3">
						{bundles.length === 0 && <p className="mt-8 text-center text-sm">No tool bundles configured yet.</p>}

						{bundles.map(bd => (
							<ToolBundleCard
								key={bd.bundle.id}
								bundle={bd.bundle}
								tools={bd.tools}
								onToolsChange={onToolsChange}
								onBundleEnableChange={onBundleEnableChange}
								onBundleDeleted={b => {
									setBundleToDelete(b);
								}}
							/>
						))}
					</div>
				</div>

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

				<AddToolBundleModal
					isOpen={isAddModalOpen}
					onClose={() => {
						setIsAddModalOpen(false);
					}}
					onSubmit={handleAddBundle}
					existingSlugs={bundles.map(b => b.bundle.slug)}
				/>

				<ActionDeniedAlertModal
					isOpen={showAlert}
					onClose={() => {
						setShowAlert(false);
						setAlertMsg('');
					}}
					message={alertMsg}
				/>
			</div>
		</PageFrame>
	);
}
