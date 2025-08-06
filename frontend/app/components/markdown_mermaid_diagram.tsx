import type { FC } from 'react';
import { useEffect, useRef, useState } from 'react';

import mermaid from 'mermaid';

import { Base64EncodeUTF8 } from '@/lib/encode_decode';
import { getUUIDv7 } from '@/lib/uuid_utils';

import DownloadButton from '@/components/download_button';

let mermaidInitialized = false;

function initializeMermaid() {
	if (!mermaidInitialized) {
		mermaid.initialize({ startOnLoad: false, theme: 'default', suppressErrorRendering: true, securityLevel: 'loose' });
		mermaidInitialized = true;
	}
}

interface MermaidDiagramProps {
	code: string;
}

const MermaidDiagram: FC<MermaidDiagramProps> = ({ code }) => {
	const inlineDiagramRef = useRef<HTMLDivElement | null>(null);
	const dialogRef = useRef<HTMLDialogElement | null>(null);
	const modalRef = useRef<HTMLDivElement | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [svgNode, setSvgNode] = useState<SVGSVGElement | null>(null);
	const [zoomOpen, setZoomOpen] = useState(false);

	const uniqueId = useRef(`mermaid-${getUUIDv7()}`);
	useEffect(() => {
		initializeMermaid();
		let isMounted = true;
		if (inlineDiagramRef.current) {
			inlineDiagramRef.current.innerHTML = '';
			mermaid
				.render(uniqueId.current, code)
				.then(renderResult => {
					if (isMounted && inlineDiagramRef.current) {
						inlineDiagramRef.current.innerHTML = renderResult.svg;
						// Center the SVG with inline styles
						const svg = inlineDiagramRef.current.querySelector('svg');
						if (svg) {
							svg.style.display = 'block';
							svg.style.margin = 'auto';
							svg.style.width = 'auto';
							svg.style.height = 'auto';
							svg.style.maxWidth = '80%';
							svg.style.maxHeight = '60vh';
							setSvgNode(svg.cloneNode(true) as SVGSVGElement);
						}
						setError(null);
					}
				})
				.catch(() => {
					if (isMounted) setError('Failed to render diagram. Please check the syntax.');
				});
		}
		return () => {
			isMounted = false;
		};
	}, [code]);

	useEffect(() => {
		const dlg = dialogRef.current;
		if (!dlg) return;

		if (zoomOpen && !dlg.open) {
			dlg.showModal();
		} else if (!zoomOpen && dlg.open) {
			dlg.close();
		}
	}, [zoomOpen]);

	/* When the dialog fires its native `close` event, update state */
	useEffect(() => {
		const dlg = dialogRef.current;
		if (!dlg) return;

		const handleClose = () => {
			setZoomOpen(false);
		};
		dlg.addEventListener('close', handleClose);
		return () => {
			dlg.removeEventListener('close', handleClose);
		};
	}, []);

	useEffect(() => {
		if (!zoomOpen || !modalRef.current || !svgNode) return;

		// Clean previous content
		modalRef.current.innerHTML = '';

		const newNode = svgNode.cloneNode(true) as SVGSVGElement;
		newNode.style.display = 'block';
		newNode.style.margin = 'auto';
		newNode.style.width = 'auto';
		newNode.style.height = 'auto';
		newNode.style.maxWidth = '90vw';
		newNode.style.maxHeight = '80vh';

		modalRef.current.appendChild(newNode);
	}, [zoomOpen, svgNode]);

	const fetchDiagramAsBlob = async (): Promise<Blob> => {
		if (!inlineDiagramRef.current) throw new Error('Container not found');
		const svg = inlineDiagramRef.current.querySelector('svg');
		if (!svg) throw new Error('SVG element not found in container');

		const svgData = new XMLSerializer().serializeToString(svg);
		// Encode SVG as base64 data URL
		const svgBase64 = Base64EncodeUTF8(svgData);
		const dataUrl = `data:image/svg+xml;base64,${svgBase64}`;

		return new Promise<Blob>((resolve, reject) => {
			const img = new window.Image();
			img.crossOrigin = 'anonymous';
			img.onload = () => {
				const scaleFactor = 2;
				const canvas = document.createElement('canvas');
				canvas.width = img.width * scaleFactor;
				canvas.height = img.height * scaleFactor;
				const ctx = canvas.getContext('2d');
				if (!ctx) {
					reject(new Error('Canvas context is null'));
					return;
				}
				ctx.scale(scaleFactor, scaleFactor);
				ctx.fillStyle = 'white';
				ctx.fillRect(0, 0, canvas.width, canvas.height);
				ctx.drawImage(img, 0, 0);
				canvas.toBlob(
					blob => {
						if (blob) resolve(blob);
						else reject(new Error('Canvas is empty'));
					},
					'image/png',
					1.0
				);
			};
			img.onerror = err => {
				reject(err);
			};
			img.src = dataUrl;
		});
	};
	return (
		<>
			{/* ---------- Inline card -------------------------------- */}
			<div className="rounded-lg my-4 overflow-hidden bg-mermaid">
				{/* header bar */}
				<div className="flex justify-between items-center bg-code-header px-4">
					<span className="text-code">Mermaid diagram</span>

					<DownloadButton
						valueFetcher={fetchDiagramAsBlob}
						size={16}
						fileprefix="diagram"
						isBinary={true}
						language="mermaid"
						className="btn btn-sm bg-transparent text-code border-none flex items-center shadow-none"
					/>
				</div>

				<div
					className="flex items-center justify-center text-center p-1 min-h-[250px] w-full overflow-auto cursor-zoom-in"
					onClick={() => {
						if (!error) {
							setZoomOpen(true);
						}
					}}
				>
					{error ? (
						<svg width="300" height="100" className="border-2 border-red-500 rounded">
							<rect width="300" height="100" fill="#fff" stroke="#e53e3e" strokeWidth="2" rx="8" />
							<text x="150" y="55" textAnchor="middle" fill="#e53e3e" fontSize="18" fontFamily="monospace">
								Mermaid syntax error
							</text>
						</svg>
					) : (
						<div ref={inlineDiagramRef} className="w-full max-h-[60vh] overflow-auto" />
					)}
				</div>
			</div>

			{/* ---------- Zoom modal (daisyUI v5) -------------------- */}
			<dialog ref={dialogRef} className="modal" aria-label="Enlarged Mermaid diagram">
				{/* backdrop - clicking it closes the dialog */}
				<form method="dialog" className="modal-backdrop">
					<button aria-label="Close"></button>
				</form>

				{/* modal box */}
				<div
					className="modal-box max-w-[90vw] h-[90vh] cursor-zoom-out flex items-center justify-center"
					onClick={() => {
						setZoomOpen(false);
					}}
				>
					{/* enlarged diagram */}
					<div ref={modalRef} className="overflow-auto w-full" style={{ pointerEvents: 'none' }} />
				</div>
			</dialog>
		</>
	);
};

export default MermaidDiagram;
