import React, { useEffect, useState } from 'react';

import type { DocStore } from '@/models/docstoremodel';

interface ModifyStoreProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (storeData: Partial<DocStore>) => void;
	initialData?: Partial<DocStore>;
	existingDocStores: DocStore[];
}

const ModifyDocStore: React.FC<ModifyStoreProps> = ({ isOpen, onClose, onSubmit, initialData, existingDocStores }) => {
	const [formData, setFormData] = useState<Partial<DocStore>>({
		name: '',
		url: '',
		description: '',
		dbName: '',
	});

	const [errors, setErrors] = useState<{
		name?: string;
		url?: string;
		dbName?: string;
	}>({});

	useEffect(() => {
		if (isOpen) {
			setFormData(
				initialData || {
					name: '',
					url: '',
					description: '',
					dbName: '',
				}
			);
			setErrors({});
		}
	}, [isOpen, initialData]);

	const validateFields = (name: string, value: string) => {
		const newErrors: { [key: string]: string | undefined } = {};

		// Check if field is empty
		if (!value.trim()) {
			newErrors[name] = 'This field is required';
		}

		// Check for uniqueness of URL and dbName combination
		if ((name === 'url' || name === 'dbName') && value.trim()) {
			const otherField = name === 'url' ? 'dbName' : 'url';
			const otherValue = formData[otherField as keyof typeof formData];

			const isUnique = !existingDocStores.some(
				s =>
					s.url === (name === 'url' ? value : otherValue) &&
					s.dbName === (name === 'dbName' ? value : otherValue) &&
					s.id !== initialData?.id
			);

			if (!isUnique) {
				newErrors[name] = `This combination of URL and Database Name already exists`;
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

	const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
		const { name, value } = e.target;
		setFormData(prev => ({ ...prev, [name]: value }));
		validateFields(name, value);
	};

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();

		// Validate all fields before submission
		const fieldsToValidate = ['name', 'url', 'dbName'] as const;
		fieldsToValidate.forEach(field => {
			const value = formData[field];
			if (typeof value === 'string') {
				validateFields(field, value);
			} else {
				validateFields(field, '');
			}
		});

		if (Object.keys(errors).length === 0) {
			onSubmit(formData);
		}
	};

	if (!isOpen) return null;

	return (
		<div className="modal modal-open">
			<div className="modal-box rounded-2xl">
				<h3 className="font-bold text-lg">{initialData ? 'Edit Document Store' : 'Add New Document Store'}</h3>
				<form onSubmit={handleSubmit} className="mt-4">
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							Name*
						</label>
						<input
							type="text"
							name="name"
							value={formData.name}
							onChange={handleChange}
							className={`input rounded-2xl ${errors.name ? 'input-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.name && <p className="text-error text-sm mt-1">{errors.name}</p>}
					</fieldset>
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							URL*
						</label>
						<input
							type="url"
							name="url"
							value={formData.url}
							onChange={handleChange}
							className={`input rounded-2xl ${errors.url ? 'input-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.url && <p className="text-error text-sm mt-1">{errors.url}</p>}
					</fieldset>
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							Description
						</label>
						<textarea
							name="description"
							value={formData.description}
							onChange={handleChange}
							className="textarea rounded-2xl"
							spellCheck="false"
						></textarea>
					</fieldset>
					<fieldset className="fieldset">
						<label className="label" htmlFor="name">
							Database Name*
						</label>
						<input
							type="text"
							name="dbName"
							value={formData.dbName}
							onChange={handleChange}
							className={`input rounded-2xl ${errors.dbName ? 'input-error' : ''}`}
							required
							spellCheck="false"
						/>
						{errors.dbName && <p className="text-error text-sm mt-1">{errors.dbName}</p>}
					</fieldset>
					<div className="modal-action">
						<button type="button" className="btn btn-ghost rounded-2xl" onClick={onClose}>
							Cancel
						</button>
						<button
							type="submit"
							className="btn btn-primary rounded-2xl"
							disabled={Object.keys(errors).length > 0 || !formData.name || !formData.url || !formData.dbName}
						>
							Save
						</button>
					</div>
				</form>
			</div>
		</div>
	);
};

export default ModifyDocStore;
