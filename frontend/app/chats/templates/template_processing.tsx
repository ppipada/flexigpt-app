import {
	type MessageBlock,
	type PreProcessorCall,
	type PromptTemplate,
	type PromptVariable,
	VarSource,
} from '@/spec/prompt';

import type { SelectedTemplateForRun, TemplateSelectionElementNode, ToolState } from '@/chats/templates/template_spec';

export function buildInitialToolStates(template?: PromptTemplate): Record<string, ToolState> {
	const states: Record<string, ToolState> = {};
	if (template?.preProcessors?.length) {
		for (const p of template.preProcessors) {
			states[p.id] = {
				args: p.args ?? {},
				status: 'pending',
			};
		}
	}
	return states;
}

/**
 * Merge templateSnapshot with local overrides to produce effective template structures.
 */
export function computeEffectiveTemplate(el: TemplateSelectionElementNode): {
	template: PromptTemplate | undefined;
	blocks: MessageBlock[];
	variablesSchema: PromptVariable[];
	preProcessors: PreProcessorCall[];
} {
	const base = el.templateSnapshot;
	const blocks = el.overrides?.blocks ?? base?.blocks ?? [];
	const variablesSchema = el.overrides?.variables ?? base?.variables ?? [];
	const preProcessors = el.overrides?.preProcessors ?? base?.preProcessors ?? [];
	return { template: base, blocks, variablesSchema, preProcessors };
}

export function effectiveVarValueLocal(
	varDef: PromptVariable,
	userValues: Record<string, unknown>,
	toolStates?: Record<string, ToolState>,
	preProcessors?: PreProcessorCall[]
): unknown {
	if (userValues[varDef.name] !== undefined && userValues[varDef.name] !== null) {
		return userValues[varDef.name];
	}
	if (varDef.source === VarSource.Static && varDef.staticVal !== undefined) {
		return varDef.staticVal;
	}
	if (varDef.default !== undefined && varDef.default !== '') {
		return varDef.default;
	}

	if (varDef.source === VarSource.Tool && toolStates) {
		const bySaveAs = new Map<string, any>();
		(preProcessors ?? []).forEach(p => {
			const st = toolStates[p.id];
			if (st.result !== undefined) {
				const val = pickByPathExpr(st.result, p.pathExpr);
				bySaveAs.set(p.saveAs, val);
			}
		});
		if (bySaveAs.has(varDef.name)) return bySaveAs.get(varDef.name);
		// fallback to “first result” if mapping not found
		const hit = Object.entries(toolStates).find(([_id, st]) => st.result !== undefined);
		if (hit) {
			const p = (preProcessors ?? []).find(pp => pp.id === hit[0]);
			return pickByPathExpr(hit[1].result, p?.pathExpr);
		}
	}
	return undefined;
}

function effectiveVarValue(
	varDef: PromptVariable,
	userValues: Record<string, unknown>,
	toolStates?: Record<string, ToolState>
): unknown {
	// Local override always wins if present
	if (userValues[varDef.name] !== undefined && userValues[varDef.name] !== null) {
		return userValues[varDef.name];
	}

	// Source & defaults
	switch (varDef.source) {
		case VarSource.Static:
			if (varDef.staticVal !== undefined && varDef.staticVal !== '') {
				return varDef.staticVal;
			}
			break;
		case VarSource.Tool: {
			if (toolStates) {
				const hit = Object.values(toolStates).find(st => st.result !== undefined);
				if (hit?.result !== undefined) return hit.result;
			}
			break;
		}
		case VarSource.User:
		default:
			break;
	}

	// Fallback declared default
	if (varDef.default !== undefined && varDef.default !== '') return varDef.default;

	return undefined;
}

export function computeRequirements(
	variablesSchema: PromptVariable[],
	variableValues: Record<string, unknown>,
	preProcessors: PreProcessorCall[] = [],
	toolStates?: Record<string, ToolState>
) {
	const requiredNames: string[] = [];
	const values: Record<string, unknown> = { ...variableValues };

	const toolResultBySaveAs = new Map<string, unknown>();
	for (const p of preProcessors) {
		const st = toolStates?.[p.id];
		if (st && st.result !== undefined) {
			toolResultBySaveAs.set(p.saveAs, pickByPathExpr(st.result, p.pathExpr));
		}
	}

	// Fill effective values
	for (const v of variablesSchema) {
		let val = values[v.name];
		if (val === undefined) {
			if (v.source === VarSource.Tool && toolResultBySaveAs.has(v.name)) {
				val = toolResultBySaveAs.get(v.name);
			} else {
				val = effectiveVarValue(v, variableValues, toolStates);
			}
		}
		values[v.name] = val;
	}

	// Identify required ones that remain empty
	for (const v of variablesSchema) {
		const val = values[v.name];
		if (v.required && (val === undefined || val === null || val === '')) {
			requiredNames.push(v.name);
		}
	}

	// Tools to run summary
	const toolsToRun = preProcessors.map(p => {
		const st = toolStates?.[p.id];
		return {
			id: p.id,
			toolID: p.toolID,
			args: st?.args ?? p.args,
			saveAs: p.saveAs,
			pathExpr: p.pathExpr,
			onError: p.onError,
			status: st?.status ?? 'pending',
		};
	});

	const pendingToolsCount = toolsToRun.filter(t => t.status === 'pending').length;

	return {
		variableValues: values,
		requiredVariables: requiredNames,
		requiredCount: requiredNames.length,
		toolsToRun,
		pendingToolsCount,
	};
}

