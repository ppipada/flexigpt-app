import type { PlateLeafProps } from 'platejs/react';
import { PlateLeaf } from 'platejs/react';

export function KbdLeaf(props: PlateLeafProps) {
	return (
		<PlateLeaf {...props} as="kbd" className="kbd kbd-sm font-mono text-base-content">
			{props.children}
		</PlateLeaf>
	);
}
