// eslint-disable-next-line no-restricted-imports
import type { FC } from 'react';

import { isOrderedList } from '@platejs/list';
import { useTodoListElement, useTodoListElementState } from '@platejs/list/react';
import type { TListElement } from 'platejs';
import { type PlateElementProps, type RenderNodeWrapper, useReadOnly } from 'platejs/react';

// Tiny helper to merge class names (no external deps)
function cx(...classes: Array<string | undefined | null | false>) {
	return classes.filter(Boolean).join(' ');
}

const config: Record<
	string,
	{
		Li: FC<PlateElementProps>;
		Marker: FC<PlateElementProps>;
	}
> = {
	todo: {
		Li: TodoLi,
		Marker: TodoMarker,
	},
};

export const BlockList: RenderNodeWrapper = props => {
	if (!props.element.listStyleType) return;
	return props => <List {...props} />;
};

/**
 * @public
 */
export function List(props: PlateElementProps) {
	const { attributes } = props;
	const { listStart, listStyleType } = props.element as TListElement;
	const { Li, Marker } = config[listStyleType] ?? {};
	const ordered = isOrderedList(props.element);

	const className = cx('relative m-0 p-0', attributes.className);

	if (ordered) {
		return (
			<ol {...attributes} className={className} style={{ listStyleType }} start={listStart}>
				{<Marker {...props} />}
				{<Li {...props} />}
			</ol>
		);
	}

	return (
		<ul {...attributes} className={className} style={{ listStyleType }}>
			{<Marker {...props} />}
			{<Li {...props} />}
		</ul>
	);
}

/**
 * @public
 */
export function TodoMarker(props: PlateElementProps) {
	const state = useTodoListElementState({ element: props.element });
	const { checkboxProps } = useTodoListElement(state);
	const readOnly = useReadOnly();

	return (
		<div contentEditable={false}>
			<input
				type="checkbox"
				className={cx('checkbox checkbox-sm checkbox-primary absolute top-1 -left-6')}
				disabled={readOnly}
				aria-disabled={readOnly || undefined}
				tabIndex={readOnly ? -1 : 0}
				{...checkboxProps}
			/>
		</div>
	);
}

/**
 * @public
 */
export function TodoLi(props: PlateElementProps) {
	const isChecked =
		typeof props.element === 'object' &&
		'checked' in props.element &&
		Boolean((props.element as { checked?: boolean }).checked);

	return <li className={cx('list-none', isChecked && 'line-through opacity-60')}>{props.children}</li>;
}
