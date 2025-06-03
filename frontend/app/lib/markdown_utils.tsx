// LaTeX processing function
const testLatexRegex = /[$\\]/;
const containsLatexRegex = /\\\(.*?\\\)|\\\[.*?\\\]|\$.*?\$|\\begin\{equation\}.*?\\end\{equation\}/;
const inlineLatex = new RegExp(/\\\((.+?)\\\)/, 'g');
const blockLatex = new RegExp(/\\\[(.*?[^\\])\\\]/, 'gs');

export function SanitizeLaTeX(content: string) {
	if (!testLatexRegex.test(content)) {
		return content;
	}
	let processedContent = content.replace(/(\$)(?=\s?\d)/g, '\\$');

	if (!containsLatexRegex.test(processedContent)) {
		return processedContent;
	}

	processedContent = processedContent
		.replace(inlineLatex, (match: string, equation: string) => `$${equation}$`)
		.replace(blockLatex, (match: string, equation: string) => `$$${equation}$$`);

	return processedContent;
}

interface LanguageMap {
	[key: string]: { extension: string; mimeType: string };
}

export const ProgrammingLanguages: LanguageMap = {
	javascript: { extension: '.js', mimeType: 'application/javascript' },
	python: { extension: '.py', mimeType: 'text/x-python' },
	java: { extension: '.java', mimeType: 'text/x-java-source' },
	c: { extension: '.c', mimeType: 'text/x-c' },
	cpp: { extension: '.cpp', mimeType: 'text/x-c++' },
	'c++': { extension: '.cpp', mimeType: 'text/x-c++' },
	'c#': { extension: '.cs', mimeType: 'text/x-csharp' },
	ruby: { extension: '.rb', mimeType: 'text/x-ruby' },
	php: { extension: '.php', mimeType: 'application/php' },
	swift: { extension: '.swift', mimeType: 'text/x-swift' },
	'objective-c': { extension: '.m', mimeType: 'text/x-objectivec' },
	kotlin: { extension: '.kt', mimeType: 'text/x-kotlin' },
	typescript: { extension: '.ts', mimeType: 'application/typescript' },
	go: { extension: '.go', mimeType: 'text/x-go' },
	perl: { extension: '.pl', mimeType: 'text/x-perl' },
	rust: { extension: '.rs', mimeType: 'text/x-rustsrc' },
	scala: { extension: '.scala', mimeType: 'text/x-scala' },
	haskell: { extension: '.hs', mimeType: 'text/x-haskell' },
	lua: { extension: '.lua', mimeType: 'text/x-lua' },
	shell: { extension: '.sh', mimeType: 'application/x-sh' },
	sql: { extension: '.sql', mimeType: 'application/sql' },
	html: { extension: '.html', mimeType: 'text/html' },
	css: { extension: '.css', mimeType: 'text/css' },
	json: { extension: '.json', mimeType: 'application/json' },
	toml: { extension: '.toml', mimeType: 'application/toml' },
	yaml: { extension: '.yaml', mimeType: 'application/yaml' },
	dart: { extension: '.dart', mimeType: 'application/dart' },
	// Add more file extensions and MIME types here
};

// Determine the file extension from the MIME type
export const MimeTypeMap: { [key: string]: string } = {
	'image/png': '.png',
	'image/jpeg': '.jpg',
	'image/gif': '.gif',
	'application/pdf': '.pdf',
	// Add more MIME types and extensions as needed
};

export const SupportedProgrammingLanguages = Object.keys(ProgrammingLanguages);
export type SupportedLanguage = (typeof SupportedProgrammingLanguages)[number];

// Fallback check for unsupported languages
export const isLanguageSupported = (lang: string): lang is SupportedLanguage => {
	return SupportedProgrammingLanguages.includes(lang);
};
