<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { toast } from '$lib/stores/toast';
	import type { Webhook, WebhookWithSecret, WebhookDelivery, ApiError } from '$lib/types';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Spinner from '$lib/components/ui/Spinner.svelte';
	import { _ } from '$lib/i18n';

	const eventId = $derived($page.params.eventId);

	const EVENT_TYPES = [
		'rsvp.created',
		'rsvp.updated',
		'rsvp.cancelled',
		'attendee.checked_in',
		'event.published',
		'event.cancelled'
	];

	// State
	let loading = $state(true);
	let webhooks = $state<Webhook[]>([]);
	let showCreateForm = $state(false);
	let creating = $state(false);

	// Create form
	let newUrl = $state('');
	let newDescription = $state('');
	let newEventTypes = $state<string[]>([...EVENT_TYPES]);

	// Edit state
	let editingWebhook = $state<Webhook | null>(null);
	let editUrl = $state('');
	let editDescription = $state('');
	let editEventTypes = $state<string[]>([]);
	let editEnabled = $state(true);
	let saving = $state(false);

	// Secret display
	let showSecretModal = $state(false);
	let displayedSecret = $state('');
	let secretCopied = $state(false);

	// Delete confirmation
	let deleteTarget = $state<Webhook | null>(null);
	let showDeleteModal = $state(false);

	// Rotate secret confirmation
	let rotateTarget = $state<Webhook | null>(null);
	let showRotateModal = $state(false);
	let rotating = $state(false);

	// Deliveries
	let expandedWebhookId = $state('');
	let deliveries = $state<WebhookDelivery[]>([]);
	let loadingDeliveries = $state(false);

	// Test
	let testingWebhookId = $state('');

	onMount(async () => {
		try {
			const result = await api.get<{ data: Webhook[] }>(`/webhooks/event/${eventId}`);
			webhooks = result.data;
		} catch (err) {
			const apiErr = err as ApiError;
			toast.error(apiErr.message || $_('events.webhooks.loadError'));
		} finally {
			loading = false;
		}
	});

	function toggleEventType(type: string, list: string[]): string[] {
		if (list.includes(type)) {
			return list.filter(t => t !== type);
		}
		return [...list, type];
	}

	async function createWebhook() {
		if (!newUrl.trim()) {
			toast.error($_('events.webhooks.urlRequired'));
			return;
		}
		if (newEventTypes.length === 0) {
			toast.error($_('events.webhooks.selectEventType'));
			return;
		}

		creating = true;
		try {
			const result = await api.post<{ data: WebhookWithSecret }>(`/webhooks/event/${eventId}`, {
				url: newUrl.trim(),
				description: newDescription.trim(),
				eventTypes: newEventTypes
			});
			webhooks = [...webhooks, result.data];
			displayedSecret = result.data.secret;
			showSecretModal = true;
			showCreateForm = false;
			newUrl = '';
			newDescription = '';
			newEventTypes = [...EVENT_TYPES];
			toast.success($_('events.webhooks.createSuccess'));
		} catch (err) {
			const apiErr = err as ApiError;
			toast.error(apiErr.message || $_('events.webhooks.createError'));
		} finally {
			creating = false;
		}
	}

	function startEdit(webhook: Webhook) {
		editingWebhook = webhook;
		editUrl = webhook.url;
		editDescription = webhook.description;
		editEventTypes = [...webhook.eventTypes];
		editEnabled = webhook.enabled;
	}

	function cancelEdit() {
		editingWebhook = null;
	}

	async function saveWebhook() {
		if (!editingWebhook) return;
		if (!editUrl.trim()) {
			toast.error($_('events.webhooks.urlRequired'));
			return;
		}
		if (editEventTypes.length === 0) {
			toast.error($_('events.webhooks.selectEventType'));
			return;
		}

		saving = true;
		try {
			const result = await api.put<{ data: Webhook }>(`/webhooks/${editingWebhook.id}`, {
				url: editUrl.trim(),
				description: editDescription.trim(),
				eventTypes: editEventTypes,
				enabled: editEnabled
			});
			webhooks = webhooks.map(w => w.id === editingWebhook!.id ? result.data : w);
			editingWebhook = null;
			toast.success($_('events.webhooks.updateSuccess'));
		} catch (err) {
			const apiErr = err as ApiError;
			toast.error(apiErr.message || $_('events.webhooks.updateError'));
		} finally {
			saving = false;
		}
	}

	async function deleteWebhook(webhookId: string) {
		try {
			await api.delete(`/webhooks/${webhookId}`);
			webhooks = webhooks.filter(w => w.id !== webhookId);
			deleteTarget = null;
			showDeleteModal = false;
			toast.success($_('events.webhooks.deleteSuccess'));
		} catch (err) {
			const apiErr = err as ApiError;
			toast.error(apiErr.message || $_('events.webhooks.deleteError'));
		}
	}

	function confirmRotateSecret(webhook: Webhook) {
		rotateTarget = webhook;
		showRotateModal = true;
	}

	async function rotateSecret(webhookId: string) {
		rotating = true;
		try {
			const result = await api.post<{ data: { secret: string } }>(`/webhooks/${webhookId}/rotate-secret`);
			displayedSecret = result.data.secret;
			showRotateModal = false;
			rotateTarget = null;
			showSecretModal = true;
			toast.success($_('events.webhooks.rotateSuccess'));
		} catch (err) {
			const apiErr = err as ApiError;
			toast.error(apiErr.message || $_('events.webhooks.rotateError'));
		} finally {
			rotating = false;
		}
	}

	async function testWebhook(webhookId: string) {
		testingWebhookId = webhookId;
		try {
			await api.post(`/webhooks/${webhookId}/test`);
			toast.success($_('events.webhooks.testSuccess'));
		} catch (err) {
			const apiErr = err as ApiError;
			toast.error(apiErr.message || $_('events.webhooks.testError'));
		} finally {
			testingWebhookId = '';
		}
	}

	async function toggleDeliveries(webhookId: string) {
		if (expandedWebhookId === webhookId) {
			expandedWebhookId = '';
			deliveries = [];
			return;
		}
		expandedWebhookId = webhookId;
		loadingDeliveries = true;
		try {
			const result = await api.get<{ data: WebhookDelivery[] }>(`/webhooks/${webhookId}/deliveries`);
			deliveries = result.data;
		} catch {
			deliveries = [];
		} finally {
			loadingDeliveries = false;
		}
	}

	async function copySecret() {
		try {
			await navigator.clipboard.writeText(displayedSecret);
			secretCopied = true;
			setTimeout(() => (secretCopied = false), 2000);
		} catch {
			toast.error($_('events.webhooks.copyError'));
		}
	}

	function statusBadgeVariant(status?: number): 'success' | 'error' | 'warning' | 'neutral' {
		if (!status) return 'error';
		if (status >= 200 && status < 300) return 'success';
		if (status >= 400) return 'error';
		return 'warning';
	}
