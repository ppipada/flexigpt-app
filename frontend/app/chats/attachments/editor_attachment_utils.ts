import { type Attachment, AttachmentKind, AttachmentMode } from '@/spec/attachment';
import type { DirectoryOverflowInfo } from '@/spec/backend';

/**
 * @public
 */
export const MAX_SINGLE_ATTACHMENT_BYTES = 16 * 1024 * 1024; // 16 MiB
export const MAX_FILES_PER_DIRECTORY = 64;

// Directory grouping is UI-only.
export interface DirectoryAttachmentGroup {
	id: string;
	dirPath: string;
	label: string;
	/**
	 * All attachment keys (editorAttachmentKey) that belong to this folder.
	 * Includes both "owned" and "referenced" attachments.
	 */
	attachmentKeys: string[];
	/**
	 * Subset of attachmentKeys created specifically for this folder selection.
	 * Used to decide which attachments to remove when the folder is removed.
	 */
	ownedAttachmentKeys: string[];
	/**
	 * Subdirectories that were not fully walked due to maxFiles limits, etc.
	 */
	overflowDirs: DirectoryOverflowInfo[];
}

/**
 * @public
 */
export enum AttachmentErrorReason {
	TooLargeSingle = 'too-large-single',
	TooLargeTotal = 'too-large-total',
	Unreadable = 'unreadable',
}

interface EditorFileAttachmentRef {
	path: string;
	sizeBytes?: number;
}

interface EditorImageAttachmentRef {
	path: string;
	sizeBytes?: number;
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

	// Error-only fields (never sent to backend as real attachments)
	isError?: boolean;
	errorReason?: AttachmentErrorReason;
}

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
	let label = att.label;
	if (att.kind === AttachmentKind.file && att.fileRef) {
		label = att.fileRef.path;
	}
	if (att.kind === AttachmentKind.image && att.imageRef) {
		label = att.imageRef.path;
	}
	if (att.kind === AttachmentKind.url && att.urlRef) {
		label = att.urlRef.url;
	}
	return `${att.kind}:${label}`;
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

function buildErrorAttachmentForLocalPath(
	path: string,
	sizeBytes: number | undefined,
	reason: AttachmentErrorReason
): EditorAttachment {
	const trimmed = path.trim();
	const label = trimmed.split(/[\\/]/).pop() ?? trimmed;

	// Unknown/binary: mark as not readable, no dropdown.
	return {
		kind: AttachmentKind.file,
		label,
		mode: AttachmentMode.notReadable,
		availableModes: [AttachmentMode.notReadable],
		fileRef: { path: trimmed, sizeBytes: sizeBytes },
		isError: true,
		errorReason: reason,
	};
}

/**
 * Classify a local file into an EditorAttachment with smart default mode
 * and allowed modes based on extension.
 */
export function buildEditorAttachmentForLocalPath(path: string, sizeBytes?: number): EditorAttachment {
	const trimmed = path.trim();

	if (typeof sizeBytes === 'number' && sizeBytes > MAX_SINGLE_ATTACHMENT_BYTES) {
		return buildErrorAttachmentForLocalPath(trimmed, sizeBytes, AttachmentErrorReason.TooLargeSingle);
	}

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
			imageRef: { path: trimmed, sizeBytes: sizeBytes },
		};
	}

	if (textLikeExts.has(ext)) {
		return {
			kind: AttachmentKind.file,
			label,
			mode: AttachmentMode.text,
			availableModes: [AttachmentMode.text],
			fileRef: { path: trimmed, sizeBytes: sizeBytes },
		};
	}

	if (docExts.has(ext) && ext === 'pdf') {
		// We dont support  extractors other than pdf as of now.
		return {
			kind: AttachmentKind.file,
			label,
			mode: AttachmentMode.file,
			availableModes: [AttachmentMode.text, AttachmentMode.file],
			fileRef: { path: trimmed, sizeBytes: sizeBytes },
		};
	}

	// Unknown/binary: mark as not readable, no dropdown.
	return buildErrorAttachmentForLocalPath(path, sizeBytes, AttachmentErrorReason.Unreadable);
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
			mode: AttachmentMode.file,
			availableModes: [AttachmentMode.text, AttachmentMode.file, AttachmentMode.link],
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

// Human-readable explanation for an error attachment, for tooltips.
export function getAttachmentErrorMessage(att: EditorAttachment): string | null {
	if (!att.isError || !att.errorReason) return null;
	const limit = formatBytes(MAX_SINGLE_ATTACHMENT_BYTES);
	let size = 0;
	if (att.fileRef && att.fileRef.sizeBytes && att.fileRef.sizeBytes > 0) {
		size = att.fileRef.sizeBytes;
	} else if (att.imageRef && att.imageRef.sizeBytes && att.imageRef.sizeBytes > 0) {
		size = att.imageRef.sizeBytes;
	}
	if (size <= 0) {
		return null;
	}
	switch (att.errorReason) {
		case AttachmentErrorReason.TooLargeSingle: {
			return `Too large to attach (${formatBytes(size)}; limit is ${limit}).`;
		}
		case AttachmentErrorReason.TooLargeTotal:
			return 'Too many or too large files in this folder were skipped; only a subset was attached.';
		case AttachmentErrorReason.Unreadable:
		default:
			return 'This file type is not supported or could not be read.';
	}
}

function formatBytes(bytes: number): string {
	if (!Number.isFinite(bytes) || bytes < 0) return `${bytes} B`;
	const units = ['B', 'KB', 'MB', 'GB', 'TB'];
	let i = 0;
	let v = bytes;
	while (v >= 1024 && i < units.length - 1) {
		v /= 1024;
		i++;
	}
	return `${v.toFixed(v >= 10 || i === 0 ? 0 : 1)} ${units[i]}`;
}
