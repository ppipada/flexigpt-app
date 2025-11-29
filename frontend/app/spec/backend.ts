export interface FileFilter {
	DisplayName: string;
	Pattern: string;
}

export interface PathInfo {
	path: string;
	name: string;
	exists: boolean;
	isDir: boolean;
	size?: number;
	// Go type: time
	modTime?: Date;
}

/**
 * @public
 */
export interface DirectoryOverflowInfo {
	dirPath: string;
	relativePath: string;
	fileCount: number;
	partial: boolean;
}

export interface WalkDirectoryWithFilesResult {
	dirPath: string;
	files: PathInfo[];
	overflowDirs: DirectoryOverflowInfo[];
	maxFiles: number;
	totalSize: number;
	hasMore: boolean;
}

export interface IBackendAPI {
	ping: () => Promise<string>;
	log: (level: string, ...args: unknown[]) => void;
	saveFile(defaultFilename: string, contentBase64: string, filters: Array<FileFilter>): Promise<void>;
	openURL(url: string): void;
	openFiles(allowMultiple: boolean, filters: Array<FileFilter>): Promise<PathInfo[]>;
	openDirectoryWithFiles(maxFiles: number): Promise<WalkDirectoryWithFilesResult>;
}
