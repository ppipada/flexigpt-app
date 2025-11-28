import { type Attachment, AttachmentKind, AttachmentMode } from '@/spec/attachment';

interface EditorFileAttachmentRef {
	path: string;
}

interface EditorImageAttachmentRef {
	path: string;
}

interface EditorURLAttachmentRef {
	url: string;
}

// EditorAttachment is the composer-side shape; it mirrors the backend
// Attachment but is limited to the fields we need in the UI.
export interface EditorAttachment {
	kind: AttachmentKind;
	label: string;

	mode: AttachmentMode;
	availableModes: AttachmentMode[];

	fileRef?: EditorFileAttachmentRef;
	imageRef?: EditorImageAttachmentRef;
	urlRef?: EditorURLAttachmentRef;
}

// Mapping from mode value → user-facing label (for the mode dropdown/badge).
export const ATTACHMENT_MODE_LABELS: Record<AttachmentMode, string> = {
	[AttachmentMode.text]: 'Text',
	[AttachmentMode.file]: 'File',
	[AttachmentMode.image]: 'Image',
	[AttachmentMode.page]: 'Page content',
	[AttachmentMode.link]: 'Link only',
	[AttachmentMode.pdfFile]: 'PDF file',
	[AttachmentMode.notReadable]: 'Not readable',
	[AttachmentMode.prDiff]: 'PR diff only',
	[AttachmentMode.prPage]: 'PR page',
	[AttachmentMode.commitDiff]: 'Diff only',
	[AttachmentMode.commitPage]: 'Commit page',
};

export const ATTACHMENT_MODE_DESC: Record<AttachmentMode, string> = {
	[AttachmentMode.text]: 'text inside message; may need processing for some file types',
	[AttachmentMode.file]: 'base64 encoded file; needs API support',
	[AttachmentMode.image]: 'base64 encoded image; needs API support',
	[AttachmentMode.page]: 'text extracted from html',
	[AttachmentMode.link]: 'link only, no processing',
	[AttachmentMode.pdfFile]: 'base64 encoded pdf; needs API support',
	[AttachmentMode.notReadable]: 'unreadable file',
	[AttachmentMode.prDiff]: 'pr diff only',
	[AttachmentMode.prPage]: 'pre page',
	[AttachmentMode.commitDiff]: 'diff only',
	[AttachmentMode.commitPage]: 'commit page',
};
// Convert editor attachment to API attachment payload.
export function editorAttachmentToConversation(att: EditorAttachment): Attachment {
	return {
		kind: att.kind,
		label: att.label,
		mode: att.mode,
		availableModes: att.availableModes,
		fileRef: att.fileRef ? { path: att.fileRef.path } : undefined,
		imageRef: att.imageRef ? { path: att.imageRef.path } : undefined,
		urlRef: att.urlRef ? { url: att.urlRef.url } : undefined,
	};
}

// Identity key used for de-duplication in the composer.
export function editorAttachmentKey(att: EditorAttachment): string {
	const id = getEditorAttachmentIdentity(att);
	return `${att.kind}:${id}`;
}

function getEditorAttachmentIdentity(att: EditorAttachment): string {
	if (att.kind === AttachmentKind.file && att.fileRef) {
		return att.fileRef.path;
	}
	if (att.kind === AttachmentKind.image && att.imageRef) {
		return att.imageRef.path;
	}
	if (att.kind === AttachmentKind.url && att.urlRef) {
		return att.urlRef.url;
	}
	return att.label;
}

// Used only for tooltip/debug display in the chips bar.
export function getEditorAttachmentPath(att: EditorAttachment): string {
	if (att.kind === AttachmentKind.file && att.fileRef) {
		return att.fileRef.path;
	}
	if (att.kind === AttachmentKind.image && att.imageRef) {
		return att.imageRef.path;
	}
	if (att.kind === AttachmentKind.url && att.urlRef) {
		return att.urlRef.url;
	}
	return '';
}

/**
 * Classify a local file into an EditorAttachment with smart default mode
 * and allowed modes based on extension.
 */
export function buildEditorAttachmentForLocalPath(path: string): EditorAttachment {
	const trimmed = path.trim();
	const label = trimmed.split(/[\\/]/).pop() ?? trimmed;
	const lower = trimmed.toLowerCase();
	const ext = lower.split('.').pop() ?? '';

	const imageExts = new Set(['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'svg']);
	const textLikeExts = new Set([
		'txt',
		'md',
		'markdown',
		'log',
		'json',
		'yaml',
		'yml',
		'toml',
		'js',
		'ts',
		'tsx',
		'jsx',
		'py',
		'go',
		'rs',
		'java',
		'c',
		'cpp',
		'h',
		'hpp',
		'cs',
		'rb',
		'php',
		'html',
		'css',
		'scss',
		'less',
		'sql',
	]);
	const docExts = new Set(['pdf', 'doc', 'docx', 'ppt', 'pptx', 'xls', 'xlsx', 'odt', 'ods']);

	if (imageExts.has(ext)) {
		return {
			kind: AttachmentKind.image,
			label,
			mode: AttachmentMode.image,
			availableModes: [AttachmentMode.image],
			imageRef: { path: trimmed },
		};
	}

	if (textLikeExts.has(ext)) {
		return {
			kind: AttachmentKind.file,
			label,
			mode: AttachmentMode.text,
			availableModes: [AttachmentMode.text],
			fileRef: { path: trimmed },
		};
	}

	if (docExts.has(ext)) {
		// UX: default to text content, but allow user to switch to file mode.
		return {
			kind: AttachmentKind.file,
			label,
			mode: AttachmentMode.file,
			availableModes: [AttachmentMode.text, AttachmentMode.file],
			fileRef: { path: trimmed },
		};
	}

	// Unknown/binary: mark as not readable, no dropdown.
	return {
		kind: AttachmentKind.file,
		label,
		mode: AttachmentMode.notReadable,
		availableModes: [AttachmentMode.notReadable],
		fileRef: { path: trimmed },
	};
}

/**
 * Build a URL-based attachment with smart default modes.
 */
export function buildEditorAttachmentForURL(rawUrl: string): EditorAttachment {
	const trimmed = rawUrl.trim();
	const label = trimmed;
	let url: URL | null = null;
	try {
		url = new URL(trimmed);
	} catch {
		// Keep as raw string; backend will validate.
	}

	const pathname = url?.pathname.toLowerCase() ?? '';
	const ext = pathname.split('.').pop() ?? '';

	const imageExts = new Set(['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp', 'svg']);
	if (imageExts.has(ext)) {
		return {
			kind: AttachmentKind.url,
			label,
			mode: AttachmentMode.image,
			availableModes: [AttachmentMode.image, AttachmentMode.link],
			urlRef: { url: trimmed },
		};
	}

	if (ext === 'pdf') {
		return {
			kind: AttachmentKind.url,
			label,
			mode: AttachmentMode.text,
			availableModes: [AttachmentMode.text, AttachmentMode.pdfFile, AttachmentMode.link],
			urlRef: { url: trimmed },
		};
	}

	// Default: treat as HTML web page.
	return {
		kind: AttachmentKind.url,
		label,
		mode: AttachmentMode.page,
		availableModes: [AttachmentMode.page, AttachmentMode.link],
		urlRef: { url: trimmed },
	};
}
