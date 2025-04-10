// DownloadButton.tsx
import type { ButtonHTMLAttributes, FC } from 'react';
import { FiDownload } from 'react-icons/fi';

import type { FileFilter } from '@/models/backendmodel';

import { backendAPI } from '@/apis/baseapi';

interface LanguageMap {
	[key: string]: { extension: string; mimeType: string };
}

/** @lintignore */
export const ProgrammingLanguages: LanguageMap = {
	javascript: { extension: '.js', mimeType: 'application/javascript' },
	python: { extension: '.py', mimeType: 'text/x-python' },
	java: { extension: '.java', mimeType: 'text/x-java-source' },
	c: { extension: '.c', mimeType: 'text/x-c' },
	cpp: { extension: '.cpp', mimeType: 'text/x-c++' },
	'c++': { extension: '.cpp', mimeType: 'text/x-c++' },
	'c#': { extension: '.cs', mimeType: 'text/x-csharp' },
	ruby: { extension: '.rb', mimeType: 'text/x-ruby' },
	php: { extension: '.php', mimeType: 'application/php' },
	swift: { extension: '.swift', mimeType: 'text/x-swift' },
	'objective-c': { extension: '.m', mimeType: 'text/x-objectivec' },
	kotlin: { extension: '.kt', mimeType: 'text/x-kotlin' },
	typescript: { extension: '.ts', mimeType: 'application/typescript' },
	go: { extension: '.go', mimeType: 'text/x-go' },
	perl: { extension: '.pl', mimeType: 'text/x-perl' },
	rust: { extension: '.rs', mimeType: 'text/x-rustsrc' },
	scala: { extension: '.scala', mimeType: 'text/x-scala' },
	haskell: { extension: '.hs', mimeType: 'text/x-haskell' },
	lua: { extension: '.lua', mimeType: 'text/x-lua' },
	shell: { extension: '.sh', mimeType: 'application/x-sh' },
	sql: { extension: '.sql', mimeType: 'application/sql' },
	html: { extension: '.html', mimeType: 'text/html' },
	css: { extension: '.css', mimeType: 'text/css' },
	json: { extension: '.json', mimeType: 'application/json' },
	dart: { extension: '.dart', mimeType: 'application/dart' },
	// Add more file extensions and MIME types here
};

// Determine the file extension from the MIME type
const mimeTypeMap: { [key: string]: string } = {
	'image/png': '.png',
	'image/jpeg': '.jpg',
	'image/gif': '.gif',
	'application/pdf': '.pdf',
	// Add more MIME types and extensions as needed
};

function generateRandomString(length: number, lowercase = false): string {
	const chars = 'ABCDEFGHJKLMNPQRSTUVWXY3456789'; // Exclude similar characters
	let result = '';
	for (let i = 0; i < length; i++) {
		result += chars.charAt(Math.floor(Math.random() * chars.length));
	}
	return lowercase ? result.toLowerCase() : result;
}

async function saveFile(defaultFilename: string, contentBase64: string, filters: Array<FileFilter>): Promise<void> {
	await backendAPI.savefile(defaultFilename, contentBase64, filters);
}

interface DownloadButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	language?: string;
	valueFetcher: () => Promise<string | Blob>;
	size: number;
	fileprefix?: string;
	isBinary?: boolean; // Indicates if content is binary (e.g., images)
}

const DownloadButton: FC<DownloadButtonProps> = ({
	language = '',
	valueFetcher,
	size,
	fileprefix = 'file',
	isBinary = false,
	...buttonProps
}) => {
	const downloadAsFile = async () => {
		const value = await valueFetcher(); // Can be string or Blob
		let contentBase64: string;
		let fileExtension: string;
		let mimeType: string;

		if (isBinary && value instanceof Blob) {
			// Handle binary content
			mimeType = value.type || 'application/octet-stream';

			// Read the Blob as base64
			contentBase64 = await new Promise<string>((resolve, reject) => {
				const reader = new FileReader();
				reader.onloadend = () => {
					const dataUrl = reader.result as string;
					const base64Data = dataUrl.split(',')[1]; // Extract base64 data
					resolve(base64Data);
				};
				reader.onerror = reject;
				reader.readAsDataURL(value);
			});

			fileExtension = mimeTypeMap[mimeType] || '';
		} else if (typeof value === 'string') {
			// Handle text content
			const languageKey = language.toLowerCase();
			const langInfo =
				languageKey in ProgrammingLanguages
					? ProgrammingLanguages[languageKey]
					: {
							extension: '.txt',
							mimeType: 'text/plain',
						};
			fileExtension = langInfo.extension;
			mimeType = langInfo.mimeType;

			// Convert string to base64
			contentBase64 = btoa(encodeURIComponent(value));
		} else {
			console.error('Unsupported content type for download');
			return;
		}

		const suggestedFileName = `${fileprefix}-${generateRandomString(3, true)}${fileExtension}`;

		// Prepare file filters
		const filters = [
			{
				DisplayName: `Files (*${fileExtension})`,
				Pattern: `*${fileExtension}`,
			},
			{
				DisplayName: 'All Files (*.*)',
				Pattern: '*.*',
			},
		];

		await saveFile(suggestedFileName, contentBase64, filters);
	};

	return (
		<button aria-label="Download" onClick={downloadAsFile} {...buttonProps}>
			<FiDownload size={size} />
		</button>
	);
};

export default DownloadButton;
