import { type RouteConfig, index, route } from '@react-router/dev/routes';

export default [
	index('./home.tsx'),
	route('agents', './agents/page.tsx'),
	route('chats', './chats/page.tsx'),
	route('docstores', './docstores/page.tsx'),
	route('prompts', './prompts/page.tsx'),
	route('settings', './settings/page.tsx'),
] satisfies RouteConfig;
