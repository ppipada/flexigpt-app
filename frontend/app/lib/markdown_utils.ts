// LaTeX processing function
const testLatexRegex = /[$\\]/;
const containsLatexRegex = /\\\(.*?\\\)|\\\[.*?\\\]|\$.*?\$|\\begin\{equation\}.*?\\end\{equation\}/;
const inlineLatex = new RegExp(/\\\((.+?)\\\)/, 'g');
const blockLatex = new RegExp(/\\\[(.*?[^\\])\\\]/, 'gs');

function SanitizeLaTeX(content: string) {
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

// Apply SanitizeLaTeX only to non-fenced regions of markdown.
export function SanitizeLaTeXOutsideFences(md: string) {
	if (!testLatexRegex.test(md)) return md;

	const lines = md.split('\n');
	const out: string[] = [];
	let outside: string[] = [];
	let fence: string | null = null; // e.g. ``` or ~~~ (or longer)

	const flushOutside = () => {
		if (outside.length === 0) return;
		out.push(SanitizeLaTeX(outside.join('\n')));
		outside = [];
	};

	for (const line of lines) {
		const m = line.match(/^([`~]{3,})/);
		if (!fence) {
			if (m) {
				flushOutside();
				fence = m[1];
				out.push(line);
			} else {
				outside.push(line);
			}
		} else {
			out.push(line);
			if (line.startsWith(fence)) {
				fence = null;
			}
		}
	}
	flushOutside();
	return out.join('\n');
}

interface LanguageMap {
	[key: string]: { extension: string; mimeType: string };
}

export const ProgrammingLanguages: LanguageMap = {
	javascript: { extension: '.js', mimeType: 'text/javascript' },
	jsx: { extension: '.jsx', mimeType: 'text/javascript' },
	python: { extension: '.py', mimeType: 'text/x-python' },
	java: { extension: '.java', mimeType: 'text/x-java-source' },
	c: { extension: '.c', mimeType: 'text/x-c' },
	cpp: { extension: '.cpp', mimeType: 'text/x-c++' },
	csharp: { extension: '.cs', mimeType: 'text/x-csharp' },
	ruby: { extension: '.rb', mimeType: 'text/x-ruby' },
	php: { extension: '.php', mimeType: 'application/php' },
	swift: { extension: '.swift', mimeType: 'text/x-swift' },
	'objective-c': { extension: '.m', mimeType: 'text/x-objectivec' },
	kotlin: { extension: '.kt', mimeType: 'text/x-kotlin' },
	typescript: { extension: '.ts', mimeType: 'application/typescript' },
	tsx: { extension: '.tsx', mimeType: 'application/typescript' },
	go: { extension: '.go', mimeType: 'text/x-go' },
	perl: { extension: '.pl', mimeType: 'text/x-perl' },
	rust: { extension: '.rs', mimeType: 'text/x-rustsrc' },
	scala: { extension: '.scala', mimeType: 'text/x-scala' },
	haskell: { extension: '.hs', mimeType: 'text/x-haskell' },
	lua: { extension: '.lua', mimeType: 'text/x-lua' },
	shell: { extension: '.sh', mimeType: 'application/x-sh' },
	bash: { extension: '.sh', mimeType: 'text/x-bash' },
	sql: { extension: '.sql', mimeType: 'application/sql' },
	html: { extension: '.html', mimeType: 'text/html' },
	css: { extension: '.css', mimeType: 'text/css' },
	json: { extension: '.json', mimeType: 'application/json' },
	toml: { extension: '.toml', mimeType: 'application/toml' },
	yaml: { extension: '.yaml', mimeType: 'application/yaml' },
	dart: { extension: '.dart', mimeType: 'application/dart' },
};

// Determine the file extension from the MIME type
export const MimeTypeMap: { [key: string]: string } = {
	'image/png': '.png',
	'image/jpeg': '.jpg',
	'image/gif': '.gif',
	'application/pdf': '.pdf',
	// Add more MIME types and extensions as needed
};
