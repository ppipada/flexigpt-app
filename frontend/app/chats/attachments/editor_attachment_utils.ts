import { type Attachment, AttachmentKind } from '@/spec/attachment';

interface EditorFileAttachmentRef {
	path: string;
}

interface EditorImageAttachmentRef {
	path: string;
}

export interface EditorAttachment {
	kind: AttachmentKind.file | AttachmentKind.image;
	label: string;
	fileRef?: EditorFileAttachmentRef;
	imageRef?: EditorImageAttachmentRef;
}

export function editorAttachmentToConversation(att: EditorAttachment): Attachment {
	return {
		kind: att.kind,
		label: att.label,
		fileRef: att.fileRef ? { path: att.fileRef.path } : undefined,
		imageRef: att.imageRef ? { path: att.imageRef.path } : undefined,
	};
}

export function editorAttachmentKey(att: EditorAttachment): string {
	const p = getEditorAttachmentPath(att);
	return `${att.kind}:${p}`;
}

export function getEditorAttachmentPath(att: EditorAttachment): string {
	let p = '';
	if (att.kind === AttachmentKind.file && att.fileRef !== undefined) {
		p = att.fileRef.path;
	}
	if (att.kind === AttachmentKind.image && att.imageRef !== undefined) {
		p = att.imageRef.path;
	}
	if (p === '') {
		throw Error('invalid editor attachment');
	}
	return p;
}
