import { programmingLanguages } from '@/models/conversationmodel';
import { ButtonHTMLAttributes, FC } from 'react';
import { FiDownload } from 'react-icons/fi';

function generateRandomString(length: number, lowercase = false): string {
	const chars = 'ABCDEFGHJKLMNPQRSTUVWXY3456789'; // excluding similar looking characters like Z, 2, I, 1, O, 0
	let result = '';
	for (let i = 0; i < length; i++) {
		result += chars.charAt(Math.floor(Math.random() * chars.length));
	}
	return lowercase ? result.toLowerCase() : result;
}

export interface DownloadButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	language: string;
	valueFetcher: () => Promise<string>;
	size: number;
	fileprefix?: string;
}

const DownloadButton: FC<DownloadButtonProps> = ({
	language,
	valueFetcher,
	size,
	fileprefix = 'file',
	...buttonProps
}) => {
	const downloadAsFile = async () => {
		if (typeof window === 'undefined') {
			return;
		}

		const value = await valueFetcher();

		const fileExtension = programmingLanguages[language] || '.txt';
		const suggestedFileName = `${fileprefix}-${generateRandomString(3, true)}${fileExtension}`;

		const blob = new Blob([value], { type: 'text/plain' });
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.download = suggestedFileName;
		link.href = url;
		link.style.display = 'none';
		document.body.appendChild(link);
		link.click();
		document.body.removeChild(link);
		URL.revokeObjectURL(url);
	};

	return (
		<button aria-label="Download" onClick={downloadAsFile} {...buttonProps}>
			<FiDownload size={size} />
		</button>
	);
};

export default DownloadButton;
