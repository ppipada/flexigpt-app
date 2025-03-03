'use client';

export enum FeatureFlag {
	PROMPTS = 'PROMPTS',
	DOCUMENT_STORES = 'DOCUMENT_STORES',
	AGENTS = 'AGENTS',
}

export function isFeatureEnabled(feature: FeatureFlag): boolean {
	return process.env[`NEXT_PUBLIC_FEATURE_${feature}`] === 'true';
}
