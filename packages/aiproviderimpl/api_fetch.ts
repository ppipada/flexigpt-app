import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, ResponseType } from 'axios';

import { APIErrorDetails, APIFetchResponse, APIResponseDetails } from 'aiprovidermodel';
import { setupInterceptors } from './api_fetch_interceptors';

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
		setupInterceptors(this.axiosInstance, this.logRequests);
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

	async request<T>(requestConfig: AxiosRequestConfig): Promise<APIFetchResponse<T>> {
		const config = this.extendAxiosRequestConfig(requestConfig);
		const resp: APIFetchResponse<T> = {};
		try {
			const response: AxiosResponse<T> = await this.axiosInstance.request(config);
			const responseDetails = (response.config as any).requestDetails as APIResponseDetails;
			resp.data = response.data;
			resp.requestDetails = responseDetails.requestDetails;
			resp.responseDetails = responseDetails;
			resp.responseDetails.requestDetails = undefined;
		} catch (error) {
			if (axios.isAxiosError(error)) {
				const errorDetails = (error as any).errorDetails as APIErrorDetails;
				resp.errorDetails = errorDetails;
			} else {
				resp.errorDetails = { message: JSON.stringify(error, null, 2) };
			}
		}
		return resp;
	}

	async requestStream<T>(
		requestConfig: AxiosRequestConfig,
		dataChunkProcessor: (data: string, details?: APIFetchResponse<T>) => Promise<void>
	): Promise<void> {
		const config = this.extendAxiosRequestConfig(requestConfig, true);
		const resp: APIFetchResponse<T> = {};
		try {
			const response = await this.axiosInstance.request(config);
			const responseDetails = (response.config as any).responseDetails as APIResponseDetails;
			resp.requestDetails = responseDetails.requestDetails;
			resp.responseDetails = responseDetails;
			resp.responseDetails.data = undefined;
			resp.responseDetails.requestDetails = undefined;
			const stream = response.data as ReadableStream<Uint8Array>;
			const reader = stream.getReader();
			const decoder = new TextDecoder();

			const processText = async ({ done, value }: ReadableStreamReadResult<Uint8Array>) => {
				if (done) {
					await dataChunkProcessor('data: [DONE]', resp);
					return;
				}

				const dataString = decoder.decode(value, { stream: true });
				await dataChunkProcessor(dataString);

				reader.read().then(processText);
			};
			reader.read().then(processText);
		} catch (error) {
			const errorDetails: APIErrorDetails = axios.isAxiosError(error)
				? ((error as any).errorDetails as APIErrorDetails)
				: { message: JSON.stringify(error, null, 2) };
			resp.errorDetails = errorDetails;
			await dataChunkProcessor('data: [ERROR]', resp);
		}
	}
}
