import type { ComponentProps } from 'react';

import { cn } from '@udecode/cn';

import { Toolbar } from '@/components/editor/nodes/toolbar';

export function FixedToolbar(props: ComponentProps<typeof Toolbar>) {
	return (
		<Toolbar
			{...props}
			className={cn(
				// Layout/position
				'sticky top-0 left-0 w-full justify-between overflow-x-auto',
				// DaisyUI look & feel
				'border-base-300 rounded-t-(--rounded-box) border-b',
				// Background + blur (with graceful supports fallback)
				'bg-base-100/95 supports-backdrop-filter:bg-base-100/60 p-1 supports-backdrop-filter:backdrop-blur-sm',
				// Merge user className
				props.className
			)}
		/>
	);
}
