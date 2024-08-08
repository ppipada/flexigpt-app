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

function checkNonEmptyObject(obj: any): boolean {
	if (obj && typeof obj === 'object' && Object.keys(obj).length !== 0) {
		return true;
	}
	return false;
}

function getRequestDetails(config: AxiosRequestConfig, logDetails = false): AxiosRequestConfig {
	const requestDetails: APIRequestDetails = {
		url: config.url,
		method: config.method,
		headers: checkNonEmptyObject(config.headers) ? filterSensitiveInfo(config.headers) : undefined,
		params: checkNonEmptyObject(config.params) ? config.params : undefined,
		data: checkNonEmptyObject(config.data) ? config.data : undefined,
		timeout: config.timeout,
	};

	requestDetails.curlCommand = generateCurlCommand(requestDetails);
	if (logDetails) {
		log.debug('Request Details:', requestDetails);
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
					headers: checkNonEmptyObject(error.config.headers) ? filterSensitiveInfo(error.config.headers) : undefined,
					params: checkNonEmptyObject(error.config.params) ? error.config.params : undefined,
					data: checkNonEmptyObject(error.config.data) ? filterSensitiveInfo(error.config.data) : undefined,
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
		data: checkNonEmptyObject(response.data) ? response.data : undefined,
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
	// log.debug('Full error', JSON.stringify(error, null, 2));
	let msg = `Message: ${error.message}`;
	if (error.code) {
		msg += `\n\nCode: ${error.code}`;
	}
	const errorDetails: APIErrorDetails = {
		message: msg,
	};
	let status = error.status;
	if (error.response) {
		const requestDetails = (error.config as any).requestDetails as APIRequestDetails;
		if (error.response.status) {
			status = error.response.status;
		}
		errorDetails.responseDetails = {
			data: checkNonEmptyObject(error.response.data) ? error.response.data : undefined,
			status: error.response.status,
			headers: filterSensitiveInfo(error.response.headers),
			requestDetails: undefined,
		};
		errorDetails.requestDetails = requestDetails;
	}
	if (status) {
		errorDetails.message += `\n\nStatus: ${status}`;
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