</script>

<svelte:head>
	<title>{$_('events.webhooks.pageTitle')}</title>
</svelte:head>

<AppShell>
	<div class="mb-6 flex items-center justify-between">
		<a href="/events/{eventId}" class="text-sm text-primary hover:text-primary-hover">&larr; {$_('events.webhooks.backToEvent')}</a>
	</div>

	<Card>
		{#snippet header()}
			<div class="flex items-center justify-between">
				<h1 class="text-xl font-bold font-display text-neutral-900">{$_('events.webhooks.heading')}</h1>
				{#if !showCreateForm}
					<Button size="sm" onclick={() => (showCreateForm = true)}>{$_('events.webhooks.addWebhook')}</Button>
				{/if}
			</div>
		{/snippet}

		{#if loading}
			<div class="flex justify-center py-8">
				<Spinner size="lg" class="text-primary" />
			</div>
		{:else}
			<!-- Create Form -->
			{#if showCreateForm}
				<div class="border border-neutral-200 rounded-lg p-4 mb-6 space-y-4 bg-neutral-50">
					<h3 class="text-sm font-semibold text-neutral-900">{$_('events.webhooks.newWebhookTitle')}</h3>
					<Input
						name="webhookUrl"
						type="url"
						label={$_('events.webhooks.endpointUrlLabel')}
						bind:value={newUrl}
						placeholder={$_('events.webhooks.urlPlaceholder')}
						required
					/>
					<Input
						name="webhookDescription"
						label={$_('events.webhooks.descriptionLabel')}
						bind:value={newDescription}
						placeholder={$_('events.webhooks.descriptionPlaceholder')}
					/>
					<fieldset>
						<legend class="block text-sm font-medium text-neutral-700 mb-2">{$_('events.webhooks.eventTypesLegend')}</legend>
						<div class="flex flex-wrap gap-2">
							{#each EVENT_TYPES as eventType}
								<label class="flex items-center gap-1.5 cursor-pointer">
									<input
										type="checkbox"
										checked={newEventTypes.includes(eventType)}
										onchange={() => (newEventTypes = toggleEventType(eventType, newEventTypes))}
										class="rounded border-neutral-300 text-primary focus:ring-primary/40"
									/>
									<span class="text-sm text-neutral-700">{eventType}</span>
								</label>
							{/each}
						</div>
					</fieldset>
					<div class="flex items-center justify-end gap-2">
						<Button variant="outline" size="sm" onclick={() => (showCreateForm = false)}>{$_('events.webhooks.cancel')}</Button>
						<Button size="sm" onclick={createWebhook} loading={creating}>{$_('events.webhooks.createWebhook')}</Button>
					</div>
				</div>
			{/if}

			<!-- Webhook List -->
			{#if webhooks.length === 0 && !showCreateForm}
				<p class="text-sm text-neutral-500 text-center py-8">
					{$_('events.webhooks.noWebhooks')}
				</p>
			{:else}
				<div class="space-y-4">
					{#each webhooks as webhook (webhook.id)}
						{#if editingWebhook?.id === webhook.id}
							<!-- Edit Form -->
							<div class="border border-primary-light rounded-lg p-4 space-y-4 bg-primary-lighter/30">
								<h3 class="text-sm font-semibold text-neutral-900">{$_('events.webhooks.editWebhookTitle')}</h3>
								<Input
									name="editWebhookUrl"
									type="url"
									label={$_('events.webhooks.endpointUrlLabel')}
									bind:value={editUrl}
									placeholder={$_('events.webhooks.urlPlaceholder')}
									required
								/>
								<Input
									name="editWebhookDescription"
									label={$_('events.webhooks.descriptionLabel')}
									bind:value={editDescription}
									placeholder={$_('events.webhooks.descriptionPlaceholder')}
								/>
								<fieldset>
									<legend class="block text-sm font-medium text-neutral-700 mb-2">{$_('events.webhooks.eventTypesLegend')}</legend>
									<div class="flex flex-wrap gap-2">
										{#each EVENT_TYPES as eventType}
											<label class="flex items-center gap-1.5 cursor-pointer">
												<input
													type="checkbox"
													checked={editEventTypes.includes(eventType)}
													onchange={() => (editEventTypes = toggleEventType(eventType, editEventTypes))}
													class="rounded border-neutral-300 text-primary focus:ring-primary/40"
												/>
												<span class="text-sm text-neutral-700">{eventType}</span>
											</label>
										{/each}
									</div>
								</fieldset>
								<label class="flex items-center gap-2 cursor-pointer">
									<input
										type="checkbox"
										bind:checked={editEnabled}
										class="rounded border-neutral-300 text-primary focus:ring-primary/40"
									/>
									<span class="text-sm text-neutral-700">{$_('events.webhooks.enabledLabel')}</span>
								</label>
								<div class="flex items-center justify-end gap-2">
									<Button variant="outline" size="sm" onclick={cancelEdit}>{$_('events.webhooks.cancel')}</Button>
									<Button size="sm" onclick={saveWebhook} loading={saving}>{$_('events.webhooks.save')}</Button>
								</div>
							</div>
						{:else}
							<!-- Webhook Card -->
							<div class="border border-neutral-200 rounded-lg p-4">
								<div class="flex items-start justify-between">
									<div class="min-w-0 flex-1">
										<div class="flex items-center gap-2 mb-1">
											<code class="text-sm font-medium text-neutral-900 truncate block max-w-md">{webhook.url}</code>
											<Badge variant={webhook.enabled ? 'success' : 'neutral'}>
												{webhook.enabled ? $_('events.webhooks.active') : $_('events.webhooks.disabled')}
											</Badge>
										</div>
										{#if webhook.description}
											<p class="text-sm text-neutral-500 mb-2">{webhook.description}</p>
										{/if}
										<div class="flex flex-wrap gap-1">
											{#each webhook.eventTypes as et}
												<span class="inline-flex items-center rounded-full bg-neutral-100 px-2 py-0.5 text-xs text-neutral-600">{et}</span>
											{/each}
										</div>
									</div>
									<div class="flex items-center gap-1 ml-4">
										<Button size="sm" variant="ghost" onclick={() => testWebhook(webhook.id)} loading={testingWebhookId === webhook.id}>{$_('events.webhooks.test')}</Button>
										<Button size="sm" variant="ghost" onclick={() => startEdit(webhook)}>{$_('events.webhooks.edit')}</Button>
										<Button size="sm" variant="ghost" onclick={() => confirmRotateSecret(webhook)}>{$_('events.webhooks.rotateSecret')}</Button>
										<Button size="sm" variant="ghost" onclick={() => { deleteTarget = webhook; showDeleteModal = true; }}>{$_('events.webhooks.delete')}</Button>
									</div>
								</div>

								<!-- Delivery log toggle -->
								<div class="mt-3 pt-3 border-t border-neutral-100">
									<button
										type="button"
										onclick={() => toggleDeliveries(webhook.id)}
										class="text-xs text-primary hover:text-primary-hover font-medium"
									>
										{expandedWebhookId === webhook.id ? $_('events.webhooks.hideDeliveries') : $_('events.webhooks.showDeliveries')}
									</button>

									{#if expandedWebhookId === webhook.id}
										<div class="mt-3">
											{#if loadingDeliveries}
												<div class="flex justify-center py-4">
													<Spinner class="text-primary" />
												</div>
											{:else if deliveries.length === 0}
												<p class="text-xs text-neutral-400 text-center py-4">{$_('events.webhooks.noDeliveries')}</p>
											{:else}
												<div class="space-y-2 max-h-64 overflow-y-auto">
													{#each deliveries as delivery (delivery.id)}
														<div class="rounded border border-neutral-100 p-2 text-xs">
															<div class="flex items-center justify-between mb-1">
																<div class="flex items-center gap-2">
																	<Badge variant={statusBadgeVariant(delivery.responseStatus)}>
																		{delivery.responseStatus || $_('events.webhooks.failed')}
																	</Badge>
																	<span class="text-neutral-600">{delivery.eventType}</span>
																</div>
																<span class="text-neutral-400">{new Date(delivery.createdAt).toLocaleString()}</span>
															</div>
															{#if delivery.error}
																<p class="text-error mt-1">{delivery.error}</p>
															{/if}
														</div>
													{/each}
												</div>
											{/if}
										</div>
									{/if}
								</div>
							</div>
						{/if}
					{/each}
				</div>
			{/if}
		{/if}
	</Card>

	<!-- Secret Modal -->
	<Modal bind:open={showSecretModal} title={$_('events.webhooks.secretModalTitle')}>
		<p class="text-sm text-neutral-600 mb-3">
			{$_('events.webhooks.secretModalBody')}
		</p>
		<div class="flex items-center gap-2 bg-neutral-50 rounded-lg px-4 py-3 border border-neutral-200">
			<code class="text-sm font-mono text-neutral-900 flex-1 break-all">{displayedSecret}</code>
			<button
				type="button"
				onclick={copySecret}
				class="text-neutral-400 hover:text-primary transition-colors flex-shrink-0"
				title={$_('events.webhooks.copySecretTitle')}
			>
				{#if secretCopied}
					<svg class="h-4 w-4 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
					</svg>
				{:else}
					<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
					</svg>
				{/if}
			</button>
		</div>
		{#snippet actions()}
			<Button size="sm" onclick={() => (showSecretModal = false)}>{$_('events.webhooks.done')}</Button>
		{/snippet}
	</Modal>

	<!-- Rotate Secret Confirmation Modal -->
	{#if rotateTarget}
		{@const target = rotateTarget}
		<Modal bind:open={showRotateModal} title={$_('events.webhooks.rotateModalTitle')}>
			<p class="text-sm text-neutral-600">
				{$_('events.webhooks.rotateModalBody', { values: { url: target.url } })}
			</p>
			{#snippet actions()}
				<Button variant="outline" size="sm" onclick={() => { showRotateModal = false; rotateTarget = null; }}>{$_('events.webhooks.cancel')}</Button>
				<Button variant="primary" size="sm" onclick={() => rotateSecret(target.id)} loading={rotating}>{$_('events.webhooks.rotateSecret')}</Button>
			{/snippet}
		</Modal>
	{/if}

	<!-- Delete Confirmation Modal -->
	{#if deleteTarget}
		{@const target = deleteTarget}
		<Modal bind:open={showDeleteModal} title={$_('events.webhooks.deleteModalTitle')}>
			<p class="text-sm text-neutral-600">
				{$_('events.webhooks.deleteModalBody', { values: { url: target.url } })}
			</p>
			{#snippet actions()}
				<Button variant="outline" size="sm" onclick={() => (showDeleteModal = false)}>{$_('events.webhooks.cancel')}</Button>
				<Button variant="danger" size="sm" onclick={() => deleteWebhook(target.id)}>{$_('events.webhooks.delete')}</Button>
			{/snippet}
		</Modal>
	{/if}
</AppShell>
