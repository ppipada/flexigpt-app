import type { PlateElementProps } from 'platejs/react';
import { PlateElement, useFocused, useReadOnly, useSelected } from 'platejs/react';

function cx(...parts: Array<string | false | null | undefined>) {
	return parts.filter(Boolean).join(' ');
}

export function HrElement(props: PlateElementProps) {
	const readOnly = useReadOnly();
	const selected = useSelected();
	const focused = useFocused();

	return (
		<PlateElement {...props}>
			<div className="py-6" contentEditable={false}>
				<hr
					className={cx(
						'h-0.5 rounded-sm border-none bg-base-300 bg-clip-content',
						selected && focused && 'ring-2 ring-primary ring-offset-2 ring-offset-base-100',
						!readOnly && 'cursor-pointer'
					)}
				/>
			</div>
			{props.children}
		</PlateElement>
	);
}
