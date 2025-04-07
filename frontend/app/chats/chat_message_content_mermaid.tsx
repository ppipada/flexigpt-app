import DownloadButton from '@/components/download_button';
import 'katex/dist/katex.min.css';
import mermaid from 'mermaid';
import type { FC } from 'react';
import { useEffect, useRef, useState } from 'react';
import { v4 as uuidv4 } from 'uuid';

let mermaidInitialized = false;

function initializeMermaid() {
	if (!mermaidInitialized) {
		mermaid.initialize({ startOnLoad: false, theme: 'default', suppressErrorRendering: true });
		mermaidInitialized = true;
	}
}

interface MermaidDiagramProps {
	code: string;
}

export const MermaidDiagram: FC<MermaidDiagramProps> = ({ code }) => {
	const containerRef = useRef<HTMLDivElement | null>(null);
	const uniqueId = useRef(`mermaid-${uuidv4()}`);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		initializeMermaid();
		if (containerRef.current) {
			// Clear the container before rendering
			containerRef.current.innerHTML = '';
			mermaid
				.render(uniqueId.current, code, containerRef.current)
				.then(renderResult => {
					if (containerRef.current) {
						containerRef.current.innerHTML = renderResult.svg;
						setError(null); // Clear any previous errors
					}
				})
				.catch((err: unknown) => {
					console.error('Mermaid rendering failed:', err);
					setError('Failed to render diagram. Please check the syntax.');
				});
		}
	}, [code]);

	const fetchDiagramAsBlob = async (): Promise<Blob> => {
		if (containerRef.current) {
			const svg = containerRef.current.querySelector('svg');
			if (svg) {
				const svgData = new XMLSerializer().serializeToString(svg);

				return new Promise<Blob>((resolve, reject) => {
					const canvas = document.createElement('canvas');
					const ctx = canvas.getContext('2d');
					const img = new Image();

					img.onload = () => {
						const scaleFactor = 2;
						canvas.width = img.width * scaleFactor;
						canvas.height = img.height * scaleFactor;

						if (ctx) {
							ctx.scale(scaleFactor, scaleFactor);
							ctx.fillStyle = 'white';
							ctx.fillRect(0, 0, canvas.width, canvas.height);
							ctx.drawImage(img, 0, 0);
						}

						canvas.toBlob(
							blob => {
								if (blob) {
									resolve(blob);
								} else {
									reject(new Error('Canvas is empty'));
								}
							},
							'image/png',
							1.0
						);
					};

					img.onerror = err => {
						reject(err);
					};

					img.src = 'data:image/svg+xml;base64,' + btoa(encodeURIComponent(svgData));
				});
			} else {
				throw new Error('SVG element not found in container');
			}
		} else {
			throw new Error('Container not found');
		}
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
			<div className="flex justify-center text-center p-1" ref={containerRef} style={{ overflow: 'auto' }}>
				{error && <div className="text-red-500">{error}</div>}
			</div>
		</div>
	);
};
