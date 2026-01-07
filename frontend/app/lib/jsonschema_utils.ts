export type JSONValue = string | number | boolean | null | JSONValue[] | { [key: string]: JSONValue };
export type JSONSchema = JSONValue;
export type JSONObject = Exclude<JSONValue, string | number | boolean | null | JSONValue[]>;

export type JSONRawString = string;

export function isJSONObject(value: JSONSchema): value is JSONObject {
	return typeof value === 'object' && value !== null && !Array.isArray(value);
}

export function getJSONObject(value: JSONSchema | undefined): JSONObject | undefined {
	if (value && isJSONObject(value)) {
		return value;
	}
	return undefined;
}

/**
 * @public
 */
export function isStringArray(value: JSONValue | undefined): value is string[] {
	return Array.isArray(value) && value.every(item => typeof item === 'string');
}

export function getRequiredFromJSONSchema(schema: JSONSchema | undefined): string[] | undefined {
	// Must be an object
	if (!schema || !isJSONObject(schema)) return undefined;

	// `required` is just another JSONValue on the object
	const candidate = schema.required as JSONValue | undefined;

	// Must be exactly an array of strings, otherwise treat as invalid
	if (!isStringArray(candidate)) return undefined;

	// Here `candidate` is typed as `string[]`
	return candidate;
}

export function getPropertiesFromJSONSchema(schema: JSONSchema | undefined): JSONObject | undefined {
	// Must be an object
	if (!schema || !isJSONObject(schema)) return undefined;
	const props = (schema as { properties?: JSONValue }).properties;

	if (props && typeof props === 'object' && !Array.isArray(props)) {
		return props;
	}
	return undefined;
}
