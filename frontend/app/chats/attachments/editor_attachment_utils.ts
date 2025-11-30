import {
	type Attachment,
	type AttachmentFileRef,
	type AttachmentGenericRef,
	type AttachmentImageRef,
	AttachmentKind,
	AttachmentMode,
	type AttachmentURLRef,
} from '@/spec/attachment';
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

// EditorAttachment is the composer-side shape; it mirrors the backend
// Attachment but is limited to the fields we need in the UI.
export type EditorAttachment = {
	kind: AttachmentKind;
	label: string;

	mode: AttachmentMode;
	availableModes: AttachmentMode[];

	fileRef?: AttachmentFileRef;
	imageRef?: AttachmentImageRef;
	urlRef?: AttachmentURLRef;
	genericRef?: AttachmentGenericRef;

	// Error-only fields (never sent to backend as real attachments)
	isError?: boolean;
	errorReason?: AttachmentErrorReason;
};

// Convert editor attachment to API attachment payload.
export function editorAttachmentToConversation(att: EditorAttachment): Attachment {
	return {
		kind: att.kind,
		label: att.label,
		mode: att.mode,
		availableModes: att.availableModes,
		fileRef: att.fileRef,
		imageRef: att.imageRef,
		urlRef: att.urlRef,
		genericRef: att.genericRef,
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

function buildErrorAttachmentForLocalPath(att: Attachment, reason: AttachmentErrorReason): EditorAttachment {
	return {
		kind: AttachmentKind.file,
		label: att.label,
		mode: AttachmentMode.notReadable,
		availableModes: [AttachmentMode.notReadable],
		fileRef: att.fileRef,
		imageRef: att.imageRef,
		urlRef: att.urlRef,
		genericRef: att.genericRef,
		isError: true,
		errorReason: reason,
	};
}

/**
 * Classify a local file into an EditorAttachment with smart default mode
 * and allowed modes based on extension.
 */
export function buildEditorAttachmentForLocalPath(att: Attachment): EditorAttachment | undefined {
	if (!att.fileRef && !att.imageRef) {
		return undefined;
	}
	let sizeBytes = 0;
	if (att.fileRef && att.fileRef.size && att.fileRef.size > 0) {
		sizeBytes = att.fileRef.size;
	} else if (att.imageRef && att.imageRef.size && att.imageRef.size > 0) {
		sizeBytes = att.imageRef.size;
	}

	if (typeof sizeBytes === 'number' && sizeBytes > MAX_SINGLE_ATTACHMENT_BYTES) {
		return buildErrorAttachmentForLocalPath(att, AttachmentErrorReason.TooLargeSingle);
	}

	if (att.mode && att.mode == AttachmentMode.notReadable) {
		return buildErrorAttachmentForLocalPath(att, AttachmentErrorReason.Unreadable);
	}
	return {
		kind: att.kind,
		label: att.label,
		mode: att.mode ?? AttachmentMode.notReadable,
		availableModes: att.availableModes ?? [AttachmentMode.notReadable],
		fileRef: att.fileRef,
		imageRef: att.imageRef,
		urlRef: att.urlRef,
		genericRef: att.genericRef,
		isError: false,
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
	if (att.fileRef && att.fileRef.size && att.fileRef.size > 0) {
		size = att.fileRef.size;
	} else if (att.imageRef && att.imageRef.size && att.imageRef.size > 0) {
		size = att.imageRef.size;
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