export function makeSelectedTemplateForRun(tsenode: TemplateSelectionElementNode): SelectedTemplateForRun {
	const { template, blocks, variablesSchema, preProcessors } = computeEffectiveTemplate(tsenode);

	const effTemplate: PromptTemplate =
		template ??
		({
			id: '',
			displayName: tsenode.templateSlug,
			slug: tsenode.templateSlug,
			isEnabled: true,
			description: '',
			tags: [],
			blocks,
			variables: variablesSchema,
			preProcessors,
			version: tsenode.templateVersion,
			createdAt: new Date().toISOString(),
			modifiedAt: new Date().toISOString(),
			isBuiltIn: false,
		} as PromptTemplate);

	const req = computeRequirements(variablesSchema, tsenode.variables, preProcessors, tsenode.toolStates);

	return {
		type: 'templateSelection',
		bundleID: tsenode.bundleID,
		templateSlug: tsenode.templateSlug,
		templateVersion: tsenode.templateVersion,
		selectionID: tsenode.selectionID,
		template: effTemplate,
		blocks,
		variablesSchema,
		preProcessors,
		variableValues: req.variableValues,
		requiredVariables: req.requiredVariables,
		requiredCount: req.requiredCount,
		toolsToRun: req.toolsToRun,
		isReady: req.requiredCount === 0 && req.toolsToRun.every(t => t.status === 'done'),
	};
}

// Returns the content of the last block as plain text if it is from role user.
// If not found, returns empty string.
export function getLastUserBlockContent(el: TemplateSelectionElementNode): string {
	const { blocks } = computeEffectiveTemplate(el);
	if (blocks.length > 0) {
		const b = blocks[blocks.length - 1];
		if (b.role.toLowerCase() === 'user') {
			return b.content;
		}
	}
	return '';
}

// JSON-like types for safe indexing
type JSONValue = null | boolean | number | string | Array<JSONValue> | JSONObject;
interface JSONObject {
	[k: string]: JSONValue;
}

function tokenizePath(expr: string): Array<string | number> {
	const tokens: Array<string | number> = [];
	let i = 0;
	let current = '';

	const pushCurrent = () => {
		if (current.length > 0) {
			tokens.push(current);
			current = '';
		}
	};

	while (i < expr.length) {
		const ch = expr[i];

		if (ch === '.') {
			pushCurrent();
			i++;
			continue;
		}

		if (ch === '[') {
			pushCurrent();
			i++; // consume '['

			// skip whitespace
			while (i < expr.length && /\s/.test(expr[i])) i++;

			// quoted key ['...'] or ["..."]
			if (expr[i] === "'" || expr[i] === '"') {
				const quote = expr[i++];
				let val = '';
				while (i < expr.length) {
					const c = expr[i++];
					if (c === '\\') {
						if (i < expr.length) val += expr[i++];
						continue;
					}
					if (c === quote) break;
					val += c;
				}
				// skip whitespace to closing bracket
				while (i < expr.length && /\s/.test(expr[i])) i++;
				if (expr[i] === ']') i++; // consume ']'
				tokens.push(val);
			} else {
				// bare number or identifier until ']'
				let raw = '';
				while (i < expr.length && expr[i] !== ']') raw += expr[i++];
				if (expr[i] === ']') i++; // consume ']'
				const val = raw.trim();
				if (/^\d+$/.test(val)) tokens.push(Number(val));
				else if (val.length > 0) tokens.push(val);
			}

			continue;
		}

		// regular char as part of a dot segment
		current += ch;
		i++;
	}

	pushCurrent();

	// remove accidental empty tokens
	return tokens.filter(t => !(typeof t === 'string' && t.length === 0));
}

function pickByPathExpr(input: unknown, pathExpr?: string): unknown {
	if (!pathExpr) return input;

	const tokens = tokenizePath(pathExpr);
	if (tokens.length === 0) return input;

	let cur: unknown = input;

	for (const tok of tokens) {
		if (cur === null || cur === undefined) return undefined;

		if (typeof tok === 'number') {
			if (!Array.isArray(cur)) return undefined;
			const arr = cur as Array<JSONValue>;
			cur = arr[tok];
			continue;
		}

		// string key
		if (typeof cur !== 'object') return undefined;
		const obj = cur as JSONObject;
		if (!(tok in obj)) return undefined;
		cur = obj[tok];
	}

	return cur;
}
