<script lang="ts">
	import '../app.css';
	import { currentUser, isLoading } from '$lib/stores/auth';
	import { i18nLoading, waitLocale } from '$lib/i18n';
	import { api } from '$lib/api/client';
	import { onMount } from 'svelte';
	import Toast from '$lib/components/ui/Toast.svelte';

	onMount(async () => {
		waitLocale();
		try {
			const user = await api.get<import('$lib/types').Organizer>('/auth/me');
			$currentUser = user;

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
