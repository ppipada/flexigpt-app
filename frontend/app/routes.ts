import { index, route, type RouteConfig } from '@react-router/dev/routes';

export default [
	index('./home/page.tsx'),
	route('agents', './agents/page.tsx'),
	route('chats', './chats/page.tsx'),
	route('docstores', './docstores/page.tsx'),
	route('prompts', './prompts/page.tsx'),
	route('settings', './settings/page.tsx'),
	route('modelpresets', './modelpresets/page.tsx'),
] satisfies RouteConfig;
