<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { toast } from '$lib/stores/toast';
	import { formatDateTime } from '$lib/utils/dates';
	import type { Message, Event, Attendee } from '$lib/types';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Spinner from '$lib/components/ui/Spinner.svelte';
	import SmsCharCounter from '$lib/components/ui/SmsCharCounter.svelte';
	import { onMount, tick } from 'svelte';
	import { _ } from '$lib/i18n';

	const eventId = $derived($page.params.eventId);

	let loading = $state(true);
	let sending = $state(false);
	let event: Event | null = $state(null);
	let messages: Message[] = $state([]);
	let attendeeMap: Record<string, string> = $state({});

	// Compose form
	let recipientType = $state('all');
	let subject = $state('');
	let body = $state('');

	// Reply state
	let replyToAttendeeId = $state('');
	let replyToAttendeeName = $state('');

	let composeErrors: Record<string, string> = $state({});

	let composeForm: HTMLFormElement | undefined = $state();

	const recipientOptions = $derived([
		{ value: 'all', label: $_('events.messages.recipient.all') },
		{ value: 'attending', label: $_('events.messages.recipient.attending') },
		{ value: 'maybe', label: $_('events.messages.recipient.maybe') },
		{ value: 'declined', label: $_('events.messages.recipient.declined') },
		{ value: 'pending', label: $_('events.messages.recipient.pending') }
	]);

	const recipientLabels: Record<string, string> = $derived({
		all: $_('events.messages.recipient.all'),
		attending: $_('events.messages.recipient.attending'),
		maybe: $_('events.messages.recipient.maybe'),
		declined: $_('events.messages.recipient.declined'),
		pending: $_('events.messages.recipient.pending')
	});

	function attendeeName(id: string): string {
		return attendeeMap[id] || $_('events.messages.unknown');
	}

	function messageLabel(msg: Message): string {
		if (msg.senderType === 'attendee') {
			return $_('events.messages.from', { values: { name: attendeeName(msg.senderId) } });
		}
		if (msg.recipientType === 'attendee') {
			return $_('events.messages.to', { values: { name: attendeeName(msg.recipientId) } });
		}
		return $_('events.messages.to', { values: { name: recipientLabels[msg.recipientId] || msg.recipientId } });
	}

	function isIncoming(msg: Message): boolean {
		return msg.senderType === 'attendee';
	}

	async function handleReply(msg: Message) {
		replyToAttendeeId = msg.senderId;
		replyToAttendeeName = attendeeName(msg.senderId);
		subject = msg.subject.startsWith('Re: ') ? msg.subject : 'Re: ' + msg.subject;
		body = '';
		await tick();
		composeForm?.scrollIntoView({ behavior: 'smooth', block: 'start' });
		// Focus the body textarea after scrolling
		setTimeout(() => {
			const textarea = composeForm?.querySelector('textarea');
			textarea?.focus();
		}, 300);
	}

	function cancelReply() {
		replyToAttendeeId = '';
		replyToAttendeeName = '';
		subject = '';
		body = '';
	}

	onMount(async () => {
		try {
			const [eventResult, messagesResult, rsvpResult] = await Promise.all([
				api.get<{ data: Event }>(`/events/${eventId}`),
				api.get<{ data: Message[] }>(`/messages/event/${eventId}`).catch(() => ({ data: [] })),
				api.get<{ data: Attendee[] }>(`/rsvp/event/${eventId}`).catch(() => ({ data: [] }))
			]);
			event = eventResult.data;
			messages = messagesResult.data;
			// Build attendee lookup map
			const map: Record<string, string> = {};
			for (const a of rsvpResult.data) {
				map[a.id] = a.name;
			}
			attendeeMap = map;
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.messages.loadError'));
		} finally {
			loading = false;
		}
	});

	async function handleSend() {
		composeErrors = {};
		if (!subject.trim()) composeErrors.subject = $_('events.messages.subjectRequired');
		if (!body.trim()) composeErrors.body = $_('events.messages.bodyRequired');
		if (Object.keys(composeErrors).length > 0) return;

		sending = true;
		try {
			const payload = replyToAttendeeId
				? {
						recipientType: 'attendee',
						recipientId: replyToAttendeeId,
						subject: subject.trim(),
						body: body.trim()
					}
				: {
						recipientType: 'group',
						recipientId: recipientType,
						subject: subject.trim(),
						body: body.trim()
					};

			const result = await api.post<{ data: Message }>(`/messages/event/${eventId}`, payload);
			messages = [result.data, ...messages];
			subject = '';
			body = '';
			replyToAttendeeId = '';
			replyToAttendeeName = '';
			toast.success($_('events.messages.sendSuccess'));
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.messages.sendError'));
		} finally {
			sending = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('events.messages.pageTitle')}</title>
</svelte:head>

<AppShell>
	<div class="max-w-3xl mx-auto">
		<div class="mb-6">
			<a href="/events/{eventId}" class="text-sm text-primary hover:text-primary-hover">&larr; {$_('events.messages.backToEvent')}</a>
			<h1 class="mt-2 text-2xl font-bold font-display text-neutral-900">{$_('events.messages.heading')}</h1>
			{#if event}
				<p class="text-sm text-neutral-500">{event.title}</p>
			{/if}
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<Spinner size="lg" class="text-primary" />
			</div>
		{:else}
			<!-- Compose form -->
			<Card class="mb-6">
				{#snippet header()}
					<h2 class="text-lg font-semibold font-display text-neutral-900">{$_('events.messages.composeTitle')}</h2>
				{/snippet}

				<form
					bind:this={composeForm}
					onsubmit={(e) => {
						e.preventDefault();
						handleSend();
					}}
					class="space-y-4"
				>
					{#if replyToAttendeeId}
						<div class="flex items-center gap-2">
							<span class="inline-flex items-center gap-1 rounded-full bg-primary-lighter px-3 py-1 text-sm font-medium text-primary">
								{$_('events.messages.replyingTo', { values: { name: replyToAttendeeName } })}
								<button
									type="button"
									onclick={cancelReply}
									class="ml-1 inline-flex h-4 w-4 items-center justify-center rounded-full text-primary hover:bg-primary-light hover:text-primary"
									aria-label={$_('events.messages.cancelReply')}
								>
									&times;
								</button>
							</span>
						</div>
					{:else}
						<Select
							label={$_('events.messages.sendToLabel')}
							name="recipientType"
							bind:value={recipientType}
							options={recipientOptions}
						/>
					{/if}

					<Input
						label={$_('events.messages.subjectLabel')}
						name="subject"
						bind:value={subject}
						placeholder={$_('events.messages.subjectPlaceholder')}
						error={composeErrors.subject || ''}
						required
					/>

					<Textarea
						label={$_('events.messages.messageLabel')}
						name="body"
						bind:value={body}
						placeholder={$_('events.messages.messagePlaceholder')}
						rows={4}
						error={composeErrors.body || ''}
						required
					/>
					<SmsCharCounter text={body} />
					<div>
						<p class="text-xs font-medium text-neutral-500 mb-2">{$_('events.messageTemplates.variablesHint')}</p>
						<div class="flex flex-wrap gap-2">
							{#each ['guestName', 'eventTitle', 'eventDate', 'location', 'rsvpLink'] as v (v)}
								<code class="text-xs bg-neutral-100 text-neutral-700 rounded px-2 py-1">{'{' + v + '}'}</code>
							{/each}
						</div>
					</div>

					<div class="flex justify-end">
						<Button type="submit" loading={sending}>{$_('events.messages.sendMessage')}</Button>
					</div>
				</form>
			</Card>

			<!-- Message list -->
			<Card>
				{#snippet header()}
					<h2 class="text-lg font-semibold font-display text-neutral-900">{$_('events.messages.allMessagesTitle')}</h2>
				{/snippet}

				{#if messages.length === 0}
					<p class="text-sm text-neutral-500 text-center py-8">{$_('events.messages.noMessages')}</p>
				{:else}
					<div class="divide-y divide-neutral-200 -mx-6 -mb-4">
						{#each messages as message (message.id)}
							<div class="px-6 py-4 {isIncoming(message) ? 'bg-primary-lighter/50' : ''}">
								<div class="flex items-start justify-between">
									<div class="flex-1 min-w-0">
										<p class="text-sm font-medium text-neutral-900">{message.subject}</p>
										<p class="text-xs text-neutral-500 mt-0.5">
											{messageLabel(message)}
											&middot; {formatDateTime(message.createdAt)}
										</p>
									</div>
									{#if isIncoming(message)}
										<button
											type="button"
											onclick={() => handleReply(message)}
											class="ml-3 shrink-0 text-xs font-medium text-primary hover:text-primary-hover"
										>
											{$_('events.messages.reply')}
										</button>
									{/if}
								</div>
								<p class="mt-2 text-sm text-neutral-700 whitespace-pre-wrap">{message.body}</p>
							</div>
						{/each}
					</div>
				{/if}
			</Card>
		{/if}
	</div>
</AppShell>
