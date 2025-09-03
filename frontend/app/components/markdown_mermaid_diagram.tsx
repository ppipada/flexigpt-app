/* eslint-disable @typescript-eslint/no-unnecessary-condition */
import type { FC } from 'react';
import { useEffect, useMemo, useRef, useState } from 'react';

import { Base64EncodeUTF8 } from '@/lib/encode_decode';
import { loadMermaid, type MermaidConfig } from '@/lib/lazy_mermaid_loader';
import { getUUIDv7 } from '@/lib/uuid_utils';

import { useIsDarkMermaid } from '@/hooks/use_is_dark_mermaid';

import DownloadButton from '@/components/download_button';

interface MermaidDiagramProps {
	code: string;
}

const MermaidDiagram: FC<MermaidDiagramProps> = ({ code }) => {
	const isDark = useIsDarkMermaid();

	const inlineDiagramRef = useRef<HTMLDivElement | null>(null);
	const dialogRef = useRef<HTMLDialogElement | null>(null);
	const modalRef = useRef<HTMLDivElement | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [svgNode, setSvgNode] = useState<SVGSVGElement | null>(null);
	const [zoomOpen, setZoomOpen] = useState(false);

	const mermaidConfig = useMemo<MermaidConfig>(
		() => ({
			startOnLoad: false,
			theme: isDark ? 'dark' : 'default',
			suppressErrorRendering: true,
			securityLevel: 'loose',
		}),
		[isDark]
	);

	// Ensure initialize happens before render, every time we render (safe to call repeatedly).
	// Also make each render id unique to avoid collisions.
	const renderSeq = useRef(0);
	useEffect(() => {
		const container = inlineDiagramRef.current;
		if (!container) return;

		const seq = ++renderSeq.current;
		let cancelled = false;

		// Clear old content and optimistic-clear error to avoid sticky error box
		container.innerHTML = '';
		setError(null);

		(async () => {
			try {
				const mermaid = await loadMermaid();
				if (cancelled || renderSeq.current !== seq) return;

				mermaid.initialize(mermaidConfig); // safe to call on every render

				// Optional: validate first for clearer failure and to avoid partial DOM updates
				await mermaid.parse(code);

				const id = `mermaid-${getUUIDv7()}`; // unique per render call
				const { svg } = await mermaid.render(id, code);

				if (cancelled || renderSeq.current !== seq) return;
				if (!inlineDiagramRef.current) return;

				inlineDiagramRef.current.innerHTML = svg;

				const svgEl = inlineDiagramRef.current.querySelector('svg');
				if (svgEl) {
					svgEl.style.display = 'block';
					svgEl.style.margin = 'auto';
					svgEl.style.width = 'auto';
					svgEl.style.height = 'auto';
					svgEl.style.maxWidth = '80%';
					svgEl.style.maxHeight = '60vh';
					setSvgNode(svgEl.cloneNode(true) as SVGSVGElement);
				}
				setError(null);
			} catch (e) {
				if (!cancelled && renderSeq.current === seq) {
					setError('Failed to render diagram. Please check the syntax.');
					console.error('syntax error:', e);
				} else {
					console.error('mermaid render error:', e);
				}
			}
		})();

		return () => {
			cancelled = true;
		};
	}, [code, mermaidConfig]);

	useEffect(() => {
		const dlg = dialogRef.current;
		if (!dlg) return;

		if (zoomOpen && !dlg.open) {
			dlg.showModal();
		} else if (!zoomOpen && dlg.open) {
			dlg.close();
		}
	}, [zoomOpen]);

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
			<div className="bg-mermaid my-4 overflow-hidden rounded-lg">
				<div className="bg-code-header flex items-center justify-between px-4">
					<span className="text-code">Mermaid diagram</span>
					<DownloadButton
						valueFetcher={fetchDiagramAsBlob}
						size={16}
						fileprefix="diagram"
						isBinary={true}
						language="mermaid"
						className="btn btn-sm text-code flex items-center border-none bg-transparent shadow-none"
					/>
				</div>

				<div
					className="flex min-h-[250px] w-full cursor-zoom-in items-center justify-center overflow-auto p-1 text-center"
					onClick={() => {
						if (!error) setZoomOpen(true);
					}}
				>
					{error ? (
						<svg width="300" height="100" className="rounded border-2 border-red-500">
							<rect width="300" height="100" fill="#fff" stroke="#e53e3e" strokeWidth="2" rx="8" />
							<text x="150" y="55" textAnchor="middle" fill="#e53e3e" fontSize="18" fontFamily="monospace">
								Mermaid syntax error
							</text>
						</svg>
					) : (
						<div ref={inlineDiagramRef} className="max-h-[60vh] w-full overflow-auto" />
					)}
				</div>
			</div>

			<dialog ref={dialogRef} className="modal" aria-label="Enlarged Mermaid diagram">
				<form method="dialog" className="modal-backdrop">
					<button aria-label="Close"></button>
				</form>
				<div
					className="modal-box bg-mermaid flex h-[90vh] max-w-[90vw] cursor-zoom-out items-center justify-center"
					onClick={() => {
						setZoomOpen(false);
					}}
				>
					<div ref={modalRef} className="w-full overflow-auto" style={{ pointerEvents: 'none' }} />
				</div>
			</dialog>
		</>
	);
};

export default MermaidDiagram;
