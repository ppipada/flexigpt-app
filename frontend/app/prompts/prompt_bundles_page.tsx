/* pages/prompt_bundles_page.tsx */
import { type FC, useEffect, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import type { PromptBundle, PromptTemplate } from '@/spec/prompt';

import { getUUIDv7 } from '@/lib/uuid_utils';

import { promptStoreAPI } from '@/apis/baseapi';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';
import Loader from '@/components/loader';

import AddBundleModal from '@/prompts/prompt_bundle_add';
import PromptBundleCard from '@/prompts/prompt_bundle_card';

interface BundleData {
	bundle: PromptBundle;
	templates: PromptTemplate[];
}

const PromptBundlesPage: FC = () => {
	const [bundles, setBundles] = useState<BundleData[]>([]);
	const [loading, setLoading] = useState(true);

	/* alerts */
	const [showAlert, setShowAlert] = useState(false);
	const [alertMsg, setAlertMsg] = useState('');

	/* bundle deletion modal */
	const [bundleToDelete, setBundleToDelete] = useState<PromptBundle | null>(null);

	/* add-bundle modal */
	const [isAddModalOpen, setIsAddModalOpen] = useState(false);

	const fetchAll = async () => {
		setLoading(true);
		try {
			const { promptBundles } = await promptStoreAPI.listPromptBundles(undefined, true);

			const bundleResults: BundleData[] = await Promise.all(
				promptBundles.map(async b => {
					try {
						const { promptTemplateListItems } = await promptStoreAPI.listPromptTemplates([b.id], undefined, true);

						const tplPromises = promptTemplateListItems.map(itm =>
							promptStoreAPI.getPromptTemplate(itm.bundleID, itm.templateSlug, itm.templateVersion)
						);
						const tpl = (await Promise.all(tplPromises)).filter((t): t is PromptTemplate => t !== undefined);
						return { bundle: b, templates: tpl };
					} catch {
						return { bundle: b, templates: [] };
					}
				})
			);

			setBundles(bundleResults);
		} catch (err) {
			console.error('Failed to load bundles:', err);
			setAlertMsg('Failed to load bundles. Please try again.');
			setShowAlert(true);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		fetchAll();
	}, []);

	const onTemplatesChange = (bundleID: string, newTpl: PromptTemplate[]) => {
		setBundles(prev => prev.map(bd => (bd.bundle.id === bundleID ? { ...bd, templates: newTpl } : bd)));
	};

	const handleBundleDelete = async () => {
		if (!bundleToDelete) return;
		try {
			await promptStoreAPI.deletePromptBundle(bundleToDelete.id);
			setBundles(prev => prev.filter(bd => bd.bundle.id !== bundleToDelete.id));
		} catch (err) {
			console.error('Delete bundle failed:', err);
			setAlertMsg('Failed to delete bundle.');
			setShowAlert(true);
		} finally {
			setBundleToDelete(null);
		}
	};

	const handleAddBundle = async (slug: string, display: string, description?: string) => {
		try {
			const id = getUUIDv7();
			await promptStoreAPI.putPromptBundle(id, slug, display, true, description);
			setIsAddModalOpen(false);
			await fetchAll();
		} catch (err) {
			console.error('Add bundle failed:', err);
			setAlertMsg('Failed to add bundle.');
			setShowAlert(true);
		}
	};

	// full-page loader while first fetch is running
	if (loading) {
		return <Loader text="Loading bundles…" />;
	}

	return (
		<div>
			{/* header */}
			<div className="w-full flex justify-center fixed top-12">
				<div className="w-10/12 lg:w-2/3 flex items-center justify-between p-0">
					<h1 className="text-xl font-semibold text-center flex-grow">Prompt Bundles</h1>

					<button
						className="btn btn-ghost rounded-2xl flex items-center"
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
				className="flex flex-col items-center w-full grow mt-12 overflow-y-auto"
				style={{ maxHeight: 'calc(100vh - 144px)' }}
			>
				<div className="flex flex-col space-y-4 w-5/6 xl:w-2/3">
					{bundles.length === 0 && <p className="text-center text-sm mt-8">No bundles configured yet.</p>}

					{bundles.map(bd => (
						<PromptBundleCard
							key={bd.bundle.id}
							bundle={bd.bundle}
							templates={bd.templates}
							onTemplatesChange={onTemplatesChange}
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
				title="Delete Prompt Bundle"
				message={`Delete bundle "${bundleToDelete?.displayName ?? ''}" and all its templates?`}
				confirmButtonText="Delete"
			/>

			{/* add bundle modal */}
			<AddBundleModal
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

export default PromptBundlesPage;
