import type { Attachment, FileFilter } from '@/spec/attachment';

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

export interface IBackendAPI {
	ping: () => Promise<string>;
	log: (level: string, ...args: unknown[]) => void;
	openURL(url: string): void;
	openURLAsAttachment(rawURL: string): Promise<Attachment | undefined>;
	saveFile(defaultFilename: string, contentBase64: string, additionalFilters?: Array<FileFilter>): Promise<void>;
	openMultipleFilesAsAttachments(allowMultiple: boolean, additionalFilters?: Array<FileFilter>): Promise<Attachment[]>;
	openDirectoryAsAttachments(maxFiles: number): Promise<DirectoryAttachmentsResult>;
}
