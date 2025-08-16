import { PlateElement, type PlateElementProps } from 'platejs/react';

export function BlockquoteElement(props: PlateElementProps) {
	return (
		<PlateElement
			as="blockquote"
			className="my-1 border-l-2 border-base-300 pl-6 italic text-base-content/80"
			{...props}
		/>
	);
}
