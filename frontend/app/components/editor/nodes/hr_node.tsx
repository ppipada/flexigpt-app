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
						'bg-base-300 h-0.5 rounded-sm border-none bg-clip-content',
						selected && focused && 'ring-primary ring-offset-base-100 ring-2 ring-offset-2',
						!readOnly && 'cursor-pointer'
					)}
				/>
			</div>
			{props.children}
		</PlateElement>
	);
}
