<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { toast } from '$lib/stores/toast';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Spinner from '$lib/components/ui/Spinner.svelte';
	import { onMount } from 'svelte';
	import { _ } from '$lib/i18n';

	interface EffectiveTemplate {
		messageType: string;
		subject: string;
		body: string;
		isCustomized: boolean;
		availableVariables: string[];
	}

	const eventId = $derived($page.params.eventId);

	let loading = $state(true);
	let templates: EffectiveTemplate[] = $state([]);
	let savingType: string | null = $state(null);
	let resettingType: string | null = $state(null);

	function labelFor(messageType: string): string {
		return $_(`events.messageTemplates.types.${messageType}`);
	}

	onMount(async () => {
		await load();
	});

	async function load() {
		loading = true;
		try {
			const result = await api.get<{ data: EffectiveTemplate[] }>(`/message-templates/event/${eventId}`);
			templates = result.data;
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.messageTemplates.loadError'));
		} finally {
			loading = false;
		}
	}

	async function save(tpl: EffectiveTemplate) {
		savingType = tpl.messageType;
		try {
			await api.put(`/message-templates/event/${eventId}/${tpl.messageType}`, {
				subject: tpl.subject,
				body: tpl.body
			});
			tpl.isCustomized = true;
			toast.success($_('events.messageTemplates.saveSuccess'));
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.messageTemplates.saveError'));
		} finally {
			savingType = null;
		}
	}

	async function resetToDefault(messageType: string) {
		resettingType = messageType;
		try {
			await api.delete(`/message-templates/event/${eventId}/${messageType}`);
			toast.success($_('events.messageTemplates.resetSuccess'));
			await load();
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.messageTemplates.resetError'));
		} finally {
			resettingType = null;
		}
	}
</script>

<svelte:head>
	<title>{$_('events.messageTemplates.pageTitle')}</title>
</svelte:head>

<AppShell>
	<div class="max-w-3xl mx-auto">
		<div class="mb-8">
			<a href="/events/{eventId}/edit" class="text-sm text-primary hover:text-primary-hover">&larr; {$_('events.edit.backToEvent')}</a>
			<h1 class="mt-2 text-2xl font-bold font-display text-neutral-900">{$_('events.messageTemplates.heading')}</h1>
			<p class="mt-1 text-sm text-neutral-500">{$_('events.messageTemplates.pageDescription')}</p>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<Spinner size="lg" class="text-primary" />
			</div>
		{:else}
			<div class="space-y-6">
				{#each templates as tpl (tpl.messageType)}
					<Card>
						<div class="flex items-center justify-between mb-4">
							<h2 class="text-base font-semibold text-neutral-900">{labelFor(tpl.messageType)}</h2>
							{#if tpl.isCustomized}
								<span class="text-xs font-medium text-primary bg-primary-light rounded-full px-2 py-0.5">
									{$_('events.messageTemplates.customizedBadge')}
								</span>
							{:else}
								<span class="text-xs font-medium text-neutral-500 bg-neutral-100 rounded-full px-2 py-0.5">
									{$_('events.messageTemplates.defaultBadge')}
								</span>
							{/if}
						</div>

						<div class="space-y-4">
							<Input
								label={$_('events.messageTemplates.subjectLabel')}
								name="subject-{tpl.messageType}"
								bind:value={tpl.subject}
							/>
							<Textarea
								label={$_('events.messageTemplates.bodyLabel')}
								name="body-{tpl.messageType}"
								bind:value={tpl.body}
								rows={8}
							/>

							{#if tpl.availableVariables?.length}
								<div>
									<p class="text-xs font-medium text-neutral-500 mb-2">{$_('events.messageTemplates.variablesHint')}</p>
									<div class="flex flex-wrap gap-2">
										{#each tpl.availableVariables as v (v)}
											<code class="text-xs bg-neutral-100 text-neutral-700 rounded px-2 py-1">{'{' + v + '}'}</code>
										{/each}
									</div>
								</div>
							{/if}

							<div class="flex items-center justify-end gap-3 pt-2 border-t border-neutral-200">
								{#if tpl.isCustomized}
									<Button
										variant="outline"
										loading={resettingType === tpl.messageType}
										onclick={() => resetToDefault(tpl.messageType)}
									>
										{$_('events.messageTemplates.resetAction')}
									</Button>
								{/if}
								<Button loading={savingType === tpl.messageType} onclick={() => save(tpl)}>
									{$_('events.messageTemplates.saveAction')}
								</Button>
							</div>
						</div>
					</Card>
				{/each}
			</div>
		{/if}
	</div>
</AppShell>
