import { type ChangeEvent, type FormEvent, useEffect, useMemo, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

import { FiAlertCircle, FiX } from 'react-icons/fi';

import type { DocStore } from '@/spec/docstore';

import { MessageEnterValidURL, validateUrlForInput } from '@/lib/url_utils';

interface ModifyDocStoreModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (storeData: Partial<DocStore>) => void;
	initialData?: Partial<DocStore>;
	existingDocStores: DocStore[];
}

type ErrorState = {
	name?: string;
	url?: string;
	dbName?: string;
};

export function ModifyDocStoreModal({
	isOpen,
	onClose,
	onSubmit,
	initialData,
	existingDocStores,
}: ModifyDocStoreModalProps) {
	const [formData, setFormData] = useState<Partial<DocStore>>({
		name: '',
		url: '',
		description: '',
		dbName: '',
	});

	const [errors, setErrors] = useState<ErrorState>({});

	const dialogRef = useRef<HTMLDialogElement | null>(null);
	const nameInputRef = useRef<HTMLInputElement | null>(null);
	const urlInputRef = useRef<HTMLInputElement | null>(null);
	const dbNameInputRef = useRef<HTMLInputElement | null>(null);

	useEffect(() => {
		if (!isOpen) return;

		setFormData(
			initialData || {
				name: '',
				url: '',
				description: '',
				dbName: '',
			}
		);
		setErrors({});
	}, [isOpen, initialData]);

	// Per-field validator (uses shared URL validator for url)
	const validateField = (field: keyof ErrorState, value: string, currentErrors: ErrorState): ErrorState => {
		const next: ErrorState = { ...currentErrors };
		const v = value.trim();

		if (field === 'name') {
			if (!v) next.name = 'This field is required';
			else delete next.name;
		}

		if (field === 'dbName') {
			if (!v) next.dbName = 'This field is required';
			else delete next.dbName;
		}

		if (field === 'url') {
			const { error } = validateUrlForInput(v, urlInputRef.current, {
				required: true,
				requiredMessage: 'This field is required',
			});

			if (error) next.url = error;
			else delete next.url;
		}

		return next;
	};

	// Full-form validation (adds uniqueness check on top of field-level checks)
	const validateAll = (state: Partial<DocStore>): ErrorState => {
		let next: ErrorState = {};

		const name = state.name ?? '';
		const url = state.url ?? '';
		const dbName = state.dbName ?? '';

		next = validateField('name', name, next);
		next = validateField('url', url, next);
		next = validateField('dbName', dbName, next);

		const trimmedUrl = url.trim();
		const trimmedDbName = dbName.trim();

		// Uniqueness only if url & dbName already pass their own validation
		if (!next.url && !next.dbName && trimmedUrl && trimmedDbName) {
			const isUnique = !existingDocStores.some(
				s => s.url === trimmedUrl && s.dbName === trimmedDbName && s.id !== initialData?.id
			);

			if (!isUnique) {
				const msg = 'This combination of URL and Database Name already exists';
				next.url = msg;
				next.dbName = msg;
			}
		}

		return next;
	};

	const handleChange = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const { name, value } = e.target;
		const updated = { ...formData, [name]: value };
		setFormData(updated);
		setErrors(validateAll(updated));
	};

	// Open/close native dialog and focus first field
	useEffect(() => {
		if (!isOpen) return;
		const dialog = dialogRef.current;
		if (!dialog) return;

		if (!dialog.open) dialog.showModal();

		// Focus "Name" field when dialog opens
		window.setTimeout(() => {
			nameInputRef.current?.focus();
		}, 0);

		return () => {
			if (dialog.open) dialog.close();
		};
	}, [isOpen]);

	const handleDialogClose = () => {
		onClose();
	};

	const handleSubmit = (e: FormEvent) => {
		e.preventDefault();
		e.stopPropagation();

		const finalErrors = validateAll(formData);
		setErrors(finalErrors);

		if (Object.keys(finalErrors).length > 0) return;

		// Normalize URL using shared validator (required)
		const urlInput = urlInputRef.current;
		const { normalized: normalizedUrl, error: urlError } = validateUrlForInput(formData.url ?? '', urlInput, {
			required: true,
			requiredMessage: 'This field is required',
		});

		if (!normalizedUrl || urlError) {
			setErrors(prev => ({
				...prev,
				url: urlError ?? MessageEnterValidURL,
			}));
			urlInput?.focus();
			return;
		}

		onSubmit({
			...formData,
			name: formData.name?.trim(),
			url: normalizedUrl,
			description: formData.description?.trim(),
			dbName: formData.dbName?.trim(),
		});

		dialogRef.current?.close();
	};

	const isFormValid = useMemo(
		() =>
			Boolean(formData.name?.trim()) &&
			Boolean(formData.url?.trim()) &&
			Boolean(formData.dbName?.trim()) &&
			Object.keys(errors).length === 0,
		[formData, errors]
	);

	if (!isOpen) return null;

	return createPortal(
		<dialog ref={dialogRef} className="modal" onClose={handleDialogClose}>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-3xl overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					<div className="mb-4 flex items-center justify-between">
						<h3 className="text-lg font-bold">{initialData ? 'Edit Document Store' : 'Add New Document Store'}</h3>
						<button
							type="button"
							className="btn btn-sm btn-circle bg-base-300"
							onClick={() => dialogRef.current?.close()}
							aria-label="Close"
						>
							<FiX size={12} />
						</button>
					</div>

					<form noValidate onSubmit={handleSubmit} className="space-y-4">
						<fieldset className="fieldset">
							<label className="label" htmlFor="name">
								<span className="label-text text-sm">Name*</span>
							</label>
							<input
								ref={nameInputRef}
								type="text"
								name="name"
								value={formData.name ?? ''}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-xl ${errors.name ? 'input-error' : ''}`}
								spellCheck="false"
								autoComplete="off"
								aria-invalid={Boolean(errors.name)}
							/>
							{errors.name && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.name}
									</span>
								</div>
							)}
						</fieldset>

						<fieldset className="fieldset">
							<label className="label" htmlFor="url">
								<span className="label-text text-sm">URL*</span>
							</label>
							<input
								ref={urlInputRef}
								type="url"
								name="url"
								value={formData.url ?? ''}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-xl ${errors.url ? 'input-error' : ''}`}
								spellCheck="false"
								autoComplete="off"
								aria-invalid={Boolean(errors.url)}
								placeholder="https://example.com OR example.com"
							/>
							{errors.url && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.url}
									</span>
								</div>
							)}
						</fieldset>

						<fieldset className="fieldset">
							<label className="label" htmlFor="description">
								<span className="label-text text-sm">Description</span>
							</label>
							<textarea
								name="description"
								value={formData.description ?? ''}
								onChange={handleChange}
								className="textarea textarea-bordered h-24 w-full rounded-xl"
								spellCheck="false"
							/>
						</fieldset>

						<fieldset className="fieldset">
							<label className="label" htmlFor="dbName">
								<span className="label-text text-sm">Database Name*</span>
							</label>
							<input
								ref={dbNameInputRef}
								type="text"
								name="dbName"
								value={formData.dbName ?? ''}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-xl ${errors.dbName ? 'input-error' : ''}`}
								spellCheck="false"
								autoComplete="off"
								aria-invalid={Boolean(errors.dbName)}
							/>
							{errors.dbName && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.dbName}
									</span>
								</div>
							)}
						</fieldset>

						<div className="modal-action">
							<button type="button" className="btn bg-base-300 rounded-xl" onClick={() => dialogRef.current?.close()}>
								Cancel
							</button>
							<button type="submit" className="btn btn-primary rounded-xl" disabled={!isFormValid}>
								Save
							</button>
						</div>
					</form>
				</div>
			</div>
			{/* NOTE: no modal-backdrop here: backdrop click should NOT close this modal */}
		</dialog>,
		document.body
	);
}
