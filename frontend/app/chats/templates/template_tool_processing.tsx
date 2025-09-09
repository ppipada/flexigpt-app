// Lightweight runner that calls your backend. Replace endpoint as needed.
/**
 * @public
 * */
export type ToolRunResponse = { ok: true; result: any } | { ok: false; error: string };

export async function runPreprocessor(toolID: string, args: Record<string, unknown>): Promise<any> {
	// Assumption:
	// POST /api/preprocessors/run { toolID, args } -> { ok: true, result } | { ok: false, error }
	// - Endpoint: POST /api/preprocessors/run
	//   - Request: { toolID: string, args: Record<string, unknown> }
	//   - Response: one of:
	//     - { ok: true, result: any }, where result is the raw tool output (object/string/number/etc).
	//     - { ok: false, error: string }
	//   - If your backend already exists under a different path or signature, update src/chats/templates/tool_runner.ts accordingly.
	// - Security/cookies: we send credentials: 'include'. Adjust CORS/session as needed.
	// - Tool output mapping: The mapping saveAs/pathExpr is used to compute an effective variable value through effectiveVarValueLocal and computeRequirements. No need to manually copy tool output into tsenode.variables; it’s derived from toolStates + preProcessors.

	const res = await fetch('/api/preprocessors/run', {
		method: 'POST',
		headers: {
			'content-type': 'application/json',
		},
		body: JSON.stringify({ toolID, args }),
		credentials: 'include',
	});

	if (!res.ok) {
		const text = await res.text().catch(() => '');
		throw new Error(`Tool HTTP ${res.status}: ${text || res.statusText}`);
	}

	const data = await res.json().catch(() => ({}));
	if (data && typeof data === 'object') {
		if ('ok' in data) {
			// eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
			if (data.ok) return (data as ToolRunResponse & { ok: true }).result;
			throw new Error((data as ToolRunResponse & { ok: false }).error || 'Tool error');
		}
		// If backend returns raw result object (no ok field), accept it
		return data;
	}

	return data;
}
