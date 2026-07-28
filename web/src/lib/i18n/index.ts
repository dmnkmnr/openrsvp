import { init, register, locale, isLoading, waitLocale } from 'svelte-i18n';

export const SUPPORTED_LOCALES = ['en', 'en-US', 'de'] as const;
export type SupportedLocale = (typeof SUPPORTED_LOCALES)[number];
export const DEFAULT_LOCALE: SupportedLocale = 'en';

const STORAGE_KEY = 'locale';

register('en', () => import('./locales/en.json'));
register('en-US', () => import('./locales/en-US.json'));
register('de', () => import('./locales/de.json'));

function isSupported(value: string | null | undefined): value is SupportedLocale {
	return !!value && (SUPPORTED_LOCALES as readonly string[]).includes(value);
}

function detectInitialLocale(): SupportedLocale {
	if (typeof window === 'undefined') return DEFAULT_LOCALE;

	const stored = localStorage.getItem(STORAGE_KEY);
	if (isSupported(stored)) return stored;

	for (const lang of navigator.languages ?? [navigator.language]) {
		// Check the full tag first (e.g. "en-US") before falling back to the
		// primary subtag (e.g. "en"), so US-English browsers land on the
		// American locale rather than the British default.
		if (isSupported(lang)) return lang;
		const primary = lang?.split('-')[0]?.toLowerCase();
		if (isSupported(primary)) return primary;
	}

	return DEFAULT_LOCALE;
}

init({
	fallbackLocale: DEFAULT_LOCALE,
	initialLocale: detectInitialLocale()
});

if (typeof window !== 'undefined') {
	locale.subscribe((value) => {
		if (isSupported(value)) localStorage.setItem(STORAGE_KEY, value);
	});
}

export { locale, waitLocale };
export { isLoading as i18nLoading };
export { _ } from 'svelte-i18n';
