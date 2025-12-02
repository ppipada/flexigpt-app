import { type FormEvent, useEffect, useRef, useState } from 'react';

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

	const dialogRef = useRef<HTMLDialogElement | null>(null);
	const inputRef = useRef<HTMLInputElement | null>(null);

	// Reset local state whenever the modal is opened
	useEffect(() => {
		if (!isOpen) return;
		setUrlValue('');
		setError(null);
		setSubmitting(false);
	}, [isOpen]);

	useEffect(() => {
		if (!isOpen) return;

		const dialog = dialogRef.current;
		if (!dialog) return;

		if (!dialog.open) {
			dialog.showModal();
		}

		return () => {
			// If the component unmounts while the dialog is still open, close it.
			if (dialog.open) {
				dialog.close();
			}
		};
	}, [isOpen]);

	// Keep parent isOpen in sync with native dialog closing
	const handleDialogClose = () => {
		onClose();
	};

	/**
	 * Normalize & validate URL.
	 */
	const normalizeUrl = (value: string): string | null => {
		const raw = value.trim();
		if (!raw) return null; // caller decides if blank is allowed

		// Only one URL at a time — reject any whitespace
		if (/\s/.test(raw)) {
			throw new Error('Only one URL is allowed at a time.');
		}

		// If user did not type a scheme, treat it as shorthand for https://
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

		const hostname = parsed.hostname;

		if (!hostname) {
			throw new Error('Please enter a valid URL host (for example, example.com).');
		}

		// Extra hardening: require a "real" hostname:
		// - must have a dot, OR be "localhost"
		if (!hostname.includes('.') && hostname !== 'localhost') {
			throw new Error('Please enter a full domain (for example, example.com).');
		}

		return parsed.toString();
	};

	/**
	 * Handle input change & validation.
	 */
	const handleChange = (value: string) => {
		setUrlValue(value);

		const trimmed = value.trim();
		if (!trimmed) {
			setError(null);
			return;
		}

		const input = inputRef.current;

		// For explicit http(s) URLs, let the browser validate first.
		if (/^https?:\/\//i.test(trimmed) && input && !input.validity.valid) {
			setError('Please enter a valid URL, for example: https://example.com');
			return;
		}

		try {
			normalizeUrl(trimmed);
			setError(null);
		} catch (err) {
			setError((err as Error).message);
		}
	};

	const handleSubmit = async (e: FormEvent) => {
		e.preventDefault();

		const trimmed = urlValue.trim();
		const input = inputRef.current;

		// Re-check browser validity for explicit http(s) URLs on submit
		if (/^https?:\/\//i.test(trimmed) && input && !input.checkValidity()) {
			setError('Please enter a valid URL, for example: https://example.com');
			input.focus();
			return;
		}

		try {
			const normalized = normalizeUrl(urlValue);

			if (!normalized) {
				setError('Please enter a URL.');
				return;
			}

			setSubmitting(true);
			await onAttachURL(normalized);

			// Close the dialog; this will trigger handleDialogClose -> parent onClose().
			dialogRef.current?.close();
		} catch (err) {
			setError((err as Error).message || 'Please enter a valid URL.');
		} finally {
			setSubmitting(false);
		}
	};

	if (!isOpen) return null;

	return createPortal(
		<dialog ref={dialogRef} className="modal" onClose={handleDialogClose}>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-md overflow-auto rounded-2xl">
				{/* header */}
				<div className="mb-4 flex items-center justify-between">
					<h3 className="flex items-center gap-2 text-lg font-bold">
						<FiLink size={16} />
						<span>Attach Link</span>
					</h3>
					<button
						type="button"
						className="btn btn-sm btn-circle bg-base-300"
						onClick={() => dialogRef.current?.close()}
						aria-label="Close"
					>
						<FiX size={12} />
					</button>
				</div>

				{/* NOTE: noValidate disables the browser's popup UI,
            but we still read input.validity/checkValidity() in JS. */}
				<form noValidate onSubmit={handleSubmit} className="space-y-4">
					{/* URL input */}
					<div>
						<label className="label p-1">
							<span className="label-text text-sm">URL</span>
						</label>
						<input
							ref={inputRef}
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
						<button type="button" className="btn bg-base-300 rounded-xl" onClick={() => dialogRef.current?.close()}>
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
			{/* NOTE: no modal-backdrop here: backdrop click should NOT close this modal */}
		</dialog>,
		document.body
	);
}
