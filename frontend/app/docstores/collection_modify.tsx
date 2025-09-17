import React, { useEffect, useState } from 'react';

import { DOCUMENT_COLLECTION_INVOKE_CHAR } from '@/spec/command';
import type { Collection } from '@/spec/docstore';

interface ModifyCollectionProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (collectionData: Partial<Collection>, docStoreID: string) => void;
	initialData?: Partial<Collection>;
	docStoreID: string;
	existingCollections: Collection[];
}

const ModifyCollection: React.FC<ModifyCollectionProps> = ({
	isOpen,
	onClose,
	onSubmit,
	initialData,
	docStoreID,
	existingCollections,
}) => {
	const [formData, setFormData] = useState<Partial<Collection>>({
		name: '',
		command: '',
	});
	const [errors, setErrors] = useState<{ name?: string; command?: string }>({});

	useEffect(() => {
		if (isOpen) {
			setFormData(initialData || { name: '', command: '' });
			setErrors({});
		}
	}, [isOpen, initialData]);

	const validateField = (name: string, value: string) => {
		const newErrors: { [key: string]: string | undefined } = {};

		if (!value.trim()) {
			newErrors[name] = 'This field is required';
		} else if (name === 'command' && value.startsWith(DOCUMENT_COLLECTION_INVOKE_CHAR)) {
			newErrors[name] = `Command should not start with ${DOCUMENT_COLLECTION_INVOKE_CHAR}`;
		} else {
			const isUnique = !existingCollections.some(
				c => c[name as keyof Collection] === value && c.id !== initialData?.id
			);
			if (!isUnique) {
				newErrors[name] = `This ${name} is already in use`;
			}
		}

		// Populate newErrors with existing errors except for the current field if it's valid
		for (const key in errors) {
			if (Object.prototype.hasOwnProperty.call(errors, key) && key !== name) {
				newErrors[key] = errors[key as keyof typeof errors];
			}
		}

		setErrors(newErrors);
	};

	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const { name, value } = e.target;
		setFormData(prev => ({ ...prev, [name]: value }));
		validateField(name, value);
	};

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		validateField('name', formData.name || '');
		validateField('command', formData.command || '');

		if (Object.keys(errors).length === 0 && formData.name && formData.command) {
			onSubmit(formData, docStoreID);
		}
	};

	if (!isOpen) return null;

	return (
		<dialog className="modal modal-open">
			<div className="modal-box bg-base-200 rounded-2xl">
				<h3 className="text-lg font-bold">{initialData ? 'Edit Collection' : 'Add New Collection'}</h3>
				<form onSubmit={handleSubmit} className="mt-4">
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							Name*
						</label>
						<input
							type="text"
							name="name"
							value={formData.name || ''}
							onChange={handleChange}
							className={`input rounded-2xl ${errors.name ? 'input-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.name && <p className="text-error mt-1 text-sm">{errors.name}</p>}
					</fieldset>
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							Command*
						</label>
						<div className="relative">
							<span className="text-neutral-custom absolute top-1/2 left-3 -translate-y-1/2 transform">
								{DOCUMENT_COLLECTION_INVOKE_CHAR}
							</span>
							<input
								type="text"
								name="command"
								value={formData.command || ''}
								onChange={handleChange}
								className={`input rounded-2xl pl-8 ${errors.command ? 'input-error' : ''}`}
								required
								spellCheck="false"
							/>
						</div>
						{errors.command && <p className="text-error mt-1 text-sm">{errors.command}</p>}
					</fieldset>
					<div className="modal-action">
						<button type="button" className="btn bg-base-300 rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button
							type="submit"
							className="btn btn-primary rounded-2xl"
							disabled={!!errors.name || !!errors.command || !formData.name || !formData.command}
						>
							Save
						</button>
					</div>
				</form>
			</div>
		</dialog>
	);
};

export default ModifyCollection;
