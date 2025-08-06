import type { FC, JSX } from 'react';
import { useMemo, useState } from 'react';

import { FiMonitor, FiMoon, FiSun } from 'react-icons/fi';

import { type AppTheme, ThemeType } from '@/spec/setting';
import { CustomThemeDark, CustomThemeLight, DAISYUI_BUILTIN_THEMES } from '@/spec/theme_consts';

import { useTheme } from '@/hooks/use_theme';

import { settingstoreAPI } from '@/apis/baseapi';
import { getStartupThemeSync, updateStartupTheme } from '@/apis/builtin_theme_cache';

import Dropdown, { type DropdownItem } from '@/components/dropdown';

type OtherThemeName = (typeof DAISYUI_BUILTIN_THEMES)[number];
const isOtherThemeName = (n: string): n is OtherThemeName => DAISYUI_BUILTIN_THEMES.includes(n);

export const toProviderName = (t: ThemeType | AppTheme): string => {
	const type = typeof t === 'string' ? t : t.type;
	if (type === ThemeType.System) return 'system';
	if (type === ThemeType.Light) return CustomThemeLight;
	if (type === ThemeType.Dark) return CustomThemeDark;
	return typeof t === 'string' ? t : t.name; /* ThemeType.Other */
};
const toThemeType = (name: string): ThemeType => {
	if (name === CustomThemeLight || name === 'light') return ThemeType.Light;
	if (name === CustomThemeDark || name === 'dark') return ThemeType.Dark;
	if (name === 'system') return ThemeType.System;
	return ThemeType.Other;
};

interface ThemeRadioProps {
	label: string;
	value: ThemeType;
	icon: JSX.Element;
	current: ThemeType;
	onChange: (v: ThemeType) => void;
}

const ThemeRadio: FC<ThemeRadioProps> = ({ label, value, icon, current, onChange }) => (
	<label className="flex items-center gap-2 cursor-pointer">
		<input
			type="radio"
			className="radio radio-accent"
			checked={current === value}
			onChange={() => {
				onChange(value);
			}}
		/>
		{icon}
		<span className="text-sm">{label}</span>
	</label>
);

export const ThemeSelector: FC = () => {
	const { theme: providerTheme, setTheme } = useTheme();
	const current = useMemo(() => toThemeType(providerTheme), [providerTheme]);

	const [otherName, setOtherName] = useState<OtherThemeName | undefined>(() => {
		const t = getStartupThemeSync();
		return t.type === ThemeType.Other && isOtherThemeName(t.name) ? t.name : undefined;
	});

	const dropdownItems = useMemo(
		() =>
			Object.fromEntries(DAISYUI_BUILTIN_THEMES.map(t => [t, { isEnabled: true }])) as Record<
				OtherThemeName,
				DropdownItem
			>,
		[]
	);

	const applyTheme = async (type: ThemeType, name?: OtherThemeName) => {
		const providerKey = type === ThemeType.Other ? (name ?? otherName) : toProviderName(type);
		if (!providerKey) return;

		/* optimistic */
		setTheme(providerKey);

		const newTheme: AppTheme = {
			type,
			name: type === ThemeType.Other ? providerKey : String(type),
		};

		try {
			await settingstoreAPI.setAppTheme(newTheme);
			updateStartupTheme(newTheme);
			console.log('[Theme] changed to', newTheme.name);
		} catch (err) {
			console.error('[Theme] failed to persist, reverting', err);
			setTheme(toProviderName(getStartupThemeSync()));
		}
	};

	return (
		<div className="flex gap-6 items-center">
			<ThemeRadio
				label="System"
				value={ThemeType.System}
				icon={<FiMonitor />}
				current={current}
				onChange={applyTheme}
			/>
			<ThemeRadio label="Light" value={ThemeType.Light} icon={<FiSun />} current={current} onChange={applyTheme} />
			<ThemeRadio label="Dark" value={ThemeType.Dark} icon={<FiMoon />} current={current} onChange={applyTheme} />

			<label className="flex items-center gap-2 cursor-pointer">
				<input
					type="radio"
					className="radio radio-accent"
					checked={current === ThemeType.Other}
					onChange={() => applyTheme(ThemeType.Other, otherName)}
				/>
				<div className="w-50">
					<Dropdown<OtherThemeName>
						dropdownItems={dropdownItems}
						selectedKey={otherName ?? DAISYUI_BUILTIN_THEMES[0]}
						onChange={async key => {
							setOtherName(key);
							await applyTheme(ThemeType.Other, key);
						}}
						filterDisabled={false}
						title="Select Theme"
						getDisplayName={k => k[0].toUpperCase() + k.slice(1)}
					/>
				</div>
			</label>
		</div>
	);
};
