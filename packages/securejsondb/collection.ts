import { log } from 'logger';
import { existsSync, mkdirSync, readdirSync, unlinkSync } from 'node:fs';
import { join, resolve } from 'node:path';
import { SecureJSONFileDB } from './filedb';
import { SecureSchema } from './schema';

export class Collection<T extends SecureSchema> {
	protected readonly PAGE_SIZE = 25;
	protected baseDir: string;
	protected defaultData: T;

	constructor(baseDir: string, defaultData: T) {
		this.baseDir = resolve(baseDir);
		this.defaultData = defaultData;

		if (!existsSync(this.baseDir)) {
			mkdirSync(this.baseDir, { recursive: true });
			log.info(`Secure collection: Base directory created: ${this.baseDir}`);
		}
	}

	private getFullPath(filename: string): string {
		return join(this.baseDir, filename);
	}

	protected async addFileByPath(filePath: string, data: T): Promise<void> {
		try {
			const secureDB = new SecureJSONFileDB<T>(filePath, this.defaultData);
			await secureDB.initialize();
			await secureDB.overwriteData(data);
			log.debug(`Secure collection: File added: ${filePath}`);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : String(error);
			log.error(`Secure collection: Failed to add file: ${errorMessage}`);
			throw error;
		}
	}

	protected async getFileByPath(filePath: string): Promise<T | null> {
		if (!existsSync(filePath)) {
			log.debug(`Secure collection: File not found: ${filePath}`);
			return null;
		}

		try {
			const secureDB = new SecureJSONFileDB<T>(filePath, this.defaultData);
			await secureDB.initialize();
			return await secureDB.getAllData();
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : String(error);
			log.error(`Secure collection: Failed to get file: ${errorMessage}`);
			throw error;
		}
	}

	protected listFilesByPath(dirPath: string): string[] {
		if (!existsSync(dirPath)) {
			throw new Error(`Secure collection: Directory not found ${dirPath}`);
		}
		const allFiles = readdirSync(dirPath);
		const allFilesFullPaths: string[] = [];
		allFiles.forEach(file => {
			allFilesFullPaths.push(join(dirPath, file));
		});
		return allFilesFullPaths;
	}

	protected async deleteFileByPath(filePath: string): Promise<void> {
		if (!existsSync(filePath)) {
			log.debug(`Secure collection: File to delete not found: ${filePath}`);
			return;
		}

		try {
			unlinkSync(filePath);
			log.info(`Secure collection: File deleted: ${filePath}`);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : String(error);
			log.error(`Secure collection: Failed to delete file: ${errorMessage}`);
			throw error;
		}
	}

	async addFile(filename: string, data: T): Promise<void> {
		const filePath = this.getFullPath(filename);
		await this.addFileByPath(filePath, data);
	}

	async deleteFile(filename: string): Promise<void> {
		const filePath = this.getFullPath(filename);
		await this.deleteFileByPath(filePath);
	}

	async getFile(filename: string): Promise<T | null> {
		const filePath = this.getFullPath(filename);
		return this.getFileByPath(filePath);
	}

	async listFiles(): Promise<{ files: string[]; nextToken?: string }> {
		const files = this.listFilesByPath(this.baseDir);
		return { files: files, nextToken: undefined };
	}
}

export class CollectionMonthPartitioned<T extends SecureSchema> extends Collection<T> {
	constructor(baseDir: string, defaultData: T) {
		super(baseDir, defaultData);
	}

	private getCurrentPartitionPath(): string {
		const now = new Date();
		const year = now.getFullYear().toString();
		const month = (now.getMonth() + 1).toString().padStart(2, '0');
		return join(this.baseDir, year, month);
	}

	private getFileCurrentPartition(filename: string, mkdirs = true): string {
		const partitionPath = this.getCurrentPartitionPath();
		if (!existsSync(partitionPath)) {
			if (mkdirs) {
				mkdirSync(partitionPath, { recursive: true });
			} else {
				throw new Error(`Secure collection: Partition not found ${partitionPath}`);
			}
		}
		return join(partitionPath, filename);
	}

	private listPartitions(): string[] {
		const partitions: string[] = [];
		const years = readdirSync(this.baseDir).sort().reverse();
		years.forEach(year => {
			const yearPath = join(this.baseDir, year);
			const months = readdirSync(yearPath).sort().reverse();
			months.forEach(month => {
				partitions.push(join(year, month));
			});
		});
		return partitions.sort().reverse();
	}

	private parseToken(token?: string): { partitionIndex: number } {
		if (!token) return { partitionIndex: 0 };
		const partitionIndex = Number(token);
		return { partitionIndex };
	}

	private createToken(partitionIndex: number): string {
		return partitionIndex.toString();
	}

	async addFile(filename: string, data: T, partitionPath?: string): Promise<void> {
		let filePath = '';
		if (!partitionPath) {
			filePath = this.getFileCurrentPartition(filename);
		} else {
			filePath = join(this.baseDir, partitionPath, filename);
		}

		await this.addFileByPath(filePath, data);
	}

	async deleteFile(filename: string, partitionPath?: string): Promise<void> {
		let filePath = '';
		if (!partitionPath) {
			filePath = this.getFileCurrentPartition(filename);
		} else {
			filePath = join(this.baseDir, partitionPath, filename);
		}

		await this.deleteFileByPath(filePath);
	}

	async getFile(filename: string, partitionPath?: string): Promise<T | null> {
		let filePath = '';
		if (!partitionPath) {
			filePath = this.getFileCurrentPartition(filename);
		} else {
			filePath = join(this.baseDir, partitionPath, filename);
		}
		return this.getFileByPath(filePath);
	}

	async listFiles(token?: string, partitionPath?: string): Promise<{ files: string[]; nextToken?: string }> {
		if (partitionPath) {
			const files = this.listFilesByPath(join(this.baseDir, partitionPath));
			return { files: files, nextToken: undefined };
		}
		const partitions = this.listPartitions();
		const { partitionIndex } = this.parseToken(token);
		let nextToken: string | undefined = undefined;
		const allFiles: string[] = [];

		for (let i = partitionIndex; i < partitions.length; i++) {
			const partPath = join(this.baseDir, partitions[i]);
			const partFiles = this.listFilesByPath(partPath);

			allFiles.push(...partFiles);

			if (allFiles.length >= this.PAGE_SIZE) {
				nextToken = this.createToken(i + 1);
				return { files: allFiles, nextToken };
			}
		}

		return { files: allFiles, nextToken };
	}
}
