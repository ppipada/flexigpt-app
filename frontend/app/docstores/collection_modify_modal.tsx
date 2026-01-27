import { type ChangeEvent, type FormEvent, useEffect, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

import { FiAlertCircle, FiX } from 'react-icons/fi';

import type { Collection } from '@/spec/docstore';

import { omitManyKeys } from '@/lib/obj_utils';

interface ModifyCollectionModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (collectionData: Partial<Collection>, docStoreID: string) => void;
	initialData?: Partial<Collection>;
	docStoreID: string;
	existingCollections: Collection[];
}

type ErrorState = {
	name?: string;
	command?: string;
};

export function ModifyCollectionModal({
	isOpen,
	onClose,
	onSubmit,
	initialData,
	docStoreID,
	existingCollections,
}: ModifyCollectionModalProps) {
	const [formData, setFormData] = useState<Partial<Collection>>({
		name: '',
		command: '',
	});
	const [errors, setErrors] = useState<ErrorState>({});

	const dialogRef = useRef<HTMLDialogElement | null>(null);

	useEffect(() => {
		if (!isOpen) return;
		setFormData(initialData || { name: '', command: '' });
		setErrors({});
	}, [isOpen, initialData]);

	// Open/close native dialog
	useEffect(() => {
		if (!isOpen) return;
		const dialog = dialogRef.current;
		if (!dialog) return;

		if (!dialog.open) dialog.showModal();

		return () => {
			if (dialog.open) dialog.close();
		};
	}, [isOpen]);

	const handleDialogClose = () => {
		onClose();
	};

	const validateField = (field: keyof ErrorState, value: string, currentErrors: ErrorState): ErrorState => {
		let nextErrors: ErrorState = { ...currentErrors };
		const v = value.trim();

		if (!v) {
			nextErrors[field] = 'This field is required';
			return nextErrors;
		}

		const isUnique = !existingCollections.some(c => {
			if (field === 'name') return c.name === v && c.id !== initialData?.id;

			if (field === 'command') return c.command === v && c.id !== initialData?.id;
			return false;
		});

		if (!isUnique) {
			nextErrors[field] = `This ${field} is already in use`;
		} else {
			nextErrors = omitManyKeys(nextErrors, [field]);
		}

		return nextErrors;
	};

	const validateForm = (state: Partial<Collection>): ErrorState => {
		let nextErrors: ErrorState = {};
		nextErrors = validateField('name', state.name || '', nextErrors);
		nextErrors = validateField('command', state.command || '', nextErrors);
		return nextErrors;
	};

	const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
		const { name, value } = e.target;
		const updated = { ...formData, [name]: value };
		setFormData(updated);
		setErrors(prev => validateField(name as keyof ErrorState, value, prev));
	};

	const handleSubmit = (e: FormEvent) => {
		e.preventDefault();
		e.stopPropagation();

		const nextErrors = validateForm(formData);
		setErrors(nextErrors);

		if (Object.keys(nextErrors).length > 0) return;
		if (!formData.name || !formData.command) return;

		onSubmit(
			{
				...formData,
				name: formData.name.trim(),
				command: formData.command.trim(),
			},
			docStoreID
		);

		dialogRef.current?.close();
	};

	const isFormValid =
		Boolean(formData.name?.trim()) && Boolean(formData.command?.trim()) && Object.keys(errors).length === 0;

	if (!isOpen) return null;

	return createPortal(
		<dialog
			ref={dialogRef}
			className="modal"
			onClose={handleDialogClose}
			onCancel={e => {
				// Form mode: do NOT allow Esc to close.
				e.preventDefault();
			}}
		>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-3xl overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					{/* Header */}
					<div className="mb-4 flex items-center justify-between">
						<h3 className="text-lg font-bold">{initialData ? 'Edit Collection' : 'Add New Collection'}</h3>
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
								type="text"
								name="name"
								value={formData.name || ''}
								onChange={handleChange}
								className={`input input-bordered w-full rounded-xl ${errors.name ? 'input-error' : ''}`}
								spellCheck="false"
								autoComplete="off"
								autoFocus
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
							<label className="label" htmlFor="command">
								<span className="label-text text-sm">Command*</span>
							</label>
							<div className="relative">
								<input
									type="text"
									name="command"
									value={formData.command || ''}
									onChange={handleChange}
									className={`input input-bordered w-full rounded-xl pl-8 ${errors.command ? 'input-error' : ''}`}
									spellCheck="false"
									autoComplete="off"
									aria-invalid={Boolean(errors.command)}
								/>
							</div>
							{errors.command && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.command}
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
