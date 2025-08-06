export const CustomThemeLight = 'nordsnowstorm';
export const CustomThemeDark = 'nordpolarnight';

const DAISYUI_LIGHT_THEMES = [
	'acid',
	'autumn',
	'bumblebee',
	'caramellatte',
	'cmyk',
	'corporate',
	'cupcake',
	'cyberpunk',
	'emerald',
	'fantasy',
	'garden',
	'lemonade',
	'light',
	'lofi',
	'nord',
	'pastel',
	'retro',
	'silk',
	'valentine',
	'winter',
	'wireframe',
];

const DAISYUI_DARK_THEMES = [
	'abyss',
	'aqua',
	'black',
	'business',
	'coffee',
	'dark',
	'dim',
	'dracula',
	'forest',
	'halloween',
	'luxury',
	'night',
	'sunset',
	'synthwave',
];

export const ALL_DARK_THEMES = [CustomThemeDark, ...DAISYUI_DARK_THEMES];
// export const ALL_LIGHT_THEMES = [CustomThemeLight, ...DAISYUI_LIGHT_THEMES];

export const DAISYUI_BUILTIN_THEMES = [...DAISYUI_LIGHT_THEMES, ...DAISYUI_DARK_THEMES] as const;
