import axios, {
	AxiosError,
	AxiosInstance,
	AxiosRequestConfig,
	AxiosResponse,
	InternalAxiosRequestConfig,
	ResponseType,
} from 'axios';
import { log } from 'logger';

export function filterSensitiveInfo(obj: any): any {
	const sensitiveKeys = ['authorization', 'key'];

	if (typeof obj !== 'object' || obj === null) {
		return obj;
	}

	if (Array.isArray(obj)) {
		return obj.map(item => filterSensitiveInfo(item));
	}

	const filteredObj: any = {};
	for (const key in obj) {
		if (!sensitiveKeys.some(sensitiveKey => key.toLowerCase().includes(sensitiveKey))) {
			filteredObj[key] = filterSensitiveInfo(obj[key]);
		}
	}
	return filteredObj;
}

export function filterSensitiveInfoFromJsonString(jsonString: string): string {
	const jsonObj = JSON.parse(jsonString);
	const filteredObj = filterSensitiveInfo(jsonObj);
	return JSON.stringify(filteredObj);
}

export class APICaller {
	origin: string;
	apiKey: string;
	apiKeyHeaderKey: string;
	timeout: number;
	headers: Record<string, string>;
	logRequests = false;
	private axiosInstance: AxiosInstance;

	constructor(
		origin: string,
		apiKey: string,
		apiKeyHeaderKey: string,
		timeout: number,
		headers: Record<string, string> = {}
	) {
		this.origin = origin;
		this.apiKeyHeaderKey = apiKeyHeaderKey;
		this.apiKey = apiKey;
		this.timeout = timeout;
		this.headers = headers;

		this.axiosInstance = axios.create({ adapter: 'fetch' });

		if (this.logRequests) {
			this.axiosInstance.interceptors.request.use(
				(config: InternalAxiosRequestConfig) => {
					log.debug('cURL Command:', this.generateCurlCommand(config));
					return config;
				},
				error => {
					throw error;
				}
			);
		}
	}

	generateCurlCommand(config: AxiosRequestConfig): string {
		let curlCommand = 'curl -X ' + config.method?.toUpperCase() + ' ';
		curlCommand += '"' + config.url + '" ';

		if (config.headers) {
			Object.keys(config.headers).forEach(key => {
				if (key != this.apiKeyHeaderKey) {
					const value = config.headers?.[key];
					curlCommand += '-H "' + key + ': ' + value + '" ';
				}
			});
		}

		if (config.data) {
			curlCommand += "-d '" + JSON.stringify(config.data) + "' ";
		}

		return curlCommand;
	}

	private extendAxiosRequestConfig(requestConfig: AxiosRequestConfig, stream = false): AxiosRequestConfig {
		let apiKey = this.apiKey;
		if (this.apiKeyHeaderKey === 'Authorization') {
			apiKey = 'Bearer ' + apiKey;
		}
		const mergedHeaders = {
			...(this.apiKeyHeaderKey ? { [this.apiKeyHeaderKey]: apiKey } : {}),
			...this.headers,
			...(requestConfig.headers || {}),
		};
		const respType: ResponseType = stream ? 'stream' : 'json';
		const config: AxiosRequestConfig = {
			...requestConfig,
			url: this.origin + (requestConfig.url || ''),
			headers: mergedHeaders,
			responseType: respType,
		};
		return config;
	}

	private handleError(error: unknown) {
		if (axios.isAxiosError(error)) {
			const axiosError = error as AxiosError;
			let errorData: string;
			if (axiosError.response) {
				const headers = filterSensitiveInfo(axiosError.response.headers);
				errorData =
					JSON.stringify(axiosError.response.data, null, 2) +
					'\n' +
					JSON.stringify(axiosError.response.status, null, 2) +
					'\n' +
					JSON.stringify(headers, null, 2) +
					'\n';
			} else {
				errorData = JSON.stringify(axiosError, null, 2) + '\n';
			}
			error.message = errorData + '\n' + error.message;
		}
		return error;
	}

	async request<T>(requestConfig: AxiosRequestConfig): Promise<T> {
		const config = this.extendAxiosRequestConfig(requestConfig);

		try {
			const response: AxiosResponse<T> = await this.axiosInstance.request(config);
			return response.data;
		} catch (error) {
			const newError = this.handleError(error);
			throw newError;
		}
	}

	async requestStream(
		requestConfig: AxiosRequestConfig,
		dataChunkProcessor: (data: string) => Promise<void>
	): Promise<void> {
		const config = this.extendAxiosRequestConfig(requestConfig, true);
		const response = await this.axiosInstance.request(config);
		const stream = response.data as ReadableStream<Uint8Array>;
		const reader = stream.getReader();
		const decoder = new TextDecoder();
		const processText = async ({ done, value }: ReadableStreamReadResult<Uint8Array>) => {
			if (done) {
				await dataChunkProcessor('data: [DONE]');
				return;
			}

			const dataString = decoder.decode(value, { stream: true });
			await dataChunkProcessor(dataString);

			reader.read().then(processText);
		};
		reader.read().then(processText);
	}
}
