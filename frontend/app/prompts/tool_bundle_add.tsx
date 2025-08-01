import { useEffect, useMemo, useState } from 'react';

import { FiAlertCircle, FiHelpCircle, FiX } from 'react-icons/fi';

import { omitManyKeys } from '@/lib/obj_utils';
import { validateSlug } from '@/lib/text_utils';

interface AddToolBundleModalProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (slug: string, display: string, description?: string) => void;
	existingSlugs: string[];
}

const AddToolBundleModal: React.FC<AddToolBundleModalProps> = ({ isOpen, onClose, onSubmit, existingSlugs }) => {
	const [form, setForm] = useState({
		slug: '',
		displayName: '',
		description: '',
	});
	const [errors, setErrors] = useState<{ slug?: string; displayName?: string }>({});

	useEffect(() => {
		if (!isOpen) return;
		setForm({ slug: '', displayName: '', description: '' });
		setErrors({});
	}, [isOpen]);

	const validate = (field: keyof typeof errors, val: string) => {
		const v = val.trim();
		let copy = { ...errors };
		if (!v) {
			copy[field] = 'This field is required.';
		} else if (field === 'slug') {
			const err = validateSlug(v);
			if (err) {
				copy.slug = err;
			} else if (existingSlugs.includes(v)) {
				copy.slug = 'Slug already in use.';
			} else {
				copy = omitManyKeys(copy, ['slug']);
			}
		} else {
			copy = omitManyKeys(copy, [field]);
		}
		setErrors(copy);
	};

	const isAllValid = useMemo(
		() => form.slug.trim() && form.displayName.trim() && Object.keys(errors).length === 0,
		[form, errors]
	);

	const handleSubmit = (e?: React.FormEvent) => {
		if (e) e.preventDefault();
		validate('slug', form.slug);
		validate('displayName', form.displayName);
		if (!isAllValid) return;
		onSubmit(form.slug.trim(), form.displayName.trim(), form.description.trim() || undefined);
	};

	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box max-w-3xl max-h-[80vh] overflow-auto rounded-2xl">
				<div className="flex justify-between items-center mb-4">
					<h3 className="font-bold text-lg">Add Tool Bundle</h3>
					<button className="btn btn-sm btn-circle" onClick={onClose} aria-label="Close" title="Close">
						<FiX size={12} />
					</button>
				</div>
				<form onSubmit={handleSubmit} className="space-y-4">
					{/* Slug */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Bundle Slug*</span>
							<span className="label-text-alt tooltip tooltip-right" data-tip="Lower-case, URL-friendly.">
								<FiHelpCircle size={12} />
							</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								className={`input input-bordered w-full rounded-2xl ${errors.slug ? 'input-error' : ''}`}
								value={form.slug}
								onChange={e => {
									setForm(p => ({ ...p, slug: e.target.value }));
									validate('slug', e.target.value);
								}}
								spellCheck="false"
								autoComplete="off"
							/>
							{errors.slug && (
								<div className="label">
									<span className="label-text-alt text-error flex items-center gap-1">
										<FiAlertCircle size={12} /> {errors.slug}
									</span>
								</div>
							)}
						</div>
					</div>
					{/* Display Name */}
					<div className="grid grid-cols-12 items-center gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Display Name*</span>
						</label>
						<div className="col-span-9">
							<input
								type="text"
								className={`input input-bordered w-full rounded-2xl ${errors.displayName ? 'input-error' : ''}`}
								value={form.displayName}
								onChange={e => {
									setForm(p => ({ ...p, displayName: e.target.value }));
									validate('displayName', e.target.value);
								}}
								spellCheck="false"
								autoComplete="off"
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
					{/* Description */}
					<div className="grid grid-cols-12 items-start gap-2">
						<label className="label col-span-3">
							<span className="label-text text-sm">Description</span>
						</label>
						<div className="col-span-9">
							<textarea
								className="textarea textarea-bordered w-full rounded-2xl h-24"
								value={form.description}
								onChange={e => {
									setForm(p => ({ ...p, description: e.target.value }));
								}}
								spellCheck="false"
							/>
						</div>
					</div>
					{/* Actions */}
					<div className="modal-action">
						<button type="button" className="btn rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button type="submit" className="btn btn-primary rounded-2xl" disabled={!isAllValid}>
							Create
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default AddToolBundleModal;
