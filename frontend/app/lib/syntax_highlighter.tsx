import bash from '@shikijs/langs/bash';
import c from '@shikijs/langs/c';
import cpp from '@shikijs/langs/cpp';
import csharp from '@shikijs/langs/csharp';
import css from '@shikijs/langs/css';
import dart from '@shikijs/langs/dart';
import go from '@shikijs/langs/go';
import haskell from '@shikijs/langs/haskell';
import html from '@shikijs/langs/html';
import java from '@shikijs/langs/java';
import javascript from '@shikijs/langs/javascript';
import json from '@shikijs/langs/json';
import jsx from '@shikijs/langs/jsx';
import kotlin from '@shikijs/langs/kotlin';
import lua from '@shikijs/langs/lua';
import objectivec from '@shikijs/langs/objective-c';
import perl from '@shikijs/langs/perl';
import php from '@shikijs/langs/php';
import python from '@shikijs/langs/python';
import ruby from '@shikijs/langs/ruby';
import rust from '@shikijs/langs/rust';
import scala from '@shikijs/langs/scala';
import shell from '@shikijs/langs/shell';
import sql from '@shikijs/langs/sql';
import swift from '@shikijs/langs/swift';
import toml from '@shikijs/langs/toml';
import tsx from '@shikijs/langs/tsx';
import typescript from '@shikijs/langs/typescript';
import yaml from '@shikijs/langs/yaml';
import monokai from '@shikijs/themes/monokai';
import { createHighlighterCore } from 'shiki/core';
import { createOnigurumaEngine } from 'shiki/engine/oniguruma';
import wasm from 'shiki/wasm';

export const highlighter = await createHighlighterCore({
	themes: [monokai],
	langs: [
		bash,
		c,
		cpp,
		csharp,
		css,
		dart,
		go,
		haskell,
		html,
		java,
		javascript,
		jsx,
		json,
		kotlin,
		lua,
		objectivec,
		perl,
		php,
		python,
		ruby,
		rust,
		scala,
		shell,
		sql,
		swift,
		toml,
		typescript,
		tsx,
		yaml,
	],
	// `shiki/wasm` contains the wasm binary inlined as base64 string.
	engine: createOnigurumaEngine(wasm),
});
