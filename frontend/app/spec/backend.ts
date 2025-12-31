import type { Attachment } from '@/spec/attachment';

/**
 * @public
 */
export interface DirectoryOverflowInfo {
	dirPath: string;
	relativePath: string;
	fileCount: number;
	partial: boolean;
}

export interface DirectoryAttachmentsResult {
	dirPath: string;
	attachments: Attachment[];
	overflowDirs: DirectoryOverflowInfo[];
	maxFiles: number;
	totalSize: number;
	hasMore: boolean;
}
