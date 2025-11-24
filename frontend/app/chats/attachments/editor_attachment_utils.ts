import { type ConversationAttachment, ConversationAttachmentKind } from '@/spec/conversation';

interface EditorFileAttachmentRef {
	path: string;
}

interface EditorImageAttachmentRef {
	path: string;
}

export interface EditorAttachment {
	kind: ConversationAttachmentKind.file | ConversationAttachmentKind.image;
	label: string;
	fileRef?: EditorFileAttachmentRef;
	imageRef?: EditorImageAttachmentRef;
}

export function editorAttachmentToConversation(att: EditorAttachment): ConversationAttachment {
	return {
		kind: att.kind,
		label: att.label,
		imageRef: att.imageRef ? { path: att.imageRef.path } : undefined,
		fileRef: att.fileRef ? { path: att.fileRef.path } : undefined,
	};
}

export function editorAttachmentKey(att: EditorAttachment): string {
	const p = getEditorAttachmentPath(att);
	return `${att.kind}:${p}`;
}

export function getEditorAttachmentPath(att: EditorAttachment): string {
	let p = '';
	if (att.kind === ConversationAttachmentKind.file && att.fileRef !== undefined) {
		p = att.fileRef.path;
	}
	if (att.kind === ConversationAttachmentKind.image && att.imageRef !== undefined) {
		p = att.imageRef.path;
	}
	if (p === '') {
		throw Error('invalid editor attachment');
	}
	return p;
}
