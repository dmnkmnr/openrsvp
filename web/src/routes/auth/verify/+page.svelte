<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { currentUser } from '$lib/stores/auth';
	import { toast } from '$lib/stores/toast';
	import Spinner from '$lib/components/ui/Spinner.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import type { Organizer } from '$lib/types';
	import { onMount } from 'svelte';
	import { _ } from '$lib/i18n';

	let verifying = $state(true);
	let error = $state('');

	onMount(async () => {
		const token = $page.url.searchParams.get('token');

		if (!token) {
			error = $_('auth.verify.noTokenError');
			verifying = false;
			return;
		}

		// Strip the raw token from the URL/history so it does not linger in
		// browser history. Use the native history API directly: users arrive
		// here from a fresh page load (the email link), where SvelteKit's
		// router may not be initialized yet. In that state $app/navigation's
		// replaceState throws, which previously aborted verification before
		// the request was ever sent and left the spinner hanging forever.
		try {
			window.history.replaceState(window.history.state, '', '/auth/verify');
		} catch {
			// Non-fatal: verification must proceed even if URL cleanup fails.
		}

		try {
			const result = await api.post<{ token: string; organizer: Organizer }>('/auth/verify', { token });
			$currentUser = result.organizer;
			toast.success($_('auth.verify.verifiedSuccess'));
			goto('/events');
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			error = apiErr.message || $_('auth.verify.verificationFailedError');
			verifying = false;
		}
	});
</script>

<svelte:head>
	<title>{$_('auth.verify.pageTitle')}</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center px-4">
	<div class="w-full max-w-md text-center">
		<a href="/" class="text-2xl font-bold text-primary">OpenRSVP</a>

		{#if verifying}
			<h1 class="font-display mt-4 text-2xl font-semibold text-neutral-900">{$_('auth.verify.verifyingHeading')}</h1>
			<div class="mt-6 flex flex-col items-center">
				<Spinner size="md" class="text-primary" />
				<p class="mt-4 text-neutral-600">{$_('auth.verify.verifyingBody')}</p>
			</div>
		{:else if error}
			<div class="mt-6">
				<div
					class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-error-light mb-4"
				>
					<svg class="h-6 w-6 text-error" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
				</div>
				<h2 class="font-display text-lg font-semibold text-neutral-900 mb-2">{$_('auth.verify.failedHeading')}</h2>
				<p class="text-sm text-neutral-600 mb-6">{error}</p>
				<Button href="/auth/login">{$_('auth.verify.tryAgain')}</Button>
			</div>
		{/if}
	</div>
</div>
