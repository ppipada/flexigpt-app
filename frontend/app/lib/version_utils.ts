export const DEFAULT_SEMVER = 'v1.0.0';

export function isSemverVersion(v: string): boolean {
	return /^(v)?\d+\.\d+\.\d+$/i.test(v.trim());
}

type ParsedSemver = {
	hasV: boolean;
	major: number;
	minor: number;
	patch: number;
};

function parseSemver(v?: string): ParsedSemver | null {
	const raw = (v ?? '').trim();
	const m = raw.match(/^(v)?(\d+)\.(\d+)\.(\d+)$/i);
	if (!m) return null;
	return {
		hasV: Boolean(m[1]),
		major: Number(m[2]),
		minor: Number(m[3]),
		patch: Number(m[4]),
	};
}

function compareSemver(a: ParsedSemver, b: ParsedSemver): number {
	if (a.major !== b.major) return a.major - b.major;
	if (a.minor !== b.minor) return a.minor - b.minor;
	return a.patch - b.patch;
}

/**
 * If `current` is semver (`v1.2.3` or `1.2.3`), returns next minor (`v1.3.0` / `1.3.0`).
 * IMPORTANT:
 * If `existingVersions` is provided, the suggestion is computed as:
 *   - same major as `current`
 *   - next minor after the MAX existing minor for that major
 * This prevents suggesting a version that already exists when creating a new version from an older base.
 * Otherwise returns `v1.0.0`.
 */
export function suggestNextMinorVersion(
	current?: string,
	existingVersions?: string[]
): { suggested: string; isSemver: boolean } {
	const cur = parseSemver(current);

	// If the current isn't semver, optionally fall back to max(existing) if possible.
	if (!cur) {
		const parsedExisting = (existingVersions ?? [])
			.map(v => parseSemver(v))
			.filter((p): p is ParsedSemver => p !== null)
			.sort(compareSemver);

		if (parsedExisting.length === 0) return { suggested: DEFAULT_SEMVER, isSemver: false };

		const max = parsedExisting[parsedExisting.length - 1];
		const prefix = max.hasV ? 'v' : '';
		return { suggested: `${prefix}${max.major}.${max.minor + 1}.0`, isSemver: true };
	}

	const parsedExistingSameMajor = (existingVersions ?? [])
		.map(v => parseSemver(v))
		.filter((p): p is ParsedSemver => p !== null)
		.filter(p => p.major === cur.major);

	// Prefer "v" prefix if current has it; otherwise adopt it if the existing set uses it.
	const prefix = cur.hasV || parsedExistingSameMajor.some(p => p.hasV) ? 'v' : '';

	// If we have existing versions for this major, suggest next minor after the max existing minor.
	if (parsedExistingSameMajor.length > 0) {
		const maxMinor = Math.max(...parsedExistingSameMajor.map(p => p.minor));
		return { suggested: `${prefix}${cur.major}.${maxMinor + 1}.0`, isSemver: true };
	}

	// Otherwise, simple next-minor from the current.
	return { suggested: `${prefix}${cur.major}.${cur.minor + 1}.0`, isSemver: true };
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
