import type { FC, JSX } from 'react';
import { useCallback, useEffect, useMemo, useState } from 'react';

import { FiMonitor, FiMoon, FiSun } from 'react-icons/fi';

import { type AppTheme, ThemeType, toProviderName, toThemeType } from '@/spec/setting';
import { DAISYUI_BUILTIN_THEMES } from '@/spec/theme_consts';

import { updateStartupTheme, useStartupTheme } from '@/hooks/use_startup_theme';
import { useTheme } from '@/hooks/use_theme_provider';

import { settingstoreAPI } from '@/apis/baseapi';

import Dropdown, { type DropdownItem } from '@/components/dropdown';

/* ———————————————————————————————————————— helpers ——————————————————————————————————————— */
type OtherThemeName = (typeof DAISYUI_BUILTIN_THEMES)[number];
const isOtherThemeName = (n: string): n is OtherThemeName => DAISYUI_BUILTIN_THEMES.includes(n);

/* ———————————————————————————————————————— sub-component ——————————————————————————————————— */
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

/* ———————————————————————————————————————— main selector —————————————————————————————————— */
export const ThemeSelector: FC = () => {
	const [startupTheme, startupReady] = useStartupTheme();
	const { theme: providerTheme, setTheme } = useTheme();

	/* derived state */
	const current = useMemo(() => toThemeType(providerTheme), [providerTheme]);

	/* dropdown items never change -> memo once */
	const dropdownItems = useMemo(
		() =>
			Object.fromEntries(DAISYUI_BUILTIN_THEMES.map(t => [t, { isEnabled: true }])) as Record<
				OtherThemeName,
				DropdownItem
			>,
		[]
	);

	const [otherName, setOtherName] = useState<OtherThemeName>();

	/* initialise “other” theme name once start-up theme is loaded */
	useEffect(() => {
		if (!startupReady || !startupTheme) return;

		if (startupTheme.type === ThemeType.Other && isOtherThemeName(startupTheme.name)) {
			setOtherName(startupTheme.name);
		}
	}, [startupReady, startupTheme]);

	/* ————————————————————————————— apply theme —————————————————————————————— */
	const applyTheme = useCallback(
		async (type: ThemeType, name?: OtherThemeName) => {
			if (!startupReady) return; // protect against very early calls

			const providerKey = type === ThemeType.Other ? (name ?? otherName) : toProviderName(type);
			if (!providerKey) return;

			/* optimistic update */
			setTheme(providerKey);

			const newTheme: AppTheme = {
				type,
				name: type === ThemeType.Other ? providerKey : type,
			};

			try {
				await settingstoreAPI.setAppTheme(newTheme);
				updateStartupTheme(newTheme);
				console.log('[Theme] changed to', newTheme.name);
			} catch (err) {
				console.error('[Theme] failed to persist, reverting', err);
				// fallback to last known persistent value
				if (startupTheme) {
					setTheme(toProviderName(startupTheme));
				}
			}
		},
		[otherName, setTheme, startupReady, startupTheme]
	);

	/* ————————————————————————————— UI —————————————————————————————— */
	if (!startupReady) {
		return <span className="loading loading-dots loading-sm" />;
	}

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
