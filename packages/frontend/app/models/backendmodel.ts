export interface IBackendAPI {
	ping: () => Promise<string>;
	log: (level: string, ...args: unknown[]) => void;
}
