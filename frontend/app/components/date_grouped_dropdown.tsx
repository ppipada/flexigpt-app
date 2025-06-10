import React, { Fragment } from 'react';

import { groupByDateBuckets } from '@/lib/date_utils';

interface GroupedDropdownProps<T> {
	items: T[];
	getDate: (item: T) => Date;
	getKey: (item: T) => React.Key;
	getLabel: (item: T) => React.ReactNode;
	onPick: (item: T) => void;
	focused: number;
	renderItemExtra?: (item: T) => React.ReactNode;
}

export function GroupedDropdown<T>({
	items,
	getDate,
	getKey,
	getLabel,
	onPick,
	focused,
	renderItemExtra,
}: GroupedDropdownProps<T>) {
	const buckets = groupByDateBuckets(items, getDate);

	let itemIndex = 0; // For tracking focused index across all groups

	return (
		<ul className="absolute left-0 right-0 mt-0 max-h-80 overflow-y-auto bg-base-200 rounded-2xl shadow-lg text-sm">
			{Object.entries(buckets)
				.filter(([, arr]) => arr.length)
				.map(([label, arr]) => (
					<Fragment key={label}>
						<li className="px-12 py-2 text-neutral/60">{label}</li>
						{arr.map(item => {
							const isFocused = itemIndex === focused;
							const li = (
								<li
									key={getKey(item)}
									onClick={() => {
										onPick(item);
									}}
									className={`flex justify-between items-center px-12 py-2 cursor-pointer hover:bg-base-100 ${
										isFocused ? 'bg-base-100' : ''
									}`}
								>
									<span className="truncate">{getLabel(item)}</span>
									{renderItemExtra && (
										<span className="hidden lg:block text-neutral text-xs">{renderItemExtra(item)}</span>
									)}
								</li>
							);
							itemIndex++;
							return li;
						})}
					</Fragment>
				))}
		</ul>
	);
}
