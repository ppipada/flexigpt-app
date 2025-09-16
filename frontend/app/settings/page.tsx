import type { FC } from 'react';
import { useEffect, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import { type AuthKeyMeta } from '@/spec/setting';

import { settingstoreAPI } from '@/apis/baseapi';

import DownloadButton from '@/components/download_button';
import PageFrame from '@/components/page_frame';

import AddEditAuthKeyModal from '@/settings/authkey_add_edit';
import AuthKeyTable from '@/settings/authkey_table';
import { ThemeSelector } from '@/settings/theme';

const SettingsPage: FC = () => {
	const [authKeys, setAuthKeys] = useState<AuthKeyMeta[]>([]);
	const [refreshToggle, setRefresh] = useState(false); // helper to force list refresh

	const [isModalOpen, setIsModalOpen] = useState(false);
	const [modalInitial, setModalInitial] = useState<AuthKeyMeta | null>(null); // null = Add

	useEffect(() => {
		(async () => {
			const settings = await settingstoreAPI.getSettings();
			setAuthKeys(settings.authKeys);
		})();
	}, [refreshToggle]);

	const refresh = () => {
		setRefresh(p => !p);
	};
	const showAddModal = () => {
		setModalInitial(null);
		setIsModalOpen(true);
	};
	const showEditModal = (meta: AuthKeyMeta) => {
		setModalInitial(meta);
		setIsModalOpen(true);
	};

	const exportSettings = async () => {
		const settings = await settingstoreAPI.getSettings(true);
		return JSON.stringify(settings, null, 2);
	};

	return (
		<PageFrame>
			<div className="flex h-full w-full flex-col items-center">
				{/* sticky header */}
				<header className="fixed mt-8 flex w-10/12 items-center p-2 lg:w-2/3">
					<h1 className="flex grow items-center justify-center text-xl font-semibold">Settings</h1>
					<DownloadButton
						title="Download Settings"
						language="json"
						valueFetcher={exportSettings}
						size={20}
						fileprefix="settings"
						className="btn btn-sm btn-ghost"
						isBinary={false}
					/>
				</header>

				<main className="mt-24 flex w-full grow flex-col items-center overflow-y-auto">
					<div className="flex w-5/6 flex-col gap-8 xl:w-2/3">
						{/* ── Theme selector ──────────────────────────── */}
						<section className="bg-base-100 flex items-center rounded-2xl p-4 shadow-lg">
							<h2 className="mr-8 ml-4 font-semibold">Theme</h2>
							<ThemeSelector />
						</section>

						{/* ── Auth-Key table ─────────────────────────── */}
						<section className="bg-base-100 rounded-2xl p-4 shadow-lg">
							<div className="mb-4 flex items-center justify-between">
								<h2 className="mr-8 ml-4 font-semibold">Auth Keys</h2>
								<button className="btn btn-ghost flex items-center rounded-2xl" onClick={showAddModal}>
									<FiPlus className="mr-1" /> Add Key
								</button>
							</div>

							<AuthKeyTable authKeys={authKeys} onEdit={showEditModal} onChanged={refresh} />
						</section>
					</div>
				</main>

				{isModalOpen && (
					<AddEditAuthKeyModal
						isOpen={isModalOpen}
						initial={modalInitial}
						existing={authKeys}
						onClose={() => {
							setIsModalOpen(false);
						}}
						onChanged={refresh}
					/>
				)}
			</div>
		</PageFrame>
	);
};

export default SettingsPage;
