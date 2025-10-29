import type { PreProcessorCall } from '@/spec/prompt';

import { toolStoreAPI } from '@/apis/baseapi';

// Lightweight runner that calls your backend. Replace endpoint as needed.
/**
 * @public
 * */
type ToolRunDescriptor = Pick<PreProcessorCall, 'toolBundleID' | 'toolSlug' | 'toolVersion'>;

export async function runPreprocessor(toolRef: ToolRunDescriptor, args: Record<string, unknown>): Promise<any> {
	let stringifiedArgs = '{}';
	try {
		stringifiedArgs = JSON.stringify(args);
	} catch (e) {
		const message =
			e instanceof Error ? e.message : 'Failed to serialize tool arguments. Please check for circular references.';
		throw new Error(message);
	}

	try {
		const resp = await toolStoreAPI.invokeTool(
			toolRef.toolBundleID,
			toolRef.toolSlug,
			toolRef.toolVersion,
			stringifiedArgs
		);
		const rawOutput = resp.output;

		try {
			return JSON.parse(rawOutput);
		} catch {
			// Return the raw string when it's not valid JSON (legacy tools might return primitives).
			return rawOutput;
		}
	} catch (err) {
		if (err instanceof Error) {
			throw err;
		}
		throw new Error(typeof err === 'string' ? err : 'Tool execution failed');
	}
}
