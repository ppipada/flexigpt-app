export const DEFAULT_SEMVER = 'v1.0.0';

export function isSemverVersion(v: string): boolean {
	return /^(v)?\d+\.\d+\.\d+$/i.test(v.trim());
}

/**
 * If `current` is semver (`v1.2.3` or `1.2.3`), returns next minor (`v1.3.0` / `1.3.0`).
 * Otherwise returns `v1.0.0`.
 */
export function suggestNextMinorVersion(current?: string): { suggested: string; isSemver: boolean } {
	const raw = (current ?? '').trim();
	const m = raw.match(/^(v)?(\d+)\.(\d+)\.(\d+)$/i);
	if (!m) return { suggested: DEFAULT_SEMVER, isSemver: false };
	const prefix = m[1] ? 'v' : '';
	const major = Number(m[2]);
	const minor = Number(m[3]);
	return { suggested: `${prefix}${major}.${minor + 1}.0`, isSemver: true };
}

/**
 * @public
 */
export function validateVersion(version: string): string | undefined {
	const trimmed = version.trim();
	if (!trimmed) return 'Version is required.';
	if (trimmed.length > 64) return 'Version must be at most 64 characters.';
	if (!/^[a-zA-Z0-9.-]+$/.test(trimmed)) {
		return 'Version may only contain letters, numbers, "-", and ".".';
	}
	return undefined;
}
