import DownloadButton from '@/components/download_button';
import 'katex/dist/katex.min.css';
import mermaid from 'mermaid';
import { FC, useEffect, useRef } from 'react';
import { v4 as uuidv4 } from 'uuid';

let mermaidInitialized = false;

function initializeMermaid() {
	if (!mermaidInitialized) {
		mermaid.initialize({ startOnLoad: false, theme: 'default' });
		mermaidInitialized = true;
	}
}
interface MermaidDiagramProps {
	code: string;
}

export const MermaidDiagram: FC<MermaidDiagramProps> = ({ code }) => {
	const containerRef = useRef<HTMLDivElement | null>(null);
	const uniqueId = useRef(`mermaid-${uuidv4()}`); // Generate a unique ID

	useEffect(() => {
		initializeMermaid();
		if (containerRef.current) {
			mermaid.initialize({
				startOnLoad: true,
				theme: 'default',
			});
			mermaid
				.render(uniqueId.current, code)
				.then(renderResult => {
					if (containerRef.current) {
						containerRef.current.innerHTML = renderResult.svg;
					}
				})
				.catch(error => {
					console.error('Mermaid rendering failed:', error);
				});
		}
	}, [code]);

	const fetchDiagramAsBlob = async (): Promise<Blob> => {
		if (containerRef.current) {
			const svg = containerRef.current.querySelector('svg');
			if (svg) {
				const svgData = new XMLSerializer().serializeToString(svg);

				// Convert SVG to PNG using a canvas
				return new Promise<Blob>((resolve, reject) => {
					const canvas = document.createElement('canvas');
					const ctx = canvas.getContext('2d');
					const img = new Image();

					img.onload = () => {
						// Set the desired resolution scale factor
						const scaleFactor = 2; // Adjust this value for higher/lower resolution

						// Set canvas dimensions based on the scale factor
						canvas.width = img.width * scaleFactor;
						canvas.height = img.height * scaleFactor;

						if (ctx) {
							// Scale the drawing context
							ctx.scale(scaleFactor, scaleFactor);

							// Fill the background with white
							ctx.fillStyle = 'white';
							ctx.fillRect(0, 0, canvas.width, canvas.height);

							// Draw the image onto the canvas
							ctx.drawImage(img, 0, 0);
						}

						// Convert the canvas to a Blob
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

					// Set the source of the image to the SVG data URL
					img.src = 'data:image/svg+xml;base64,' + btoa(unescape(encodeURIComponent(svgData)));
				});
			} else {
				throw new Error('SVG element not found in container');
			}
		} else {
			throw new Error('Container not found');
		}
	};

	return (
		<div className="rounded-md my-4 items-start overflow-hidden" style={{ backgroundColor: '#E5E9F0' }}>
			<div className="flex justify-between items-center bg-gray-700 px-4">
				<span className="text-white">Mermaid Diagram</span>
				<DownloadButton
					valueFetcher={fetchDiagramAsBlob}
					size={16}
					fileprefix="diagram"
					isBinary={true} // Important for binary content
					language="mermaid"
					className="btn btn-sm bg-transparent text-white border-none flex items-center shadow-none"
				/>
			</div>
			<div className="flex justify-center text-center p-1" ref={containerRef}></div>
		</div>
	);
};
