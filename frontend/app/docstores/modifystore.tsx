import { DocStoreServer } from '@/models/docstoremodel';
import React, { useEffect, useState } from 'react';

interface ModifyStoreProps {
	isOpen: boolean;
	onClose: () => void;
	onSubmit: (storeData: Partial<DocStoreServer>) => void;
	initialData?: Partial<DocStoreServer>;
	existingServers: DocStoreServer[];
}

const ModifyDocStore: React.FC<ModifyStoreProps> = ({ isOpen, onClose, onSubmit, initialData, existingServers }) => {
	const [formData, setFormData] = useState<Partial<DocStoreServer>>({
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
		const newErrors: typeof errors = { ...errors };

		// Check if field is empty
		if (!value.trim()) {
			newErrors[name as keyof typeof errors] = 'This field is required';
		} else {
			delete newErrors[name as keyof typeof errors];
		}

		// Check for uniqueness of URL and dbName combination
		if ((name === 'url' || name === 'dbName') && value.trim()) {
			const otherField = name === 'url' ? 'dbName' : 'url';
			const otherValue = formData[otherField as keyof typeof formData];

			const isUnique = !existingServers.some(
				s =>
					s.url === (name === 'url' ? value : otherValue) &&
					s.dbName === (name === 'dbName' ? value : otherValue) &&
					s.id !== initialData?.id
			);

			if (!isUnique) {
				newErrors[name as keyof typeof errors] = `This combination of URL and Database Name already exists`;
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
					<div className="form-control">
						<label className="label">
							<span className="label-text">Name*</span>
						</label>
						<input
							type="text"
							name="name"
							value={formData.name}
							onChange={handleChange}
							className={`input input-bordered rounded-2xl ${errors.name ? 'input-error' : ''}`}
							required
						/>
						{errors.name && <p className="text-error text-sm mt-1">{errors.name}</p>}
					</div>
					<div className="form-control">
						<label className="label">
							<span className="label-text">URL*</span>
						</label>
						<input
							type="url"
							name="url"
							value={formData.url}
							onChange={handleChange}
							className={`input input-bordered rounded-2xl ${errors.url ? 'input-error' : ''}`}
							required
						/>
						{errors.url && <p className="text-error text-sm mt-1">{errors.url}</p>}
					</div>
					<div className="form-control">
						<label className="label">
							<span className="label-text">Description</span>
						</label>
						<textarea
							name="description"
							value={formData.description}
							onChange={handleChange}
							className="textarea textarea-bordered rounded-2xl"
						></textarea>
					</div>
					<div className="form-control">
						<label className="label">
							<span className="label-text">Database Name*</span>
						</label>
						<input
							type="text"
							name="dbName"
							value={formData.dbName}
							onChange={handleChange}
							className={`input input-bordered rounded-2xl ${errors.dbName ? 'input-error' : ''}`}
							required
						/>
						{errors.dbName && <p className="text-error text-sm mt-1">{errors.dbName}</p>}
					</div>
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
