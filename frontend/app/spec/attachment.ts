// AttachmentKind enumerates contextual attachment categories that can be
// associated with messages sent to the inference layer.
export enum AttachmentKind {
	file = 'file',
	image = 'image',
	url = 'url',
	docIndex = 'docIndex',
	pr = 'pr',
	commit = 'commit',
}

export enum AttachmentMode {
	text = 'text',
	file = 'file',
	image = 'image',
	page = 'page',
	link = 'link',
	notReadable = 'not-readable',

	prDiff = 'pr-diff',
	prPage = 'pr-page',
	commitDiff = 'commit-diff',
	commitPage = 'commit-page',
}

// Mapping from mode value → user-facing label (for the mode dropdown/badge).
export const ATTACHMENT_MODE_LABELS: Record<AttachmentMode, string> = {
	[AttachmentMode.text]: 'Text',
	[AttachmentMode.file]: 'File',
	[AttachmentMode.image]: 'Image',
	[AttachmentMode.page]: 'Page content',
	[AttachmentMode.link]: 'Link only',
	[AttachmentMode.notReadable]: 'Not readable',
	[AttachmentMode.prDiff]: 'PR diff only',
	[AttachmentMode.prPage]: 'PR page',
	[AttachmentMode.commitDiff]: 'Diff only',
	[AttachmentMode.commitPage]: 'Commit page',
};

export const ATTACHMENT_MODE_DESC: Record<AttachmentMode, string> = {
	[AttachmentMode.text]: 'Inline text extracted into the message.',
	[AttachmentMode.file]: 'Send the file as a binary attachment (requires API support).',
	[AttachmentMode.image]: 'Send the image as a binary attachment (requires API support).',
	[AttachmentMode.page]: 'Text content extracted from the HTML page.',
	[AttachmentMode.link]: 'Only send the link; do not fetch or parse content.',
	[AttachmentMode.notReadable]: 'This file could not be read.',
	[AttachmentMode.prDiff]: 'Only the pull request diff is sent.',
	[AttachmentMode.prPage]: 'Send the pull request page content.',
	[AttachmentMode.commitDiff]: 'Only the commit diff is sent.',
	[AttachmentMode.commitPage]: 'Send the commit page content.',
};

export interface AttachmentFileRef {
	path: string;
	name: string;
	exists: boolean;
	isDir: boolean;
	size?: number;
	// Go type: time
	modTime?: Date;
}

export interface AttachmentImageRef {
	path: string;
	name: string;
	exists: boolean;
	isDir: boolean;
	size?: number;
	// Go type: time
	modTime?: Date;

	width?: number;
	height?: number;
	format?: string;
	mimeType?: string;
}

export interface AttachmentURLRef {
	url: string;
	normalized?: string;
}

export interface AttachmentGenericRef {
	handle: string;
}

// Attachment references contextual artifacts (files, images, doc handles, etc.).
export interface Attachment {
	kind: AttachmentKind;
	label: string;

	mode?: AttachmentMode;
	availableModes?: AttachmentMode[];

	fileRef?: AttachmentFileRef;
	imageRef?: AttachmentImageRef;
	urlRef?: AttachmentURLRef;
	genericRef?: AttachmentGenericRef;
}

export interface FileFilter {
	DisplayName: string;
	Extensions: string[];
}
