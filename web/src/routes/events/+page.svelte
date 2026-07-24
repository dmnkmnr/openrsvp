<script lang="ts">
	import { api } from '$lib/api/client';
	import { toast } from '$lib/stores/toast';
	import { currentUser } from '$lib/stores/auth';
	import { events, eventsLoading } from '$lib/stores/events';
	import { formatDateTime, daysUntil } from '$lib/utils/dates';
	import type { Event } from '$lib/types';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Spinner from '$lib/components/ui/Spinner.svelte';
	import { onMount } from 'svelte';
	import { _ } from '$lib/i18n';

	onMount(async () => {
		$eventsLoading = true;
		try {
			const result = await api.get<{ data: Event[] }>('/events');
			$events = result.data;
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.list.loadError'));
		} finally {
			$eventsLoading = false;
		}
	});

	function statusVariant(status: Event['status']): 'success' | 'warning' | 'error' | 'info' | 'neutral' {
		const map: Record<Event['status'], 'success' | 'warning' | 'error' | 'info' | 'neutral'> = {
			draft: 'neutral',
			published: 'success',
			cancelled: 'error',
			archived: 'warning'
		};
		return map[status];
	}

	function eventStatusLabel(status: Event['status']): string {
		return $_('events.detail.eventStatus.' + status);
	}
</script>

<svelte:head>
	<title>{$_('events.list.pageTitle')}</title>
</svelte:head>

<AppShell>
	<div class="flex items-center justify-between mb-8">
		<h1 class="text-2xl font-bold font-display text-neutral-900">{$_('events.list.heading')}</h1>
		<div class="flex items-center gap-3">
			<Button variant="outline" href="/events/series">{$_('events.list.series')}</Button>
			<Button href="/events/new">{$_('events.list.createEvent')}</Button>
		</div>
	</div>

	{#if $eventsLoading}
		<div class="flex items-center justify-center py-16">
			<Spinner size="lg" class="text-primary" />
		</div>
	{:else if $events.length === 0}
		<!-- Empty state -->
		<Card>
			<div class="text-center py-12">
				<svg
					class="mx-auto h-12 w-12 text-neutral-400"
					fill="none"
					viewBox="0 0 24 24"
					stroke="currentColor"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="1.5"
						d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
					/>
				</svg>
				<h3 class="mt-4 text-lg font-medium font-display text-neutral-900">{$_('events.list.emptyTitle')}</h3>
				<p class="mt-2 text-sm text-neutral-500">{$_('events.list.emptyBody')}</p>
				<div class="mt-6">
					<Button href="/events/new">{$_('events.list.createFirstEvent')}</Button>
				</div>
			</div>
		</Card>
	{:else}
		<!-- Event grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
			{#each $events as event (event.id)}
				<a href="/events/{event.id}" class="block group">
					<Card class="transition-shadow group-hover:shadow-md">
						<div class="flex items-start justify-between">
							<div class="flex-1 min-w-0">
								<h3
									class="text-lg font-semibold font-display text-neutral-900 group-hover:text-primary transition-colors truncate"
								>
									{event.title}
								</h3>
								<p class="mt-1 text-sm text-neutral-600">{formatDateTime(event.eventDate, event.timezone)}</p>
								{#if event.location}
									<p class="mt-1 text-sm text-neutral-500 flex items-center gap-1">
										<svg
											class="h-4 w-4 shrink-0"
											fill="none"
											viewBox="0 0 24 24"
											stroke="currentColor"
										>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z"
											/>
											<path
												stroke-linecap="round"
												stroke-linejoin="round"
												stroke-width="2"
												d="M15 11a3 3 0 11-6 0 3 3 0 016 0z"
											/>
										</svg>
										<span class="truncate">{event.location}</span>
									</p>
								{/if}
							</div>
							<div class="flex items-center gap-2">
								{#if event.seriesId}
									<span class="inline-flex items-center rounded-full bg-primary-lighter px-2 py-0.5 text-xs font-medium text-primary">{$_('events.list.seriesBadge')}</span>
								{/if}
								{#if event.organizerId !== $currentUser?.id}
									<span class="inline-flex items-center rounded-full bg-info-light px-2 py-0.5 text-xs font-medium text-info">{$_('events.list.cohostBadge')}</span>
								{/if}
								<Badge variant={statusVariant(event.status)}>
									{eventStatusLabel(event.status)}
								</Badge>
							</div>
						</div>
						<div class="mt-4 flex items-center justify-between text-xs text-neutral-500">
							<span>
								{#if daysUntil(event.eventDate) > 0}
									{$_('events.list.daysAway', { values: { count: daysUntil(event.eventDate) } })}
								{:else if daysUntil(event.eventDate) === 0}
									{$_('events.list.today')}
								{:else}
									{$_('events.list.daysAgo', { values: { count: Math.abs(daysUntil(event.eventDate)) } })}
								{/if}
							</span>
							<span class="font-mono text-neutral-400">{event.shareToken}</span>
						</div>
					</Card>
				</a>
			{/each}
		</div>
	{/if}
</AppShell>
