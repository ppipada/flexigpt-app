import type { FC } from 'react';
import { useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import type { ProviderName, ProviderPreset } from '@/spec/modelpreset';
import type { AuthKeyMeta } from '@/spec/setting';
import { AuthKeyTypeProvider } from '@/spec/setting';

import { omitManyKeys } from '@/lib/obj_utils';

import { settingstoreAPI } from '@/apis/baseapi';
import { getAllProviderPresetsMap } from '@/apis/modelpreset_helper';

import type { DropdownItem } from '@/components/dropdown';
import Dropdown from '@/components/dropdown';

/* ────────────────────────── props & helpers ────────────────────────── */
interface Props {
	isOpen: boolean;
	initial: AuthKeyMeta | null; // null ⇒ create mode
	existing: AuthKeyMeta[]; // existing list (for duplicates / filtering)
	onClose: () => void;
	onChanged: () => void; // parent should refetch on success
}
const sentinelAddNew = '__add_new__';

interface FormData {
	type: string;
	keyName: string;
	secret: string;
	newType: string; // only when sentinel chosen
}

const AddEditAuthKeyModal: FC<Props> = ({ isOpen, initial, existing, onClose, onChanged }) => {
	const isEdit = Boolean(initial);

	/* ───────────────────────────── state ───────────────────────────── */
	const [form, setForm] = useState<FormData>({
		type: initial?.type ?? AuthKeyTypeProvider,
		keyName: initial?.keyName ?? '',
		secret: '',
		newType: '',
	});
	const [errors, setErrors] = useState<Partial<Record<keyof FormData, string>>>({});

	/* raw provider presets fetched from backend */
	const [providerPresets, setProviderPresets] = useState<Record<ProviderName, ProviderPreset>>({});

	/* ───────────────────────────── derived ───────────────────────────── */

	/* list of *types* that already exist (for dropdown) */
	const existingTypes = useMemo(() => [...new Set(existing.map(k => k.type))], [existing]);

	/* provider names that already HAVE an auth-key */
	const usedProviderNames = useMemo(
		() => new Set(existing.filter(k => k.type === AuthKeyTypeProvider).map(k => k.keyName)),
		[existing]
	);

	/* provider presets that are still AVAILABLE for “create” -------------
	   (i.e. have no auth-key yet)                                         */
	const availableProviderPresets = useMemo(() => {
		/* in edit mode we must still allow the current provider name */
		const allowed = new Set<ProviderName>();
		if (isEdit && initial?.type === AuthKeyTypeProvider) allowed.add(initial.keyName);

		const out: Record<ProviderName, ProviderPreset> = {};
		Object.entries(providerPresets).forEach(([name, preset]) => {
			if (!usedProviderNames.has(name) || allowed.has(name)) {
				out[name] = preset;
			}
		});
		return out;
	}, [providerPresets, usedProviderNames, isEdit, initial]);

	/* dropdown items for provider-name selection (create-mode only) */
	const providerDropdownItems = useMemo(() => {
		const obj: Record<ProviderName, DropdownItem> = {};
		Object.keys(availableProviderPresets).forEach(
			name =>
				(obj[name] = {
					isEnabled: true,
				})
		);
		return obj;
	}, [availableProviderPresets]);

	/* whether *no* provider is available to create a new key for */
	const noProviderAvailable =
		!isEdit && form.type === AuthKeyTypeProvider && Object.keys(providerDropdownItems).length === 0;

	/* type dropdown items (include existing + sentinel) */
	const typeDropdownItems = useMemo(() => {
		const obj: Record<string, DropdownItem> = {};
		existingTypes.forEach(t => (obj[t] = { isEnabled: true }));
		obj[sentinelAddNew] = { isEnabled: true };
		return obj;
	}, [existingTypes]);

	/* ───────────────────────────── effects ───────────────────────────── */

	/* reset the form when (re)opened */
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

	/* fetch provider presets once needed */
	useEffect(() => {
		if (isOpen && form.type === AuthKeyTypeProvider && Object.keys(providerPresets).length === 0) {
			(async () => {
				try {
					const prov = await getAllProviderPresetsMap(true);
					setProviderPresets(prov);
				} catch (err) {
					console.error('Failed fetching provider presets', err);
				}
			})();
		}
	}, [isOpen, form.type, providerPresets]);

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

			case 'keyName': {
				if (!value.trim()) next.keyName = 'Key name is required';
				else {
					const t = form.type === sentinelAddNew ? form.newType : form.type;
					if (checkDuplicate(t, value.trim())) next.keyName = 'Duplicate (type, key) pair';
				}
				break;
			}

			case 'secret':
				if (!value.trim()) next.secret = 'Secret cannot be empty';
				break;
		}
		setErrors(next);
	};

	/* ─────────── handlers ─────────── */
	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
		const { name, value } = e.target;
		setForm(prev => ({ ...prev, [name]: value }));
		validate(name as keyof FormData, value);
	};

	const handleTypeSelect = (k: string) => {
		setForm(prev => ({
			...prev,
			type: k,
			newType: '',
			/* reset key-name when switching away from provider */
			keyName: k === AuthKeyTypeProvider ? prev.keyName : '',
		}));
		validate('type', k);
	};

	const handleKeyNameSelect = (k: ProviderName) => {
		setForm(prev => ({ ...prev, keyName: k }));
		validate('keyName', k);
	};

	/* overall validity */
	const isAllValid = useMemo(() => {
		if (noProviderAvailable) return false;

		if (!form.secret.trim()) return false;
		if (form.type === sentinelAddNew && !form.newType.trim()) return false;
		if (!form.keyName.trim()) return false;
		return Object.values(errors).every(v => !v);
	}, [form, errors, noProviderAvailable]);

	/* ───────────────────────── submit ───────────────────────── */
	const submit = async (e: React.FormEvent) => {
		e.preventDefault();

		/* run validation on every field explicitly */
		(Object.entries(form) as [keyof FormData, string][]).forEach(([k, v]) => {
			validate(k, v);
		});
		if (!isAllValid) return;

		const finalType = form.type === sentinelAddNew ? form.newType.trim() : form.type;

		await settingstoreAPI.setAuthKey(finalType, form.keyName.trim(), form.secret.trim());

		onChanged();
		onClose();
	};

	/* closed → nothing to render */
	if (!isOpen) return null;

	/* ─────────────────────────── render ─────────────────────────── */
	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				{/* header */}
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">{isEdit ? 'Edit Auth Key' : 'Add Auth Key'}</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose}>
						<FiX size={12} />
					</button>
				</div>

				<form onSubmit={submit} className="space-y-4">
					{/* TYPE ----------------------------------------------------------------- */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Type*</span>
							<span className="label-text-alt tooltip" data-tip="Logical grouping of keys">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							{isEdit ? (
								<input className="input input-bordered w-full rounded-2xl" value={form.type} disabled />
							) : (
								<Dropdown<string>
									dropdownItems={typeDropdownItems}
									selectedKey={form.type}
									onChange={handleTypeSelect}
									filterDisabled={false}
									title="Select type"
									getDisplayName={k => (k === sentinelAddNew ? 'Add new type…' : k)}
								/>
							)}
							{errors.type && <FieldError msg={errors.type} />}
						</div>
					</div>

					{/* NEW TYPE (only when sentinel) -------------------------------------- */}
					{!isEdit && form.type === sentinelAddNew && (
						<div className="grid grid-cols-12 gap-2 items-center">
							<label className="label col-span-3">
								<span className="label-text text-sm">New Type*</span>
							</label>
							<div className="col-span-9">
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

					{/* KEY NAME ------------------------------------------------------------- */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Key Name*</span>
						</label>
						<div className="col-span-9">
							{/* provider-type dropdown (create) */}
							{!isEdit && form.type === AuthKeyTypeProvider ? (
								Object.keys(providerDropdownItems).length > 0 ? (
									<Dropdown<ProviderName>
										dropdownItems={providerDropdownItems}
										selectedKey={form.keyName}
										onChange={handleKeyNameSelect}
										filterDisabled={false}
										title="Select provider"
										getDisplayName={k => availableProviderPresets[k].displayName || k}
									/>
								) : (
									/* no provider left ─ show static disabled input */
									<input
										className="input input-bordered w-full rounded-2xl"
										value="All providers already configured"
										disabled
									/>
								)
							) : (
								/* simple text input (non-provider or edit-mode) */
								<input
									type="text"
									name="keyName"
									value={form.keyName}
									onChange={handleChange}
									className={`input input-bordered w-full rounded-2xl ${errors.keyName ? 'input-error' : ''}`}
									disabled={isEdit && form.type === AuthKeyTypeProvider}
									spellCheck="false"
								/>
							)}
							{errors.keyName && <FieldError msg={errors.keyName} />}
						</div>
					</div>

					{/* SECRET --------------------------------------------------------------- */}
					<div className="grid grid-cols-12 gap-2 items-center">
						<label className="label col-span-3">
							<span className="label-text text-sm">Secret*</span>
						</label>
						<div className="col-span-9">
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

					{/* ACTIONS -------------------------------------------------------------- */}
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

/* ───────────────────────── field-error helper ───────────────────────── */
const FieldError: FC<{ msg?: string }> = ({ msg }) =>
	msg ? (
		<div className="label">
			<span className="label-text-alt text-error flex items-center gap-1">
				<FiAlertCircle size={12} /> {msg}
			</span>
		</div>
	) : null;

export default AddEditAuthKeyModal;
