// DownloadButton.tsx
import { type ButtonHTMLAttributes, type FC, useState } from 'react';

import { FiDownload } from 'react-icons/fi';

import type { FileFilter } from '@/models/backendmodel';

import { backendAPI } from '@/apis/baseapi';

import { Base64EncodeUTF8 } from '@/lib/encode_decode';

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
	const [downloading, setDownloading] = useState(false);

	const downloadAsFile = async () => {
		setDownloading(true);
		try {
			const value = await valueFetcher();
			let contentBase64: string;
			let fileExtension: string;
			let mimeType: string;

			if (isBinary && value instanceof Blob) {
				mimeType = value.type || 'application/octet-stream';
				contentBase64 = await new Promise<string>((resolve, reject) => {
					const reader = new FileReader();
					reader.onloadend = () => {
						const dataUrl = reader.result as string;
						const base64Data = dataUrl.split(',')[1];
						resolve(base64Data);
					};
					reader.onerror = reject;
					reader.readAsDataURL(value);
				});
				fileExtension = mimeTypeMap[mimeType] || '';
			} else if (typeof value === 'string') {
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
				contentBase64 = Base64EncodeUTF8(value);
			} else {
				throw new Error('Unsupported content type for download');
			}

			const suggestedFileName = `${fileprefix}-${generateRandomString(3, true)}${fileExtension}`;
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
		} catch (err: any) {
			console.error('Download failed:', err, JSON.stringify(err));
		} finally {
			setDownloading(false);
		}
	};

	return (
		<button aria-label="Download" title="Download" onClick={downloadAsFile} disabled={downloading} {...buttonProps}>
			<FiDownload size={size} />
		</button>
	);
};

export default DownloadButton;
