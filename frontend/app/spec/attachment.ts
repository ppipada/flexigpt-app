// AttachmentKind enumerates contextual attachment categories that can be
// associated with messages sent to the inference layer.
export enum AttachmentKind {
	file = 'file',
	image = 'image',
	docIndex = 'docIndex',
	pr = 'pr',
	commit = 'commit',
	snapshot = 'snapshot',
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

export interface AttachmentGenericRef {
	handle: string;
}

// Attachment references contextual artifacts (files, images, doc handles, etc.).
export interface Attachment {
	kind: AttachmentKind;
	label: string;
	fileRef?: AttachmentFileRef;
	imageRef?: AttachmentImageRef;
	genericRef?: AttachmentGenericRef;
}
