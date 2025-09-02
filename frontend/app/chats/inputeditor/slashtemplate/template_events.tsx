import { emitCustomEvent, useEvent } from '@/hooks/use_event';

const TEMPLATE_FLASH_EVENT = 'tpl-toolbar:flash';
const TEMPLATE_VARS_UPDATED_EVENT = 'tpl-vars:updated';

/**
 * Returns a boolean that becomes true for `durationMs` after each flash event.
 */
export function useTemplateFlashEvent(durationMs = 800, target?: EventTarget | null) {
	const { pulse } = useEvent<undefined, CustomEvent<void>>(TEMPLATE_FLASH_EVENT, {
		target,
		pulseMs: durationMs,
	});
	return pulse;
}
export function dispatchTemplateFlashEvent(target?: EventTarget | null) {
	emitCustomEvent(TEMPLATE_FLASH_EVENT, undefined, target);
}

type TemplateVarsUpdatedDetail = { selectionID?: string };
/**
 * @public
 * Low-level subscription with typed detail passthrough.
 */
export function useTemplateVarsUpdatedEvent(
	onUpdate: (detail: TemplateVarsUpdatedDetail, ev: CustomEvent<TemplateVarsUpdatedDetail>) => void,
	target?: EventTarget | null
) {
	useEvent<TemplateVarsUpdatedDetail, CustomEvent<TemplateVarsUpdatedDetail>>(TEMPLATE_VARS_UPDATED_EVENT, {
		target,
		handler: ev => {
			onUpdate(ev.detail, ev);
		},
	});
}

export function dispatchTemplateVarsUpdated(selectionID?: string, target?: EventTarget | null) {
	emitCustomEvent<TemplateVarsUpdatedDetail>(TEMPLATE_VARS_UPDATED_EVENT, { selectionID }, target);
}

/**
 * Convenience: only fires callback when selectionID matches.
 */
export function useTemplateVarsUpdatedForSelection(
	selectionID: string | undefined,
	onMatch: () => void,
	target?: EventTarget | null
) {
	useEvent<TemplateVarsUpdatedDetail, CustomEvent<TemplateVarsUpdatedDetail>>(TEMPLATE_VARS_UPDATED_EVENT, {
		target,
		filter: ev => Boolean(selectionID) && ev.detail.selectionID === selectionID,
		handler: () => {
			onMatch();
		},
	});
}
