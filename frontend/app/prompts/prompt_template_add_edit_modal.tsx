import { type ChangeEvent, type FormEvent, useEffect, useMemo, useRef, useState } from 'react';

import { createPortal } from 'react-dom';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import { PromptRoleEnum, type PromptTemplate } from '@/spec/prompt';

import { omitManyKeys } from '@/lib/obj_utils';
import { validateSlug, validateTags } from '@/lib/text_utils';
import { getUUIDv7 } from '@/lib/uuid_utils';

interface TemplateItem {
	template: PromptTemplate;
	bundleID: string;
	templateSlug: string;
}

interface AddEditPromptTemplateModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (templateData: Partial<PromptTemplate>) => void;
	initialData?: TemplateItem; // when editing
	existingTemplates: TemplateItem[];
}

type ErrorState = {
	displayName?: string;
	slug?: string;
	content?: string;
	tags?: string;
};

export function AddEditPromptTemplateModal({
	isOpen,
	onClose,
	onSubmit,
	initialData,
	existingTemplates,
}: AddEditPromptTemplateModalProps) {
	/* ---------- form state ---------- */
	const [formData, setFormData] = useState({
		displayName: '',
		slug: '',
		description: '',
		content: '',
		tags: '',
		isEnabled: true,
	});

	const [errors, setErrors] = useState<ErrorState>({});
	const isEditMode = Boolean(initialData);

	const dialogRef = useRef<HTMLDialogElement | null>(null);

	/* ---------- sync prop -> state ---------- */
	useEffect(() => {
		if (!isOpen) return;

		if (initialData) {
			setFormData({
				displayName: initialData.template.displayName,
				slug: initialData.template.slug,
				description: initialData.template.description ?? '',
				content: initialData.template.blocks[0]?.content ?? '',
				tags: (initialData.template.tags ?? []).join(', '),
				isEnabled: initialData.template.isEnabled,
			});
		} else {
			setFormData({
				displayName: '',
				slug: '',
				description: '',
				content: '',
				tags: '',
				isEnabled: true,
			});
		}
		setErrors({});
	}, [isOpen, initialData]);

	// Open/close the native <dialog> when isOpen changes
	useEffect(() => {
		if (!isOpen) return;

		const dialog = dialogRef.current;
		if (!dialog) return;

		if (!dialog.open) {
			dialog.showModal();
		}

		return () => {
			if (dialog.open) {
				dialog.close();
			}
		};
	}, [isOpen]);

	// Sync parent when dialog closes
	const handleDialogClose = () => {
		onClose();
	};

	/* ---------- validation helpers ---------- */
	const validateField = (field: keyof ErrorState, val: string, currentErrors: ErrorState): ErrorState => {
		let newErrs: ErrorState = { ...currentErrors };
		const v = val.trim();

		// Required fields
		if (!v) {
			newErrs[field] = 'This field is required.';
			return newErrs;
		}

		if (field === 'slug') {
			const err = validateSlug(v);
			if (err) {
				newErrs.slug = err;
			} else {
				const clash = existingTemplates.some(t => t.template.slug === v && t.template.id !== initialData?.template.id);
				if (clash) newErrs.slug = 'Slug already in use.';
				else newErrs = omitManyKeys(newErrs, ['slug']);
			}
		} else if (field === 'tags') {
			const err = validateTags(val);
			if (err) newErrs.tags = err;
			else newErrs = omitManyKeys(newErrs, ['tags']);
		} else {
			newErrs = omitManyKeys(newErrs, [field]);
		}

		return newErrs;
	};

	const validateForm = (state: typeof formData): ErrorState => {
		let newErrs: ErrorState = {};
		newErrs = validateField('displayName', state.displayName, newErrs);
		newErrs = validateField('slug', state.slug, newErrs);
		newErrs = validateField('content', state.content, newErrs);
		if (state.tags || state.tags.trim() !== '') {
			newErrs = validateField('tags', state.tags, newErrs);
		}
		return newErrs;
	};

	/* ---------- generic change handler ---------- */
	const handleInput = (e: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const { name, value, type, checked } = e.target as HTMLInputElement;
		const newVal = type === 'checkbox' ? checked : value;

		setFormData(prev => ({ ...prev, [name]: newVal }));

		if (name === 'displayName' || name === 'slug' || name === 'content' || name === 'tags') {
			setErrors(prev => validateField(name as keyof ErrorState, String(newVal), prev));
		}
	};

	/* ---------- overall validity ---------- */
	const isAllValid = useMemo(() => {
		const requiredFilled = formData.displayName.trim() && formData.slug.trim() && formData.content.trim();
		const hasErrors = Object.keys(errors).length > 0;
		return Boolean(requiredFilled) && !hasErrors;
	}, [formData, errors]);

	/* ---------- submit ---------- */
	const handleSubmit = (e: FormEvent) => {
		e.preventDefault();
		e.stopPropagation();

		const trimmed = {
			displayName: formData.displayName.trim(),
			slug: formData.slug.trim(),
			description: formData.description.trim(),
			content: formData.content,
			tags: formData.tags,
			isEnabled: formData.isEnabled,
		};

		const newErrs = validateForm(trimmed);
		setErrors(newErrs);

		if (Object.keys(newErrs).length > 0) return;

		const tagsArr = trimmed.tags
			.split(',')
			.map(t => t.trim())
			.filter(Boolean);

		onSubmit({
			displayName: trimmed.displayName,
			slug: trimmed.slug,
			description: trimmed.description || undefined,
			isEnabled: trimmed.isEnabled,
			tags: tagsArr.length ? tagsArr : undefined,
			blocks: [
				{
					id: initialData?.template.blocks[0]?.id ?? getUUIDv7(),
					role: PromptRoleEnum.User,
					content: trimmed.content,
				},
			],
		});

		dialogRef.current?.close();
	};

	/* ---------- early-return ---------- */
	if (!isOpen) return null;

	/* ---------- render ---------- */
	return createPortal(
		<dialog ref={dialogRef} className="modal" onClose={handleDialogClose}>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-3xl overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					{/* header */}
					<div className="mb-4 flex items-center justify-between">
						<h3 className="text-lg font-bold">{isEditMode ? 'Edit Prompt Template' : 'Add Prompt Template'}</h3>
						<button
							type="button"
							className="btn btn-sm btn-circle bg-base-300"
							onClick={() => dialogRef.current?.close()}
							aria-label="Close"
						>
							<FiX size={12} />
						</button>
					</div>

					{/* form */}
					<form noValidate onSubmit={handleSubmit} className="space-y-4">
						{/* Display Name */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Display Name*</span>
							</label>
							<div className="col-span-9">
								<input
									type="text"
									name="displayName"
									value={formData.displayName}
									onChange={handleInput}
									className={`input input-bordered w-full rounded-xl ${errors.displayName ? 'input-error' : ''}`}
									spellCheck="false"
									autoComplete="off"
									autoFocus
									aria-invalid={Boolean(errors.displayName)}
								/>
								{errors.displayName && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.displayName}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* Slug */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Slug*</span>
								<span className="label-text-alt tooltip tooltip-right" data-tip={`Short user friendly command`}>
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<div className="relative">
									<input
										type="text"
										name="slug"
										value={formData.slug}
										onChange={handleInput}
										className={`input input-bordered w-full rounded-xl pl-8 ${errors.slug ? 'input-error' : ''}`}
										spellCheck="false"
										autoComplete="off"
										disabled={isEditMode && initialData?.template.isBuiltIn}
										aria-invalid={Boolean(errors.slug)}
									/>
								</div>
								{errors.slug && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.slug}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* Enabled toggle */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3 cursor-pointer">
								<span className="label-text text-sm">Enabled</span>
							</label>
							<div className="col-span-9">
								<input
									type="checkbox"
									name="isEnabled"
									checked={formData.isEnabled}
									onChange={handleInput}
									className="toggle toggle-accent"
								/>
							</div>
						</div>

						{/* Description */}
						<div className="grid grid-cols-12 items-start gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Description</span>
							</label>
							<div className="col-span-9">
								<textarea
									name="description"
									value={formData.description}
									onChange={handleInput}
									className="textarea textarea-bordered h-20 w-full rounded-xl"
									spellCheck="false"
								/>
							</div>
						</div>

						{/* Content */}
						<div className="grid grid-cols-12 items-start gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Prompt Content*</span>
								<span
									className="label-text-alt tooltip tooltip-right"
									data-tip="First (user) message - add advanced messages later"
								>
									<FiHelpCircle size={12} />
								</span>
							</label>
							<div className="col-span-9">
								<textarea
									name="content"
									value={formData.content}
									onChange={handleInput}
									className={`textarea textarea-bordered h-32 w-full rounded-xl ${
										errors.content ? 'textarea-error' : ''
									}`}
									spellCheck="false"
									aria-invalid={Boolean(errors.content)}
								/>
								{errors.content && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.content}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* Tags */}
						<div className="grid grid-cols-12 items-center gap-2">
							<label className="label col-span-3">
								<span className="label-text text-sm">Tags</span>
							</label>
							<div className="col-span-9">
								<input
									type="text"
									name="tags"
									value={formData.tags}
									onChange={handleInput}
									className={`input input-bordered w-full rounded-xl ${errors.tags ? 'input-error' : ''}`}
									placeholder="comma, separated, tags"
									spellCheck="false"
									aria-invalid={Boolean(errors.tags)}
								/>
								{errors.tags && (
									<div className="label">
										<span className="label-text-alt text-error flex items-center gap-1">
											<FiAlertCircle size={12} /> {errors.tags}
										</span>
									</div>
								)}
							</div>
						</div>

						{/* actions */}
						<div className="modal-action">
							<button type="button" className="btn bg-base-300 rounded-xl" onClick={() => dialogRef.current?.close()}>
								Cancel
							</button>
							<button type="submit" className="btn btn-primary rounded-xl" disabled={!isAllValid}>
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
