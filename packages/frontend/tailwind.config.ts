import type { Config } from 'tailwindcss';

const config: Config = {
	darkMode: 'class',
	content: ['./app/**/*.{js,ts,jsx,tsx,mdx}'],
	theme: {},
	plugins: [require('daisyui')],
	daisyui: {
		themes: ['nord', 'dracula'],
		darkTheme: 'dracula',
		prefix: '',
	},
};

export default config;
