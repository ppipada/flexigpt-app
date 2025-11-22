export interface FileFilter {
	DisplayName: string;
	Pattern: string;
}

export interface IBackendAPI {
	ping: () => Promise<string>;
	log: (level: string, ...args: unknown[]) => void;
	savefile(defaultFilename: string, contentBase64: string, filters: Array<FileFilter>): Promise<void>;
	openurl(url: string): void;
	openfiles(allowMultiple: boolean, filters: Array<FileFilter>): Promise<string[]>;
}
