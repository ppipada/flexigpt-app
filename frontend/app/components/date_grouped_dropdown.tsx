import React, { Fragment } from 'react';

import { groupByDateBuckets } from '@/lib/date_utils';

interface GroupedDropdownProps<T> {
	items: T[];
	focused: number;
	getDate: (item: T) => Date;
	getKey: (item: T) => React.Key;
	getLabel: (item: T) => React.ReactNode;
	onPick: (item: T) => void;
	renderItemExtra?: (item: T) => React.ReactNode;
}

export function GroupedDropdown<T>({
	items,
	focused,
	getDate,
	getKey,
	getLabel,
	onPick,
	renderItemExtra,
}: GroupedDropdownProps<T>) {
	const buckets = groupByDateBuckets(items, getDate);

	let globalIndex = 0; // single running index for keyboard focus

	return (
		<ul className="w-full text-sm">
			{Object.entries(buckets)
				.filter(([, arr]) => arr.length)
				.map(([label, arr]) => (
					<Fragment key={label}>
						<li className="px-12 py-2 text-neutral/60 text-xs">{label}</li>

						{arr.map(item => {
							const isFocused = globalIndex === focused;

							const li = (
								<li
									key={getKey(item)}
									data-index={globalIndex}
									onClick={() => {
										onPick(item);
									}}
									className={`flex justify-between items-center px-12 py-2 cursor-pointer hover:bg-base-100 ${
										isFocused ? 'bg-base-100' : ''
									}`}
								>
									<span className="truncate">{getLabel(item)}</span>

									{renderItemExtra && (
										<span className="hidden lg:block text-neutral/60 text-xs">{renderItemExtra(item)}</span>
									)}
								</li>
							);

							globalIndex++;
							return li;
						})}
					</Fragment>
				))}
		</ul>
	);
}
