import { useEffect, useRef } from 'react';

import { createPortal } from 'react-dom';

import { FiTool, FiX } from 'react-icons/fi';

import type { ToolOutput } from '@/spec/tool';

interface ToolOutputModalProps {
	isOpen: boolean;
	onClose: () => void;
	output: ToolOutput | null;
}

/**
 * Simple modal to inspect a single tool-output payload.
 * Shows pretty-printed JSON when possible, otherwise raw text.
 */
export function ToolOutputModal({ isOpen, onClose, output }: ToolOutputModalProps) {
	const dialogRef = useRef<HTMLDialogElement | null>(null);

	// Open / close native dialog when isOpen changes
	useEffect(() => {
		if (!isOpen || !output) return;

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
	}, [isOpen, output]);

	const handleDialogClose = () => {
		onClose();
	};

	if (!isOpen || !output) return null;

	let renderedBody = 'Tool returned no output.';
	if (output.rawOutput && output.rawOutput.trim() !== '') {
		const raw = output.rawOutput.trim();
		try {
			const parsed = JSON.parse(raw);
			renderedBody = JSON.stringify(parsed, null, 2);
		} catch {
			renderedBody = raw;
		}
	}

	return createPortal(
		<dialog ref={dialogRef} className="modal" onClose={handleDialogClose}>
			<div className="modal-box bg-base-200 max-h-[80vh] max-w-[80vw] overflow-hidden rounded-2xl p-0">
				<div className="max-h-[80vh] overflow-y-auto p-6">
					{/* header */}
					<div className="mb-4 flex items-center justify-between">
						<h3 className="flex items-center gap-2 text-lg font-bold">
							<FiTool size={16} />
							<span>{output.summary}</span>
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

					<div className="text-base-content/70 mb-2 text-xs">
						<div>
							<strong>Tool:</strong> {output.name}
						</div>
						<div>
							<strong>Call ID:</strong> {output.callID}
						</div>
					</div>

					<pre className="bg-base-300/40 text-base-content max-h-[60vh] overflow-auto rounded-xl p-3 text-xs whitespace-pre-wrap">
						{renderedBody}
					</pre>
				</div>
			</div>

			{/* DaisyUI backdrop */}
			<form method="dialog" className="modal-backdrop">
				<button aria-label="Close" />
			</form>
		</dialog>,
		document.body
	);
}
