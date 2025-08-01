import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import type { AuthKeyMeta } from '@/spec/setting';
import { AuthKeyTypeProvider } from '@/spec/setting';

import { omitManyKeys } from '@/lib/obj_utils';

import { settingstoreAPI } from '@/apis/baseapi';

interface Props {
	isOpen: boolean;
	initial: AuthKeyMeta | null; // null = create
	existing: AuthKeyMeta[]; // list for uniqueness check
	onClose: () => void;
	onChanged: () => void; // parent should refetch on success
}

const sentinelAddNew = '__add_new__';

interface FormData {
	type: string;
	keyName: string;
	secret: string;
	newType: string; // only used when sentinel chosen
}

const AddEditAuthKeyModal: FC<Props> = ({ isOpen, initial, existing, onClose, onChanged }) => {
	const isEdit = Boolean(initial);

	const [form, setForm] = useState<FormData>({
		type: initial?.type ?? AuthKeyTypeProvider,
		keyName: initial?.keyName ?? '',
		secret: '',
		newType: '',
	});
	const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});

	const existingTypes = useMemo(() => [...new Set(existing.map(k => k.type))], [existing]);

	/* reset on open/close change */
	useEffect(() => {
		if (!isOpen) {
			setForm({
				type: initial?.type ?? AuthKeyTypeProvider,
				keyName: initial?.keyName ?? '',
				secret: '',
				newType: '',
			});
			setErrors({});
		}
	}, [isOpen, initial]);

	/* validations */
	const validate = (field: keyof FormData, value: string) => {
		const next = omitManyKeys(errors, [field]) as Partial<Record<keyof FormData, string>>;

		const checkDuplicate = (t: string, n: string) =>
			existing.some(k => k.type === t && k.keyName === n && !(isEdit && k === initial));

		switch (field) {
			case 'type':
				if (!value) next.type = 'Select a type';
				break;
			case 'newType':
				if (form.type === sentinelAddNew && !value.trim()) next.newType = 'New type required';
				break;
			case 'keyName':
				if (!value.trim()) next.keyName = 'Key name is required';
				else {
					const t = form.type === sentinelAddNew ? form.newType : form.type;
					if (checkDuplicate(t, value.trim())) next.keyName = 'Duplicate (type, key) pair';
				}
				break;
			case 'secret':
				if (!value.trim()) next.secret = 'Secret cannot be empty';
				break;
		}

		setErrors(next);
	};

	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
		const { name, value } = e.target;
		setForm(prev => ({ ...prev, [name]: value }));
		validate(name as keyof FormData, value);
	};

	const isAllValid = useMemo(() => {
		if (!form.secret.trim()) return false;
		if (form.type === sentinelAddNew && !form.newType.trim()) return false;
		if (!form.keyName.trim()) return false;
		return Object.values(errors).every(v => !v);
	}, [form, errors]);

	const submit = async (e: React.FormEvent) => {
		e.preventDefault();
		Object.entries(form).forEach(([k, v]) => {
			validate(k as keyof FormData, v);
		});
		if (!isAllValid) return;

		const finalType = form.type === sentinelAddNew ? form.newType.trim() : form.type;

		await settingstoreAPI.setAuthKey(finalType, form.keyName.trim(), form.secret.trim());
		onChanged();
		onClose();
	};

	if (!isOpen) return null;

	/* ui */
	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-2xl rounded-2xl">
				{/* title */}
				<div className="flex justify-between items-center mb-2">
					<h3 className="font-bold text-lg">{isEdit ? 'Edit Auth Key' : 'Add Auth Key'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose}>
						<FiX size={12} />
					</button>
				</div>

				<form onSubmit={submit} className="space-y-4">
					{/* ----- type ----- */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-4">
							<span className="label-text text-sm">Type*</span>
							<span className="label-text-alt tooltip" data-tip="Logical grouping of keys">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-8">
							{isEdit ? (
								<input className="input input-bordered w-full rounded-2xl" value={form.type} disabled />
							) : (
								<select
									name="type"
									className="select select-bordered w-full rounded-2xl"
									value={form.type}
									onChange={handleChange}
								>
									{existingTypes.map(t => (
										<option key={t} value={t}>
											{t}
										</option>
									))}
									<option value={sentinelAddNew}>Add new type…</option>
								</select>
							)}
							{errors.type && <FieldError msg={errors.type} />}
						</div>
					</div>

					{/* new type input */}
					{!isEdit && form.type === sentinelAddNew && (
						<div className="grid grid-cols-12 gap-2 items-center">
							<label className="label col-span-4">
								<span className="label-text text-sm">New Type*</span>
							</label>
							<div className="col-span-8">
								<input
									type="text"
									name="newType"
									value={form.newType}
									onChange={handleChange}
									className={`input input-bordered w-full rounded-2xl ${errors.newType ? 'input-error' : ''}`}
									spellCheck="false"
								/>
								{errors.newType && <FieldError msg={errors.newType} />}
							</div>
						</div>
					)}

					{/* key name */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-4">
							<span className="label-text text-sm">Key Name*</span>
						</label>
						<div className="col-span-8">
							<input
								type="text"
								name="keyName"
								value={form.keyName}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-2xl ${errors.keyName ? 'input-error' : ''}`}
								disabled={isEdit}
								spellCheck="false"
							/>
							{errors.keyName && <FieldError msg={errors.keyName} />}
						</div>
					</div>

					{/* secret */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-4">
							<span className="label-text text-sm">Secret*</span>
						</label>
						<div className="col-span-8">
							<input
								type="password"
								name="secret"
								value={form.secret}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-2xl ${errors.secret ? 'input-error' : ''}`}
								spellCheck="false"
							/>
							{errors.secret && <FieldError msg={errors.secret} />}
						</div>
					</div>

					{/* actions */}
					<div className="modal-action mt-6">
						<button type="button" className="btn rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" disabled={!isAllValid} className="btn btn-primary rounded-2xl">
							{isEdit ? 'Update' : 'Add'}
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

/* tiny helper */
const FieldError: FC<{ msg?: string }> = ({ msg }) =>
	msg ? (
		<div className="label">
			<span className="label-text-alt text-error flex items-center gap-1">
				<FiAlertCircle size={12} /> {msg}
			</span>
		</div>
	) : null;

export default AddEditAuthKeyModal;
