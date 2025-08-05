import { type FC, useState } from 'react';

import { FiCheckCircle, FiEdit2, FiRefreshCw, FiTrash2, FiXCircle } from 'react-icons/fi';

import type { AuthKeyMeta } from '@/spec/setting';

import { settingstoreAPI } from '@/apis/baseapi';
import { isBuiltInProviderAuthKeyName } from '@/apis/builtin_provider_cache';

import ActionDeniedAlert from '@/components/action_denied';
import DeleteConfirmationModal from '@/components/delete_confirmation';

/* table wrapper */
interface AuthKeyTableProps {
	authKeys: AuthKeyMeta[];
	onEdit: (meta: AuthKeyMeta) => void;
	onChanged: () => void; // parent refetch
}
const AuthKeyTable: FC<AuthKeyTableProps> = ({ authKeys, onEdit, onChanged }) => {
	if (!authKeys.length) return <p className="text-sm text-center text-neutral/70 my-6">No keys defined.</p>;

	return (
		<div className="overflow-x-auto rounded-2xl">
			<table className="table table-zebra w-full">
				<thead className="text-sm font-semibold bg-base-300">
					<tr className="text-sm">
						<th>Type</th>
						<th>Key Name</th>
						<th className="w-[11rem]">SHA-256</th>
						<th>Status</th>
						<th className="text-right">Actions</th>
					</tr>
				</thead>
				<tbody>
					{authKeys.map(k => (
						<AuthKeyRow
							key={`${k.type}:${k.keyName}`}
							meta={k}
							onEdit={() => {
								onEdit(k);
							}}
							onChanged={onChanged}
						/>
					))}
				</tbody>
			</table>
		</div>
	);
};

/* single row */
interface RowProps {
	meta: AuthKeyMeta;
	onEdit: () => void;
	onChanged: () => void;
}
const AuthKeyRow: FC<RowProps> = ({ meta, onEdit, onChanged }) => {
	const [showDelete, setShowDelete] = useState(false);
	const [denyAlert, setDenyAlert] = useState<{ visible: boolean; msg: string }>({ visible: false, msg: '' });

	const inbuilt = isBuiltInProviderAuthKeyName(meta.type, meta.keyName);

	const deleteKey = async () => {
		try {
			await settingstoreAPI.deleteAuthKey(meta.type, meta.keyName);
			onChanged();
		} finally {
			setShowDelete(false);
		}
	};

	const resetKey = async () => {
		await settingstoreAPI.setAuthKey(meta.type, meta.keyName, '');
		onChanged();
	};

	const handleDelete = () => {
		if (inbuilt) {
			setDenyAlert({ visible: true, msg: 'In-built keys cannot be deleted. You can only reset them.' });
		} else setShowDelete(true);
	};

	return (
		<>
			<tr className="hover:bg-base-300 border-none shadow-none">
				<td className="capitalize">{meta.type}</td>
				<td>{meta.keyName}</td>
				<td className="font-mono text-xs">{meta.nonEmpty ? meta.sha256.slice(0, 10) + '…' : '—'}</td>
				<td>
					{meta.nonEmpty ? (
						<FiCheckCircle className="text-success inline" />
					) : (
						<FiXCircle className="text-error inline" />
					)}
				</td>
				<td className="flex gap-3 justify-end">
					<button className="btn btn-xs btn-ghost" onClick={onEdit} title="Edit">
						<FiEdit2 size={14} />
					</button>
					{inbuilt ? (
						<button className="btn btn-xs btn-ghost" onClick={resetKey} title="Reset">
							<FiRefreshCw size={14} />
						</button>
					) : (
						<button className="btn btn-xs btn-ghost" onClick={handleDelete} title="Delete">
							<FiTrash2 size={14} />
						</button>
					)}
				</td>
			</tr>

			{showDelete && (
				<DeleteConfirmationModal
					isOpen={showDelete}
					title="Delete Auth Key"
					message={`Delete key "${meta.keyName}" of type "${meta.type}"?  This cannot be undone.`}
					confirmButtonText="Delete"
					onConfirm={deleteKey}
					onClose={() => {
						setShowDelete(false);
					}}
				/>
			)}

			{denyAlert.visible && (
				<ActionDeniedAlert
					isOpen={denyAlert.visible}
					message={denyAlert.msg}
					onClose={() => {
						setDenyAlert({ visible: false, msg: '' });
					}}
				/>
			)}
		</>
	);
};

export default AuthKeyTable;
