<script lang="ts">
	import '../app.css';
	import { currentUser, isLoading } from '$lib/stores/auth';
	import { i18nLoading, waitLocale, locale, SUPPORTED_LOCALES } from '$lib/i18n';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import Toast from '$lib/components/ui/Toast.svelte';

	onMount(async () => {
		waitLocale();
		try {
			const user = await api.get<import('$lib/types').Organizer>('/auth/me');
			$currentUser = user;

			// Follow the organizer's saved language preference so the UI
			// matches their account setting across devices/sessions, not just
			// whatever this browser previously detected/stored.
			if (user.language && (SUPPORTED_LOCALES as readonly string[]).includes(user.language)) {
				$locale = user.language;
			}

			// Auto-save browser timezone to profile if not set yet.
			if (!user.timezone) {
				const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
				if (tz) {
					api.patch<import('$lib/types').Organizer>('/auth/me', { timezone: tz })
						.then((updated) => { $currentUser = updated; })
						.catch(() => {});
				}
			}
		} catch {
			$currentUser = null;
		} finally {
			$isLoading = false;
		}
	});

	let { children } = $props();
</script>

<div class="min-h-screen bg-neutral-50">
	{#if $i18nLoading}
		<div class="flex min-h-screen items-center justify-center">
			<div class="loading-spinner"></div>
		</div>
	{:else}
		{@render children()}
	{/if}
	<Toast />
</div>
