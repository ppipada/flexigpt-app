import type { FC } from 'react';
import { useEffect, useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import { type AuthKeyMeta } from '@/spec/setting';

import { settingstoreAPI } from '@/apis/baseapi';

import DownloadButton from '@/components/download_button';
import { ThemeSelector } from '@/components/theme';

import AddEditAuthKeyModal from '@/settings/authkey_modal';
import AuthKeyTable from '@/settings/authkey_table';

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
		<div className="flex flex-col items-center w-full h-full">
			{/* sticky header */}
			<header className="w-full flex justify-center fixed top-8 z-20">
				<div className="w-10/12 lg:w-2/3 flex items-center justify-between p-2">
					<h1 className="text-xl font-semibold">Settings</h1>
					<DownloadButton
						title="Download Settings"
						language="json"
						valueFetcher={exportSettings}
						size={24}
						fileprefix="settings"
						className="btn btn-sm btn-ghost"
						isBinary={false}
					/>
				</div>
			</header>

			<main
				className="flex flex-col items-center w-full grow mt-24 overflow-y-auto"
				style={{ maxHeight: 'calc(100vh - 128px)' }}
			>
				<div className="flex flex-col gap-8 w-5/6 xl:w-2/3">
					{/* ── Theme selector ──────────────────────────── */}
					<section className="flex items-center bg-base-100 rounded-2xl shadow-lg p-4">
						<h2 className="font-semibold mr-8">Theme</h2>
						<ThemeSelector />
					</section>

					{/* ── Auth-Key table ─────────────────────────── */}
					<section className="bg-base-100 rounded-2xl shadow-lg p-4">
						<div className="flex justify-between items-center mb-4">
							<h2 className="font-semibold">Auth Keys</h2>
							<button className="btn btn-ghost rounded-2xl flex items-center" onClick={showAddModal}>
								<FiPlus className="mr-1" /> Add Key
							</button>
						</div>

						<AuthKeyTable authKeys={authKeys} onEdit={showEditModal} onChanged={refresh} />
					</section>
				</div>
			</main>

			{/* modal */}
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
	);
};

export default SettingsPage;
