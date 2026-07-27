<script lang="ts">
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { toast } from '$lib/stores/toast';
	import { currentUser, isLoading } from '$lib/stores/auth';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Spinner from '$lib/components/ui/Spinner.svelte';
	import { _, locale, SUPPORTED_LOCALES } from '$lib/i18n';
	import type { Organizer } from '$lib/types';

	// Mirror the events layout guard: redirect to login once auth has loaded
	// and there's no current user.
	$effect(() => {
		if (!$isLoading && !$currentUser) {
			goto('/auth/login');
		}
	});

	let exporting = $state(false);

	let language = $state('en');
	let savingLanguage = $state(false);
	$effect(() => {
		if ($currentUser) language = $currentUser.language || 'en';
	});

	const languageNames: Record<string, string> = { en: 'English', de: 'Deutsch' };
	const languageOptions = SUPPORTED_LOCALES.map((code) => ({ value: code, label: languageNames[code] ?? code }));

	async function handleSaveLanguage() {
		savingLanguage = true;
		try {
			const updated = await api.patch<Organizer>('/auth/me', { language });
			$currentUser = updated;
			$locale = language;
			toast.success($_('account.preferencesSaved'));
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('account.preferencesError'));
		} finally {
			savingLanguage = false;
		}
	}

	let deleteModalOpen = $state(false);
	let deleteConfirmText = $state('');
	let deleting = $state(false);

	const canDelete = $derived(deleteConfirmText.trim() === 'DELETE');

	async function handleExport() {
		exporting = true;
		try {
			await api.download('/auth/me/export', 'openrsvp-export.json');
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('account.exportError'));
		} finally {
			exporting = false;
		}
	}

	function openDeleteModal() {
		deleteConfirmText = '';
		deleteModalOpen = true;
	}

	async function handleDelete() {
		if (!canDelete) return;

		deleting = true;
		try {
			await api.delete('/auth/me');
			deleteModalOpen = false;
			$currentUser = null;
			toast.success($_('account.deleteSuccess'));
			goto('/');
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('account.deleteError'));
		} finally {
			deleting = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('account.pageTitle')}</title>
</svelte:head>

{#if $isLoading}
	<div class="flex items-center justify-center min-h-screen">
		<Spinner size="lg" class="text-primary" />
	</div>
{:else if $currentUser}
	<AppShell>
		<div class="max-w-3xl mx-auto">
			<div class="mb-8">
				<h1 class="text-2xl font-bold font-display text-neutral-900">{$_('account.heading')}</h1>
				<p class="mt-1 text-sm text-neutral-500">{$currentUser.email}</p>
			</div>

			<!-- Section 0: Preferences -->
			<Card>
				<h2 class="text-lg font-display font-semibold text-neutral-900 mb-1">{$_('account.preferencesTitle')}</h2>
				<p class="text-sm text-neutral-600 mb-4">{$_('account.preferencesBody')}</p>
				<div class="flex flex-col sm:flex-row sm:items-end gap-4">
					<div class="flex-1 max-w-xs">
						<Select
							label={$_('account.languageLabel')}
							name="language"
							bind:value={language}
							options={languageOptions}
						/>
					</div>
					<Button loading={savingLanguage} onclick={handleSaveLanguage}>
						{$_('account.savePreferences')}
					</Button>
				</div>
			</Card>

			<!-- Section 1: Your data -->
			<Card class="mt-6">
				<div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
					<div class="sm:pr-8">
						<h2 class="text-lg font-display font-semibold text-neutral-900">{$_('account.dataTitle')}</h2>
						<p class="mt-1 text-sm text-neutral-600">
							{$_('account.dataBody')}
						</p>
					</div>
					<div class="flex-shrink-0">
						<Button variant="outline" loading={exporting} onclick={handleExport}>
							{$_('account.exportData')}
						</Button>
					</div>
				</div>
			</Card>

			<!-- Section 2: Danger zone -->
			<Card class="mt-6 border-error-light">
				<div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
					<div class="sm:pr-8">
						<h2 class="text-lg font-display font-semibold text-error">{$_('account.dangerZoneTitle')}</h2>
						<p class="mt-1 text-sm text-neutral-600">
							{$_('account.dangerZoneBody')}
						</p>
					</div>
					<div class="flex-shrink-0">
						<Button variant="danger" onclick={openDeleteModal}>{$_('account.deleteAccount')}</Button>
					</div>
				</div>
			</Card>
		</div>
	</AppShell>

	<Modal bind:open={deleteModalOpen} title={$_('account.deleteAccount')}>
		<div class="space-y-4">
			<div class="rounded-lg bg-error-light border border-error px-4 py-3 text-sm text-error flex items-start gap-2">
				<svg class="h-4 w-4 text-error mt-0.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
				</svg>
				<span>
					{$_('account.deleteModalWarningPrefix')} <strong>{$_('account.deleteModalWarningAll')}</strong> {$_('account.deleteModalWarningSuffix')}
				</span>
			</div>
			<p class="text-sm text-neutral-600">
				{$_('account.deleteConfirmPrompt')}
			</p>
			<Input
				name="deleteConfirm"
				bind:value={deleteConfirmText}
				placeholder="DELETE"
			/>
		</div>

		{#snippet actions()}
			<Button variant="outline" onclick={() => (deleteModalOpen = false)} disabled={deleting}>
				{$_('common.cancel')}
			</Button>
			<Button variant="danger" loading={deleting} disabled={!canDelete} onclick={handleDelete}>
				{$_('account.deleteMyAccount')}
			</Button>
		{/snippet}
	</Modal>
{/if}
