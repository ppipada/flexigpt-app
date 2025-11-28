import { type FormEvent, useEffect, useState } from 'react';

import { createPortal } from 'react-dom';

import { FiAlertCircle, FiLink, FiX } from 'react-icons/fi';

type UrlAttachmentModalProps = {
	isOpen: boolean;
	onClose: () => void;
	onAttachURL: (url: string) => Promise<void> | void;
};

export function UrlAttachmentModal({ isOpen, onClose, onAttachURL }: UrlAttachmentModalProps) {
	const [urlValue, setUrlValue] = useState('');
	const [error, setError] = useState<string | null>(null);
	const [submitting, setSubmitting] = useState(false);

	// Reset local state whenever the modal is opened
	useEffect(() => {
		if (!isOpen) return;
		setUrlValue('');
		setError(null);
		setSubmitting(false);
	}, [isOpen]);

	/**
	 * Normalize & validate URL.
	 * - returns normalized URL string if valid
	 * - throws Error with message if invalid
	 */
	const normalizeUrl = (value: string): string | null => {
		const raw = value.trim();
		if (!raw) return null; // caller decides if blank is allowed

		// Only one URL at a time — reject any whitespace
		if (/\s/.test(raw)) {
			throw new Error('Only one URL is allowed at a time.');
		}

		const candidate = /^https?:\/\//i.test(raw) ? raw : `https://${raw}`;

		let parsed: URL;
		try {
			parsed = new URL(candidate);
		} catch {
			throw new Error('Please enter a valid URL (for example, https://example.com).');
		}

		if (parsed.protocol !== 'http:' && parsed.protocol !== 'https:') {
			throw new Error('Only HTTP and HTTPS URLs are supported.');
		}

		return parsed.toString();
	};

	const handleChange = (value: string) => {
		setUrlValue(value);

		// Inline validation:
		// - allow blank without error
		// - error only when non-blank but invalid
		const trimmed = value.trim();
		if (!trimmed) {
			setError(null);
			return;
		}

		try {
			normalizeUrl(value);
			setError(null);
		} catch (err) {
			setError((err as Error).message);
		}
	};

	const handleSubmit = async (e: FormEvent) => {
		e.preventDefault();

		try {
			const normalized = normalizeUrl(urlValue);

			if (!normalized) {
				setError('Please enter a URL.');
				return;
			}

			setSubmitting(true);
			await onAttachURL(normalized);
			onClose();
		} catch (err) {
			setError((err as Error).message || 'Please enter a valid URL.');
		} finally {
			setSubmitting(false);
		}
	};

	if (!isOpen) return null;

	return createPortal(
		<dialog className="modal modal-open">
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-md overflow-auto rounded-2xl">
				{/* header */}
				<div className="mb-4 flex items-center justify-between">
					<h3 className="flex items-center gap-2 text-lg font-bold">
						<FiLink size={16} />
						<span>Attach Link</span>
					</h3>
					<button className="btn btn-sm btn-circle bg-base-300" onClick={onClose} aria-label="Close">
						<FiX size={12} />
					</button>
				</div>

				<form onSubmit={handleSubmit} className="space-y-4">
					{/* URL input */}
					<div>
						<label className="label p-1">
							<span className="label-text text-sm">URL</span>
						</label>
						<input
							autoFocus
							type="url"
							value={urlValue}
							onChange={e => {
								handleChange(e.target.value);
							}}
							className={`input input-bordered w-full rounded-xl ${error ? 'input-error' : ''}`}
							placeholder="https://example.com/resource"
							spellCheck="false"
						/>
						<p className="text-base-content/70 p-1 text-xs">Paste a single URL to attach to this message.</p>

						{/* Fixed-height error area to avoid layout shift */}
						<div className="mt-1 h-5 text-xs">
							{error && (
								<span className="text-error flex items-center gap-1">
									<FiAlertCircle size={12} /> {error}
								</span>
							)}
						</div>
					</div>

					{/* footer buttons */}
					<div className="modal-action">
						<button type="button" className="btn bg-base-300 rounded-xl" onClick={onClose}>
							Cancel
						</button>
						<button
							type="submit"
							className="btn btn-primary rounded-xl"
							disabled={submitting || !urlValue.trim() || !!error}
						>
							{submitting ? (
								<>
									<span className="loading loading-spinner loading-xs" />
									Attaching…
								</>
							) : (
								'Attach'
							)}
						</button>
					</div>
				</form>
			</div>
		</dialog>,
		document.body
	);
}
