import { useCallback, useEffect, useMemo, useState } from 'react';

import { FiMonitor, FiMoon, FiSun } from 'react-icons/fi';

import { type AppTheme, ThemeType } from '@/spec/setting';
import { CustomThemeDark, CustomThemeLight, CustomThemeSystem, DAISYUI_BUILTIN_THEMES } from '@/spec/theme_consts';

import { updateStartupTheme, useStartupTheme } from '@/hooks/use_startup_theme';
import { useTheme } from '@/hooks/use_theme_provider';

import { settingstoreAPI } from '@/apis/baseapi';

import { Dropdown, type DropdownItem } from '@/components/dropdown';

const isOtherThemeName = (n: string): boolean => DAISYUI_BUILTIN_THEMES.includes(n);
const toThemeType = (name: string): ThemeType => {
	if (name === CustomThemeLight) return ThemeType.Light;
	if (name === CustomThemeDark) return ThemeType.Dark;
	if (name === 'system') return ThemeType.System;
	return ThemeType.Other;
};

/* ———————————————————————————————————————— main selector —————————————————————————————————— */
export function ThemeSelector() {
	const [startupTheme, startupReady] = useStartupTheme();
	const { theme: providerTheme, setTheme } = useTheme();

	/* derived state */
	const current = useMemo(() => toThemeType(providerTheme), [providerTheme]);

	/* dropdown items never change -> memo once */
	const dropdownItems = useMemo(
		() => Object.fromEntries(DAISYUI_BUILTIN_THEMES.map(t => [t, { isEnabled: true }])) as Record<string, DropdownItem>,
		[]
	);

	const [otherName, setOtherName] = useState<string>(DAISYUI_BUILTIN_THEMES[0]);

	/* initialise “other” theme name once start-up theme is loaded */
	useEffect(() => {
		if (!startupReady || !startupTheme) return;

		if (startupTheme.type === ThemeType.Other && isOtherThemeName(startupTheme.name)) {
			setOtherName(startupTheme.name);
		}
	}, [startupReady, startupTheme]);

	/* ————————————————————————————— apply theme —————————————————————————————— */
	const applyTheme = useCallback(
		async (type: ThemeType, name: string) => {
			if (!startupReady) return; // protect against very early calls
			if (name === '') {
				console.error('[Theme] empty name recieved');
				return;
			}
			/* optimistic update */
			setTheme(name);

			const newTheme: AppTheme = {
				type: type,
				name: name,
			};

			try {
				await settingstoreAPI.setAppTheme(newTheme);
				updateStartupTheme(newTheme);
				console.log('[Theme] changed to', newTheme.type, newTheme.name);
			} catch (err) {
				console.error('[Theme] failed to persist, reverting', err);
				// fallback to last known persistent value
				if (startupTheme) {
					setTheme(startupTheme.name);
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
		<div className="flex items-center gap-6">
			<label className="flex cursor-pointer items-center gap-2">
				<input
					type="radio"
					className="radio radio-accent"
					checked={current === ThemeType.System}
					onChange={() => {
						applyTheme(ThemeType.System, CustomThemeSystem);
					}}
				/>
				<FiMonitor />
				<span className="text-sm">System</span>
			</label>
			<label className="flex cursor-pointer items-center gap-2">
				<input
					type="radio"
					className="radio radio-accent"
					checked={current === ThemeType.Light}
					onChange={() => {
						applyTheme(ThemeType.Light, CustomThemeLight);
					}}
				/>
				<FiSun />
				<span className="text-sm">Light</span>
			</label>

			<label className="flex cursor-pointer items-center gap-2">
				<input
					type="radio"
					className="radio radio-accent"
					checked={current === ThemeType.Dark}
					onChange={() => {
						applyTheme(ThemeType.Dark, CustomThemeDark);
					}}
				/>
				<FiMoon />
				<span className="text-sm">Dark</span>
			</label>

			<label className="flex cursor-pointer items-center gap-2">
				<input
					type="radio"
					className="radio radio-accent"
					checked={current === ThemeType.Other}
					onChange={() => applyTheme(ThemeType.Other, otherName)}
				/>
				<div className="w-50">
					<Dropdown<string>
						dropdownItems={dropdownItems}
						selectedKey={otherName}
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
}
