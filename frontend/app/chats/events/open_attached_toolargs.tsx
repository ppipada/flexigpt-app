import { emitCustomEvent, useEvent } from '@/hooks/use_event';

import type { ToolArgsTarget } from '@/chats/tools/tool_user_args_modal';

const OPEN_TOOL_ARGS_EVENT = 'tool-args:open';
type Detail = { target: ToolArgsTarget };

export function dispatchOpenToolArgs(target: ToolArgsTarget, eventTarget?: EventTarget | null) {
	emitCustomEvent<Detail>(OPEN_TOOL_ARGS_EVENT, { target }, eventTarget);
}

export function useOpenToolArgs(handler: (target: ToolArgsTarget) => void, eventTarget?: EventTarget | null) {
	useEvent<Detail, CustomEvent<Detail>>(OPEN_TOOL_ARGS_EVENT, {
		target: eventTarget,
		handler: ev => {
			handler(ev.detail.target);
		},
	});
}
