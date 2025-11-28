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
export interface AttachmentFileRef {
	path: string;
	exists?: boolean;
	sizeBytes?: number;
	modTime?: Date;
}

export interface AttachmentImageRef {
	path: string;
	exists?: boolean;
	width?: number;
	height?: number;
	format?: string;
	sizeBytes?: number;
	modTime?: Date;
}

export interface AttachmentURLRef {
	url: string;
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
