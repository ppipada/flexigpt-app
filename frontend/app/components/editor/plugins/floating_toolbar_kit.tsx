import { createPlatePlugin } from 'platejs/react';

import { FloatingToolbar } from '@/components/editor/nodes/floating_toolbar';
import { FloatingToolbarButtons } from '@/components/editor/nodes/floating_toolbar_buttons';

export const FloatingToolbarKit = [
	createPlatePlugin({
		key: 'floating-toolbar',
		render: {
			afterEditable: () => (
				<FloatingToolbar>
					<FloatingToolbarButtons />
				</FloatingToolbar>
			),
		},
	}),
];
