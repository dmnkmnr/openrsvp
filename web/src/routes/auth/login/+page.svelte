<script lang="ts">
	import { api } from '$lib/api/client';
	import { toast } from '$lib/stores/toast';
	import { isValidEmail } from '$lib/utils/validation';
	import { _, locale, SUPPORTED_LOCALES } from '$lib/i18n';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Select from '$lib/components/ui/Select.svelte';

	let email = $state('');
	let loading = $state(false);
	let sent = $state(false);
	let emailError = $state('');
	let language = $state($locale ?? 'en');

	const languageNames: Record<string, string> = { en: 'English', de: 'Deutsch' };
	const languageOptions = SUPPORTED_LOCALES.map((code) => ({ value: code, label: languageNames[code] ?? code }));

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		emailError = '';

		if (!email.trim()) {
			emailError = $_('auth.login.emailRequired');
			return;
		}

		if (!isValidEmail(email)) {
			emailError = $_('auth.login.emailInvalid');
			return;
		}

		loading = true;
		try {
			await api.post('/auth/magic-link', { email, language });
			sent = true;
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('auth.login.sendFailed'));
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('auth.login.pageTitle')}</title>
</svelte:head>

<div class="min-h-screen flex items-center justify-center px-4">
	<div class="w-full max-w-md">
		<div class="text-center mb-8">
			<a href="/" class="text-2xl font-bold text-primary">OpenRSVP</a>
			<h1 class="font-display mt-4 text-2xl font-semibold text-neutral-900">{$_('auth.login.heading')}</h1>
			<p class="mt-2 text-neutral-600">{$_('auth.login.subheading')}</p>
		</div>

		<div class="bg-surface rounded-lg shadow-sm border border-neutral-200 p-8">
			{#if sent}
				<!-- Success state -->
				<div class="text-center">
					<div
						class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-success-light mb-4"
					>
						<svg
							class="h-6 w-6 text-success"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"
							/>
						</svg>
					</div>
					<h2 class="text-lg font-semibold text-neutral-900 mb-2">{$_('auth.login.checkEmailHeading')}</h2>
					<p class="text-sm text-neutral-600 mb-4">
						{$_('auth.login.checkEmailBody', { values: { email } })}
					</p>
					<p class="text-xs text-neutral-500 mb-6">
						{$_('auth.login.notReceived')}
					</p>
					<Button
						variant="outline"
						onclick={() => {
							sent = false;
							email = '';
						}}
					>
						{$_('auth.login.tryDifferentEmail')}
					</Button>
				</div>
			{:else}
				<!-- Login form -->
				<form onsubmit={handleSubmit} class="space-y-6">
					<Input
						label={$_('auth.login.emailLabel')}
						name="email"
						type="email"
						bind:value={email}
						placeholder={$_('auth.login.emailPlaceholder')}
						error={emailError}
						required
					/>

					<div>
						<Select
							label={$_('auth.login.languageLabel')}
							name="language"
							bind:value={language}
							options={languageOptions}
						/>
						<p class="mt-1.5 text-xs text-neutral-400">{$_('auth.login.languageHelper')}</p>
					</div>

					<Button type="submit" {loading} class="w-full">
						{loading ? $_('auth.login.sending') : $_('auth.login.sendMagicLink')}
					</Button>
				</form>

				<div class="mt-6 text-center">
					<a href="/" class="text-sm text-primary hover:text-primary-hover">
						{$_('auth.login.backToHome')}
					</a>
				</div>
			{/if}
		</div>
	</div>
</div>
