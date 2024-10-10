export interface IBackendAPI {
	ping: () => string;
	log: (level: string, ...args: unknown[]) => void;
}
