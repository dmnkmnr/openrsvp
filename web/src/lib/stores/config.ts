import { writable } from 'svelte/store';
import { api } from '$lib/api/client';

export const smsEnabled = writable(false);
export const supportedLanguages = writable<string[]>(['en']);

let loaded = false;

export async function loadAppConfig() {
	if (loaded) return;
	try {
		const result = await api.get<{ data: { smsEnabled: boolean; supportedLanguages: string[] } }>('/config');
		smsEnabled.set(result.data.smsEnabled);
		if (result.data.supportedLanguages?.length) {
			supportedLanguages.set(result.data.supportedLanguages);
		}
		loaded = true;
	} catch {
		smsEnabled.set(false);
	}
}
