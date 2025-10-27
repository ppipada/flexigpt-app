// DownloadButton.tsx
import { type ButtonHTMLAttributes, useState } from 'react';

import { FiDownload } from 'react-icons/fi';

import type { FileFilter } from '@/spec/backend';

import { Base64EncodeUTF8, GenerateRandomString } from '@/lib/encode_decode';
import { MimeTypeMap, ProgrammingLanguages } from '@/lib/markdown_utils';

import { backendAPI } from '@/apis/baseapi';

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

export function DownloadButton({
	language = '',
	valueFetcher,
	size,
	fileprefix = 'file',
	isBinary = false,
	...buttonProps
}: DownloadButtonProps) {
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
				fileExtension = MimeTypeMap[mimeType] || '';
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

			const suggestedFileName = `${fileprefix}-${GenerateRandomString(3, true)}${fileExtension}`;
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
}
