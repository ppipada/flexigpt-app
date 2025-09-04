import { useEffect, useState } from 'react';

import type { Tool, ToolListItem } from '@/spec/tool';

import { toolStoreAPI } from '@/apis/baseapi';
import { getAllTools } from '@/apis/list_helper';

export function useTools() {
	const [data, setData] = useState<ToolListItem[]>([]);
	const [loading, setLoading] = useState(true);

	useEffect(() => {
		let cancelled = false;
		setLoading(true);
		getAllTools()
			.then(res => {
				if (cancelled) {
					return;
				}
				setData(res);
			})
			.finally(() => {
				if (cancelled) {
					return;
				}
				setLoading(false);
			});
		return () => {
			cancelled = true;
		};
	}, []);

	return { data, loading };
}

/**
 * @public
 */
export function useTool(bundleID: string, slug: string, version: string) {
	const [tool, setTool] = useState<Tool | undefined>();
	useEffect(() => {
		if (!bundleID || !slug || !version) return;

		let cancelled = false;
		toolStoreAPI.getTool(bundleID, slug, version).then(res => {
			if (cancelled) {
				return;
			}
			setTool(res);
		});
		return () => {
			cancelled = true;
		};
	}, [bundleID, slug, version]);

	return tool;
}
