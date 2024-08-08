import { APIErrorDetails, APIRequestDetails, APIResponseDetails } from 'aiprovidermodel';
import { AxiosError, AxiosRequestConfig, AxiosResponse } from 'axios';
import { log } from 'logger';

export function filterSensitiveInfoFromJsonString(jsonString: string): string {
	const jsonObj = JSON.parse(jsonString);
	const filteredObj = filterSensitiveInfo(jsonObj);
	return JSON.stringify(filteredObj);
}

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
		if (sensitiveKeys.some(sensitiveKey => key.toLowerCase().includes(sensitiveKey))) {
			filteredObj[key] = '****' + obj[key].slice(-4);
		} else {
			filteredObj[key] = filterSensitiveInfo(obj[key]);
		}
	}
	return filteredObj;
}

function generateCurlCommand(config: APIRequestDetails): string {
	let curlCommand = 'curl -X ' + config.method?.toUpperCase() + ' ';
	curlCommand += '"' + config.url + '" ';

	if (config.headers) {
		Object.keys(config.headers).forEach(key => {
			const value = config.headers[key];
			curlCommand += '-H "' + key + ': ' + value + '" ';
		});
	}

	if (config.data) {
		curlCommand += "-d '" + JSON.stringify(config.data) + "' ";
	}

	return curlCommand;
}

function getRequestDetails(config: AxiosRequestConfig, logDetails = false): AxiosRequestConfig {
	const requestDetails: APIRequestDetails = {
		url: config.url,
		method: config.method,
		headers: filterSensitiveInfo(config.headers),
		params: config.params,
		data: config.data,
		timeout: config.timeout,
	};

	requestDetails.curlCommand = generateCurlCommand(requestDetails);
	if (logDetails) {
		log.debug('Request Details:', requestDetails);
		log.debug('cURL Command:', requestDetails.curlCommand);
	}

	// Attach requestDetails to config for later use in response interceptor
	(config as any).requestDetails = requestDetails;

	return config;
}

function getRequestErrorDetails(error: AxiosError, logDetails = false): APIErrorDetails {
	const errorDetails: APIErrorDetails = {
		message: error.message,
		requestDetails: error.config
			? {
					url: error.config.url,
					method: error.config.method,
					headers: filterSensitiveInfo(error.config.headers),
					params: error.config.params,
					data: filterSensitiveInfo(error.config.data),
					timeout: error.config.timeout,
				}
			: undefined,
	};
	if (logDetails) {
		log.error('Request Error Details:', errorDetails);
	}

	return errorDetails;
}

function getResponseDetails(response: AxiosResponse, logDetails = false): AxiosResponse {
	const requestDetails = (response.config as any).requestDetails as APIRequestDetails;
	const responseDetails: APIResponseDetails = {
		data: response.data,
		status: response.status,
		headers: response.headers,
		requestDetails: requestDetails,
	};
	if (logDetails) {
		log.debug('Response Details:', responseDetails);
	}
	(response.config as any).responseDetails = responseDetails;
	return response;
}

function getResponseErrorDetails(error: AxiosError, logDetails = false): APIErrorDetails {
	const errorDetails: APIErrorDetails = {
		message: error.message,
	};

	if (error.response) {
		const requestDetails = (error.config as any).requestDetails as APIRequestDetails;
		errorDetails.responseDetails = {
			data: error.response.data,
			status: error.response.status,
			headers: filterSensitiveInfo(error.response.headers),
			requestDetails: requestDetails,
		};
		errorDetails.requestDetails = requestDetails;
	} else if (error.request) {
		errorDetails.request = error.request;
	}
	if (logDetails) {
		log.error('Response Error Details:', errorDetails);
	}
	(error as any).errorDetails = errorDetails; // Attach errorDetails to error
	return errorDetails;
}

function setRequestInterceptor(axiosInstance: any, logRequests: boolean) {
	axiosInstance.interceptors.request.use(
		(config: AxiosRequestConfig) => getRequestDetails(config, logRequests),
		(error: AxiosError) => {
			const errorDetails = getRequestErrorDetails(error, logRequests);
			return Promise.reject(errorDetails);
		}
	);
}

function setResponseInterceptor(axiosInstance: any, logRequests: boolean) {
	axiosInstance.interceptors.response.use(
		(response: AxiosResponse) => getResponseDetails(response, logRequests),
		(error: AxiosError) => {
			const errorDetails = getResponseErrorDetails(error, logRequests);
			return Promise.reject(errorDetails);
		}
	);
}

export function setupInterceptors(axiosInstance: any, logRequests: boolean) {
	setRequestInterceptor(axiosInstance, logRequests);
	setResponseInterceptor(axiosInstance, logRequests);
}
