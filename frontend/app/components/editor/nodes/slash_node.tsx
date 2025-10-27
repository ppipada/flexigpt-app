import type { ReactNode } from 'react';

import type { TComboboxInputElement } from 'platejs';
import type { PlateEditor, PlateElementProps } from 'platejs/react';
import { PlateElement } from 'platejs/react';

import {
	InlineCombobox,
	InlineComboboxContent,
	InlineComboboxEmpty,
	InlineComboboxGroup,
	InlineComboboxGroupLabel,
	InlineComboboxInput,
	InlineComboboxItem,
} from '@/components/editor/nodes/inline_combobox';

/**
 * @public
 */
export interface ComboboxItem<T = unknown> {
	slug: string; // used as `value`
	displayName: string; // visible text
	icon?: ReactNode;
	keywords?: string[];
	focusEditor?: boolean;
	data: T; // the raw template/tool/etc you passed in
	onSelect: (editor: PlateEditor, item: T) => void;
}

/**
 * @public
 */
export interface ComboboxGroup<T = unknown> {
	group: string;
	items: ComboboxItem<T>[];
}

/**
 * @public
 */
export interface TriggerComboboxProps<T = unknown> extends PlateElementProps<TComboboxInputElement> {
	trigger: string;
	groups: ComboboxGroup<T>[];
}

/**
 * Re-usable combobox input that renders **whatever** groups/items you feed it.
 */
export function SlashInputElement<T = unknown>(props: TriggerComboboxProps<T>) {
	const { trigger, groups, element, editor, children, ...rest } = props;

	return (
		<PlateElement {...rest} editor={editor} element={element} as="span">
			<InlineCombobox element={element} trigger={trigger}>
				<InlineComboboxInput />

				<InlineComboboxContent>
					<InlineComboboxEmpty>No results</InlineComboboxEmpty>

					{groups.map(({ group, items }) => (
						<InlineComboboxGroup key={group}>
							<InlineComboboxGroupLabel>{group}</InlineComboboxGroupLabel>

							{items.map(item => (
								<InlineComboboxItem
									key={item.slug}
									value={item.slug}
									onClick={() => {
										item.onSelect(editor, item.data);
									}}
									label={item.displayName}
									group={group}
									focusEditor={item.focusEditor}
									keywords={item.keywords}
								>
									{item.icon && <div className="text-base-content/60 mr-2">{item.icon}</div>}
									{item.displayName}
								</InlineComboboxItem>
							))}
						</InlineComboboxGroup>
					))}
				</InlineComboboxContent>
			</InlineCombobox>

			{children}
		</PlateElement>
	);
}
