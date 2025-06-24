import type { FC } from 'react';
import { useEffect, useRef, useState } from 'react';

import mermaid from 'mermaid';
import { v4 as uuidv4 } from 'uuid';

import { Base64EncodeUTF8 } from '@/lib/encode_decode';

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
	const containerRef = useRef<HTMLDivElement | null>(null);
	const uniqueId = useRef(`mermaid-${uuidv4()}`);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		initializeMermaid();
		let isMounted = true;
		if (containerRef.current) {
			containerRef.current.innerHTML = '';
			mermaid
				.render(uniqueId.current, code)
				.then(renderResult => {
					if (isMounted && containerRef.current) {
						containerRef.current.innerHTML = renderResult.svg;
						// Center the SVG with inline styles
						const svg = containerRef.current.querySelector('svg');
						if (svg) {
							svg.style.display = 'block';
							svg.style.marginLeft = 'auto';
							svg.style.marginRight = 'auto';
							svg.style.width = 'auto';
							svg.style.height = 'auto';
							svg.style.maxHeight = '60vh';
							svg.style.maxWidth = '80%';
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

	const fetchDiagramAsBlob = async (): Promise<Blob> => {
		if (!containerRef.current) throw new Error('Container not found');
		const svg = containerRef.current.querySelector('svg');
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
		<div className="rounded-lg my-4 items-start overflow-hidden" style={{ backgroundColor: '#E5E9F0' }}>
			<div className="flex justify-between items-center bg-gray-700 px-4">
				<span className="text-white">Mermaid Diagram</span>
				<DownloadButton
					valueFetcher={fetchDiagramAsBlob}
					size={16}
					fileprefix="diagram"
					isBinary={true}
					language="mermaid"
					className="btn btn-sm bg-transparent text-white border-none flex items-center shadow-none"
				/>
			</div>
			<div className="flex items-center justify-center text-center p-1 min-h-[250px] w-full overflow-auto">
				{error ? (
					<svg width="300" height="100" style={{ background: '#fff', border: '2px solid #e53e3e', borderRadius: 8 }}>
						<rect x="0" y="0" width="300" height="100" fill="#fff" stroke="#e53e3e" strokeWidth="2" rx="8" />
						<text x="150" y="55" textAnchor="middle" fill="#e53e3e" fontSize="18" fontFamily="monospace">
							Mermaid syntax error
						</text>
					</svg>
				) : (
					<div ref={containerRef} className="w-full max-h-[60vh] overflow-auto" />
				)}
			</div>
		</div>
	);
};

export default MermaidDiagram;
