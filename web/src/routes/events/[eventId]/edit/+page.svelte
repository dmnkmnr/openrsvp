<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { api } from '$lib/api/client';
	import { toast } from '$lib/stores/toast';
	import { smsEnabled, supportedLanguages, loadAppConfig } from '$lib/stores/config';
	import { toISOLocal, datetimeLocalToUTC, utcToDatetimeLocal } from '$lib/utils/dates';
	import { getTimezoneOptions } from '$lib/utils/timezones';
	import type { Event, RSVPStats } from '$lib/types';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import DateTimePicker from '$lib/components/ui/DateTimePicker.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Spinner from '$lib/components/ui/Spinner.svelte';
	import QuestionBuilder from '$lib/components/questions/QuestionBuilder.svelte';
	import { onMount } from 'svelte';
	import { _ } from '$lib/i18n';

	const eventId = $derived($page.params.eventId);

	let loading = $state(true);
	let saving = $state(false);

	let title = $state('');
	let eventDate = $state('');
	let endDate = $state('');
	let location = $state('');
	let timezone = $state('');
	let description = $state('');
	let contactRequirement = $state('email_or_phone');
	let mapProvider = $state('google');
	let mapCustomUrl = $state('');
	let language = $state('en');
	let showHeadcount = $state(false);
	let showGuestList = $state(false);
	let rsvpDeadline = $state('');
	let maxCapacity = $state('');
	let retentionDays = $state('30');
	let showRetention = $state(false);
	let attendingHeadcount = $state(0);
	let waitlistEnabled = $state(false);

	const capacityWarning = $derived(
		maxCapacity && parseInt(maxCapacity) > 0 && attendingHeadcount > parseInt(maxCapacity)
	);

	const contactRequirementOptions = $derived([
		{ value: 'email_or_phone', label: $_('events.new.contactOptions.emailOrPhone') },
		{ value: 'email', label: $_('events.new.contactOptions.emailOnly') },
		{ value: 'phone', label: $_('events.new.contactOptions.phoneOnly') },
		{ value: 'email_and_phone', label: $_('events.new.contactOptions.emailAndPhone') }
	]);

	const filteredContactOptions = $derived(
		$smsEnabled
			? contactRequirementOptions
			: contactRequirementOptions.filter(o => o.value !== 'phone')
	);

	const mapProviderOptions = $derived([
		{ value: 'google', label: $_('events.new.mapProviderOptions.google') },
		{ value: 'osm', label: $_('events.new.mapProviderOptions.osm') },
		{ value: 'custom', label: $_('events.new.mapProviderOptions.custom') },
		{ value: 'none', label: $_('events.new.mapProviderOptions.none') }
	]);

	const languageNames: Record<string, string> = { en: 'English', de: 'Deutsch' };
	const languageOptions = $derived(
		$supportedLanguages.map(code => ({ value: code, label: languageNames[code] || code }))
	);

	let errors: Record<string, string> = $state({});

	let tzOptions = $state(getTimezoneOptions());

	onMount(async () => {
		loadAppConfig();
		try {
			const [eventResult, statsResult] = await Promise.all([
				api.get<{ data: Event }>(`/events/${eventId}`),
				api.get<{ data: RSVPStats }>(`/rsvp/event/${eventId}/stats`).catch(() => ({
					data: { attending: 0, attendingHeadcount: 0, attendingChildren: 0, maybe: 0, maybeHeadcount: 0, declined: 0, pending: 0, waitlisted: 0, total: 0, totalHeadcount: 0 }
				}))
			]);
			const e = eventResult.data;
			title = e.title;
			eventDate = e.eventDate ? utcToDatetimeLocal(e.eventDate, e.timezone) : '';
			endDate = e.endDate ? utcToDatetimeLocal(e.endDate, e.timezone) : '';
			location = e.location;
			timezone = e.timezone;
			tzOptions = getTimezoneOptions(e.timezone);
			description = e.description;
			contactRequirement = e.contactRequirement || 'email_or_phone';
			mapProvider = e.mapProvider || 'google';
			mapCustomUrl = e.mapCustomUrl || '';
			language = e.language || 'en';
			showHeadcount = e.showHeadcount ?? false;
			showGuestList = e.showGuestList ?? false;
			rsvpDeadline = e.rsvpDeadline ? utcToDatetimeLocal(e.rsvpDeadline, e.timezone) : '';
			maxCapacity = e.maxCapacity ? String(e.maxCapacity) : '';
			retentionDays = String(e.retentionDays);
			showRetention = e.retentionDays !== 30;
			waitlistEnabled = e.waitlistEnabled ?? false;
			attendingHeadcount = statsResult.data.attendingHeadcount;
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.edit.loadError'));
		} finally {
			loading = false;
		}
	});

	function validate(): boolean {
		errors = {};
		if (!title.trim()) errors.title = $_('events.new.titleRequired');
		if (!eventDate) errors.eventDate = $_('events.new.eventDateRequired');
		if (!timezone) errors.timezone = $_('events.new.timezoneRequired');
		if (showRetention) {
			const days = parseInt(retentionDays);
			if (isNaN(days) || days < 1 || days > 365) {
				errors.retentionDays = $_('events.new.retentionInvalid');
			}
		}
		if (maxCapacity) {
			const parsed = Number(maxCapacity);
			if (!Number.isInteger(parsed) || parsed < 1) {
				errors.maxCapacity = $_('events.new.maxCapacityInvalid');
			}
		}
		if (mapProvider === 'custom' && !/^https?:\/\/.+/.test(mapCustomUrl.trim())) {
			errors.mapCustomUrl = $_('events.new.mapCustomUrlInvalid');
		}
		return Object.keys(errors).length === 0;
	}

	async function handleSave() {
		if (!validate()) return;

		saving = true;
		try {
			const body: Record<string, unknown> = {
				title: title.trim(),
				eventDate: eventDate ? datetimeLocalToUTC(eventDate, timezone) : eventDate,
				location: location.trim(),
				timezone,
				description: description.trim(),
				language,
				contactRequirement,
				mapProvider,
				mapCustomUrl: mapProvider === 'custom' ? mapCustomUrl.trim() : '',
				showHeadcount,
				showGuestList,
				retentionDays: parseInt(retentionDays)
			};
			if (endDate) body.endDate = datetimeLocalToUTC(endDate, timezone);
			if (rsvpDeadline) body.rsvpDeadline = datetimeLocalToUTC(rsvpDeadline, timezone);
			else body.rsvpDeadline = '';
			if (maxCapacity) {
				body.maxCapacity = parseInt(maxCapacity);
				body.waitlistEnabled = waitlistEnabled;
			} else {
				body.maxCapacity = 0;
				body.waitlistEnabled = false;
			}

			await api.put(`/events/${eventId}`, body);
			toast.success($_('events.edit.updateSuccess'));
			goto(`/events/${eventId}`);
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.edit.updateError'));
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('events.edit.pageTitle')}</title>
</svelte:head>

<AppShell>
	<div class="max-w-3xl mx-auto">
		<div class="mb-8">
			<a href="/events/{eventId}" class="text-sm text-primary hover:text-primary-hover">&larr; {$_('events.edit.backToEvent')}</a>
			<h1 class="mt-2 text-2xl font-bold font-display text-neutral-900">{$_('events.edit.heading')}</h1>
		</div>

		{#if loading}
			<div class="flex items-center justify-center py-16">
				<Spinner size="lg" class="text-primary" />
			</div>
		{:else}
			<Card>
				<form
					onsubmit={(e) => {
						e.preventDefault();
						handleSave();
					}}
					class="space-y-6"
				>
					<Input
						label={$_('events.new.titleLabel')}
						name="title"
						bind:value={title}
						placeholder={$_('events.new.titlePlaceholder')}
						error={errors.title || ''}
						required
					/>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<DateTimePicker
							label={$_('events.new.eventDateLabel')}
							name="eventDate"
							bind:value={eventDate}
							error={errors.eventDate || ''}
							required
						/>
						<DateTimePicker
							label={$_('events.new.endDateLabel')}
							name="endDate"
							bind:value={endDate}
							min={eventDate}
						/>
					</div>

					<Input
						label={$_('events.new.locationLabel')}
						name="location"
						bind:value={location}
						placeholder={$_('events.new.locationPlaceholder')}
					/>

					<Select
						label={$_('events.new.timezoneLabel')}
						name="timezone"
						bind:value={timezone}
						options={tzOptions}
						error={errors.timezone || ''}
						required
					/>

					<Textarea
						label={$_('events.new.descriptionLabel')}
						name="description"
						bind:value={description}
						placeholder={$_('events.new.descriptionPlaceholder')}
						rows={6}
					/>

					<div>
						<Select
							label={$_('events.new.languageLabel')}
							name="language"
							bind:value={language}
							options={languageOptions}
						/>
						<p class="mt-1.5 text-xs text-neutral-400">{$_('events.new.languageHelper')}</p>
					</div>

					<Select
						label={$_('events.new.contactRequirementLabel')}
						name="contactRequirement"
						bind:value={contactRequirement}
						options={filteredContactOptions}
					/>

					<div>
						<Select
							label={$_('events.new.mapProviderLabel')}
							name="mapProvider"
							bind:value={mapProvider}
							options={mapProviderOptions}
						/>
						<p class="mt-1.5 text-xs text-neutral-400">{$_('events.new.mapProviderHelper')}</p>
					</div>

					{#if mapProvider === 'custom'}
						<Input
							label={$_('events.new.mapCustomUrlLabel')}
							name="mapCustomUrl"
							type="url"
							bind:value={mapCustomUrl}
							placeholder="https://maps.example.com/..."
							error={errors.mapCustomUrl || ''}
						/>
					{/if}

					<fieldset class="pt-2">
						<legend class="text-sm font-medium text-neutral-700 mb-3">{$_('events.new.guestVisibilityLegend')}</legend>
						<p class="text-xs text-neutral-400 mb-3">{$_('events.new.guestVisibilityHelper')}</p>
						<div class="space-y-2">
							<label class="flex items-center gap-3 cursor-pointer">
								<input
									type="checkbox"
									bind:checked={showHeadcount}
									class="rounded border-neutral-300 text-primary focus:ring-primary/40"
								/>
								<span class="text-sm text-neutral-700">{$_('events.new.showHeadcountLabel')}</span>
							</label>
							<label class="flex items-center gap-3 cursor-pointer">
								<input
									type="checkbox"
									bind:checked={showGuestList}
									class="rounded border-neutral-300 text-primary focus:ring-primary/40"
								/>
								<span class="text-sm text-neutral-700">{$_('events.new.showGuestListLabel')}</span>
							</label>
						</div>
					</fieldset>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<DateTimePicker
							label={$_('events.new.rsvpDeadlineLabel')}
							name="rsvpDeadline"
							bind:value={rsvpDeadline}
							max={eventDate || undefined}
							helper={$_('events.new.rsvpDeadlineHelper')}
						/>
						<Input
							label={$_('events.new.maxCapacityLabel')}
							name="maxCapacity"
							type="number"
							bind:value={maxCapacity}
							placeholder={$_('events.new.maxCapacityPlaceholder')}
							helper={$_('events.new.maxCapacityHelper')}
							error={errors.maxCapacity || ''}
						/>
					</div>

					{#if capacityWarning}
						<div class="rounded-lg bg-warning-light border border-warning px-4 py-3 text-sm text-warning flex items-start gap-2">
							<svg class="h-4 w-4 text-warning mt-0.5 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
							</svg>
							<span>
								{$_('events.edit.capacityWarning', { values: { count: attendingHeadcount } })}
							</span>
						</div>
					{/if}

					{#if maxCapacity}
						<label class="flex items-center gap-3 cursor-pointer">
							<input
								type="checkbox"
								bind:checked={waitlistEnabled}
								class="rounded border-neutral-300 text-primary focus:ring-primary/40"
							/>
							<div>
								<span class="text-sm text-neutral-700">{$_('events.new.waitlistLabel')}</span>
								<p class="text-xs text-neutral-400">{$_('events.new.waitlistHelper')}</p>
							</div>
						</label>
					{/if}

					<div class="pt-2">
						{#if showRetention}
							<Input
								label={$_('events.new.retentionLabel')}
								name="retentionDays"
								type="number"
								bind:value={retentionDays}
								helper={$_('events.new.retentionHelper')}
								error={errors.retentionDays || ''}
							/>
						{:else}
							<p class="text-xs text-neutral-400">
								{$_('events.new.retentionDefaultNote')}
								<button
									type="button"
									class="text-primary hover:text-primary-hover underline underline-offset-2"
									onclick={() => (showRetention = true)}
								>
									{$_('events.new.retentionCustomize')}
								</button>
							</p>
						{/if}
					</div>

					<div class="flex items-center justify-end gap-3 pt-4 border-t border-neutral-200">
						<Button variant="outline" href="/events/{eventId}">{$_('events.edit.cancel')}</Button>
						<Button type="submit" loading={saving}>{$_('events.edit.saveChanges')}</Button>
					</div>
				</form>
			</Card>

			<Card class="mt-6">
				<div class="flex items-center justify-between">
					<div>
						<h2 class="text-base font-semibold text-neutral-900">{$_('events.messageTemplates.cardHeading')}</h2>
						<p class="text-sm text-neutral-500 mt-1">{$_('events.messageTemplates.cardDescription')}</p>
					</div>
					<Button variant="outline" href="/events/{eventId}/message-templates">{$_('events.messageTemplates.cardAction')}</Button>
				</div>
			</Card>

			<Card class="mt-6">
				<QuestionBuilder eventId={eventId ?? ''} />
			</Card>
		{/if}
	</div>
</AppShell>
