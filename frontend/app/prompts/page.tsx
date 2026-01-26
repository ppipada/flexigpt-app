import { useEffect, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import type { PromptBundle, PromptTemplate } from '@/spec/prompt';

import { getUUIDv7 } from '@/lib/uuid_utils';

import { promptStoreAPI } from '@/apis/baseapi';
import { getAllPromptBundles, getAllPromptTemplates } from '@/apis/list_helper';

import { ActionDeniedAlertModal } from '@/components/action_denied_modal';
import { DeleteConfirmationModal } from '@/components/delete_confirmation_modal';
import { Loader } from '@/components/loader';
import { PageFrame } from '@/components/page_frame';

import { AddBundleModal } from '@/prompts/prompt_bundle_add_modal';
import { PromptBundleCard } from '@/prompts/prompt_bundle_card';

interface BundleData {
	bundle: PromptBundle;
	templates: PromptTemplate[];
}

// eslint-disable-next-line no-restricted-exports
export default function PromptsPage() {
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
			const promptBundles = await getAllPromptBundles(undefined, true);

			const bundleResults: BundleData[] = await Promise.all(
				promptBundles.map(async b => {
					try {
						const promptTemplateListItems = await getAllPromptTemplates([b.id], undefined, true);

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

	const onBundleEnableChange = (bundleID: string, enabled: boolean) => {
		setBundles(prev =>
			prev.map(bd => (bd.bundle.id === bundleID ? { ...bd, bundle: { ...bd.bundle, isEnabled: enabled } } : bd))
		);
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

	if (loading) {
		return <Loader text="Loading bundles…" />;
	}

	return (
		<PageFrame>
			<div className="flex h-full w-full flex-col items-center">
				<div className="fixed mt-8 flex w-10/12 items-center p-2 lg:w-2/3">
					<h1 className="flex grow items-center justify-center text-xl font-semibold">Prompt Bundles</h1>
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
						{bundles.length === 0 && <p className="mt-8 text-center text-sm">No bundles configured yet.</p>}

						{bundles.map(bd => (
							<PromptBundleCard
								key={bd.bundle.id}
								bundle={bd.bundle}
								templates={bd.templates}
								onTemplatesChange={onTemplatesChange}
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
					title="Delete Prompt Bundle"
					message={`Delete bundle "${bundleToDelete?.displayName ?? ''}" and all its templates?`}
					confirmButtonText="Delete"
				/>

				<AddBundleModal
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
