/* ────────────────────────────────────────────────────────────────
   src/lib/theme_provider.tsx
   ──────────────────────────────────────────────────────────────── */
import type { FC, JSX } from 'react';
import React, { useState } from 'react';

import { ThemeProvider, useTheme } from 'next-themes';
import { FiMonitor, FiMoon, FiSun } from 'react-icons/fi';

import { type AppTheme, ThemeType } from '@/spec/setting';

import { settingstoreAPI } from '@/apis/baseapi';
import { getStartupThemeSync, updateStartupTheme } from '@/apis/builtin_theme_cache';

/* ------------------------------------------------------------------
			small helpers
			------------------------------------------------------------------ */

const toNextThemesName = (t: ThemeType | AppTheme): string => {
	const type = typeof t === 'string' ? t : t.type;
	switch (type) {
		case ThemeType.System:
			return 'system';
		case ThemeType.Light:
			return 'light';
		case ThemeType.Dark:
			return 'dark';
		default:
			return typeof t === 'string' ? t : t.name; // ThemeType.Other
	}
};

const toThemeType = (name: string): ThemeType => {
	switch (name) {
		case 'light':
			return ThemeType.Light;
		case 'dark':
			return ThemeType.Dark;
		case 'system':
			return ThemeType.System;
		default:
			return ThemeType.Other;
	}
};

/* ------------------------------------------------------------------
			1.  Provider that must wrap the whole app
			------------------------------------------------------------------ */

export const ThemeSwitchProvider: FC<{ children: React.ReactNode }> = ({ children }) => {
	// guaranteed to be initialised by rootLoader()
	const startupTheme = getStartupThemeSync();

	return (
		<ThemeProvider
			attribute="data-theme"
			/* daisyUI palette map */
			value={{ light: 'nordsnowstorm', dark: 'nordpolarnight' }}
			/* we supply the default ourselves – no automatic system switch */
			defaultTheme={toNextThemesName(startupTheme)}
			enableSystem={false}
		>
			{children}
		</ThemeProvider>
	);
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
		{icon} <span className="text-sm">{label}</span>
	</label>
);

export const ThemeSelector: FC = () => {
	const { setTheme, theme } = useTheme();
	const [current, setCurrent] = useState<ThemeType>(toThemeType(theme ?? toNextThemesName(getStartupThemeSync())));

	/* when user picks a different option -------------------------------- */
	const handleChange = async (type: ThemeType) => {
		setCurrent(type);

		/* 1. apply visually right away */
		setTheme(toNextThemesName(type));

		/* 2. update session-wide singleton cache */
		const newTheme: AppTheme = { type, name: type };
		updateStartupTheme(newTheme);

		/* 3. persist to backend */
		await settingstoreAPI.setAppTheme(newTheme);
	};

	return (
		<div className="flex gap-6">
			<ThemeRadio
				label="System"
				value={ThemeType.System}
				icon={<FiMonitor />}
				current={current}
				onChange={handleChange}
			/>
			<ThemeRadio label="Light" value={ThemeType.Light} icon={<FiSun />} current={current} onChange={handleChange} />
			<ThemeRadio label="Dark" value={ThemeType.Dark} icon={<FiMoon />} current={current} onChange={handleChange} />
		</div>
	);
};
