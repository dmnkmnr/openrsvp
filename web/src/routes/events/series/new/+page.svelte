<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { currentUser } from '$lib/stores/auth';
	import { toast } from '$lib/stores/toast';
	import { smsEnabled, loadAppConfig } from '$lib/stores/config';
	import { datetimeLocalToUTC } from '$lib/utils/dates';
	import { getTimezoneOptions, getTimezoneLabel } from '$lib/utils/timezones';
	import type { EventSeries } from '$lib/types';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import { _ } from '$lib/i18n';

	const defaultTz = $currentUser?.timezone
		|| Intl.DateTimeFormat().resolvedOptions().timeZone
		|| '';

	let submitting = $state(false);

	// Form fields
	let title = $state('');
	let description = $state('');
	let location = $state('');
	let timezone = $state(defaultTz);
	let startDate = $state('');
	let eventTime = $state('');
	let durationMinutes = $state('');
	let recurrenceRule = $state('weekly');
	let endCondition = $state<'none' | 'count' | 'date'>('none');
	let maxOccurrences = $state('');
	let recurrenceEnd = $state('');
	let contactRequirement = $state('email');
	let mapProvider = $state('google');
	let mapCustomUrl = $state('');
	let showHeadcount = $state(false);
	let showGuestList = $state(false);
	let rsvpDeadlineOffsetHours = $state('');
	let maxCapacity = $state('');
	let retentionDays = $state('30');
	let showRetention = $state(false);

	// Validation
	let errors: Record<string, string> = $state({});

	const tzOptions = getTimezoneOptions(defaultTz);

	const recurrenceOptions = $derived([
		{ value: 'weekly', label: $_('events.seriesNew.recurrenceOptions.weekly') },
		{ value: 'biweekly', label: $_('events.seriesNew.recurrenceOptions.biweekly') },
		{ value: 'monthly', label: $_('events.seriesNew.recurrenceOptions.monthly') }
	]);

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

	// Today's date for the min attribute on the date picker
	const today = new Date().toISOString().split('T')[0];

	onMount(() => {
		loadAppConfig();
	});

	function validate(): boolean {
		errors = {};
		if (!title.trim()) errors.title = $_('events.seriesNew.titleRequired');
		if (!startDate) errors.startDate = $_('events.seriesNew.startDateRequired');
		if (!eventTime) errors.eventTime = $_('events.seriesNew.eventTimeRequired');
		if (!timezone) errors.timezone = $_('events.seriesNew.timezoneRequired');

		if (endCondition === 'count') {
			const n = parseInt(maxOccurrences);
			if (isNaN(n) || n < 1) errors.maxOccurrences = $_('events.seriesNew.maxOccurrencesInvalid');
		}
		if (endCondition === 'date' && !recurrenceEnd) {
			errors.recurrenceEnd = $_('events.seriesNew.recurrenceEndRequired');
		}
		if (durationMinutes) {
			const d = parseInt(durationMinutes);
			if (isNaN(d) || d < 1) errors.durationMinutes = $_('events.seriesNew.durationInvalid');
		}
		if (maxCapacity) {
			const parsed = Number(maxCapacity);
			if (!Number.isInteger(parsed) || parsed < 1) {
				errors.maxCapacity = $_('events.seriesNew.maxCapacityInvalid');
			}
		}
		if (rsvpDeadlineOffsetHours) {
			const h = parseInt(rsvpDeadlineOffsetHours);
			if (isNaN(h) || h < 1) errors.rsvpDeadlineOffsetHours = $_('events.seriesNew.rsvpOffsetInvalid');
		}
		if (showRetention) {
			const days = parseInt(retentionDays);
			if (isNaN(days) || days < 1 || days > 365) {
				errors.retentionDays = $_('events.seriesNew.retentionInvalid');
			}
		}
		if (mapProvider === 'custom' && !/^https?:\/\/.+/.test(mapCustomUrl.trim())) {
			errors.mapCustomUrl = $_('events.new.mapCustomUrlInvalid');
		}
		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validate()) return;

		submitting = true;
		try {
			const body: Record<string, unknown> = {
				title: title.trim(),
				description: description.trim(),
				location: location.trim(),
				timezone,
				startDate,
				eventTime,
				recurrenceRule,
				contactRequirement,
				mapProvider,
				mapCustomUrl: mapProvider === 'custom' ? mapCustomUrl.trim() : '',
				showHeadcount,
				showGuestList,
				retentionDays: parseInt(retentionDays)
			};
			if (durationMinutes) body.durationMinutes = parseInt(durationMinutes);
			if (endCondition === 'count' && maxOccurrences) body.maxOccurrences = parseInt(maxOccurrences);
			if (endCondition === 'date' && recurrenceEnd) body.recurrenceEnd = datetimeLocalToUTC(recurrenceEnd + 'T23:59:59', timezone);
			if (rsvpDeadlineOffsetHours) body.rsvpDeadlineOffsetHours = parseInt(rsvpDeadlineOffsetHours);
			if (maxCapacity) body.maxCapacity = parseInt(maxCapacity);

			const result = await api.post<{ data: EventSeries }>('/events/series', body);
			toast.success($_('events.seriesNew.createSuccess'));
			goto(`/events/series/${result.data.id}`);
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.seriesNew.createError'));
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('events.seriesNew.pageTitle')}</title>
</svelte:head>

<AppShell>
	<div class="max-w-3xl mx-auto">
		<div class="mb-8">
			<a href="/events/series" class="text-sm text-primary hover:text-primary-hover">&larr; {$_('events.seriesNew.backToSeries')}</a>
			<h1 class="mt-2 text-2xl font-bold font-display text-neutral-900">{$_('events.seriesNew.heading')}</h1>
			<p class="mt-1 text-sm text-neutral-500">{$_('events.seriesNew.subheading')}</p>
		</div>

		<form
			onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}
		>
			<Card class="mb-6">
				<div class="space-y-6">
					<h2 class="text-lg font-semibold font-display text-neutral-900">{$_('events.seriesNew.eventDetailsHeading')}</h2>

					<Input
						label={$_('events.seriesNew.seriesTitleLabel')}
						name="title"
						bind:value={title}
						placeholder={$_('events.seriesNew.seriesTitlePlaceholder')}
						error={errors.title || ''}
						required
					/>

					<Textarea
						label={$_('events.seriesNew.descriptionLabel')}
						name="description"
						bind:value={description}
						placeholder={$_('events.seriesNew.descriptionPlaceholder')}
						rows={4}
					/>

					<Input
						label={$_('events.seriesNew.locationLabel')}
						name="location"
						bind:value={location}
						placeholder={$_('events.seriesNew.locationPlaceholder')}
					/>

					<Select
						label={$_('events.seriesNew.timezoneLabel')}
						name="timezone"
						bind:value={timezone}
						options={tzOptions}
						error={errors.timezone || ''}
						required
					/>

					<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
						<div class="space-y-1">
							<label for="startDate" class="block text-sm font-medium text-neutral-700">
								{$_('events.seriesNew.startDateLabel')} <span class="text-error">*</span>
							</label>
							<input
								id="startDate"
								type="date"
								bind:value={startDate}
								min={today}
								required
								class="block w-full rounded-lg border px-3 py-2 text-sm shadow-sm transition-colors focus:outline-none focus:ring-2 focus:ring-offset-0 {errors.startDate
									? 'border-error text-error focus:border-error focus:ring-error'
									: 'border-neutral-300 text-neutral-900 focus:border-primary focus:ring-primary'}"
							/>
							{#if errors.startDate}
								<p class="text-sm text-error">{errors.startDate}</p>
							{/if}
						</div>

						<div class="space-y-1">
							<label for="eventTime" class="block text-sm font-medium text-neutral-700">
								{$_('events.seriesNew.eventTimeLabel')} <span class="text-error">*</span>
							</label>
							<input
								id="eventTime"
								type="time"
								bind:value={eventTime}
								required
								class="block w-full rounded-lg border px-3 py-2 text-sm shadow-sm transition-colors focus:outline-none focus:ring-2 focus:ring-offset-0 {errors.eventTime
									? 'border-error text-error focus:border-error focus:ring-error'
									: 'border-neutral-300 text-neutral-900 focus:border-primary focus:ring-primary'}"
							/>
							{#if errors.eventTime}
								<p class="text-sm text-error">{errors.eventTime}</p>
							{/if}
						</div>

						<Input
							label={$_('events.seriesNew.durationLabel')}
							name="durationMinutes"
							type="number"
							bind:value={durationMinutes}
							placeholder={$_('events.seriesNew.durationPlaceholder')}
							error={errors.durationMinutes || ''}
						/>
					</div>
				</div>
			</Card>

			<Card class="mb-6">
				<div class="space-y-6">
					<h2 class="text-lg font-semibold font-display text-neutral-900">{$_('events.seriesNew.recurrenceHeading')}</h2>

					<Select
						label={$_('events.seriesNew.repeatLabel')}
						name="recurrenceRule"
						bind:value={recurrenceRule}
						options={recurrenceOptions}
						required
					/>

					<fieldset>
						<legend class="text-sm font-medium text-neutral-700 mb-3">{$_('events.seriesNew.endConditionLegend')}</legend>
						<div class="space-y-3">
							<label class="flex items-center gap-3 cursor-pointer">
								<input
									type="radio"
									name="endCondition"
									value="none"
									bind:group={endCondition}
									class="text-primary focus:ring-primary/40"
								/>
								<span class="text-sm text-neutral-700">{$_('events.seriesNew.endConditionNone')}</span>
							</label>
							<label class="flex items-center gap-3 cursor-pointer">
								<input
									type="radio"
									name="endCondition"
									value="count"
									bind:group={endCondition}
									class="text-primary focus:ring-primary/40"
								/>
								<span class="text-sm text-neutral-700">{$_('events.seriesNew.endConditionCount')}</span>
							</label>
							{#if endCondition === 'count'}
								<div class="ml-7">
									<Input
										name="maxOccurrences"
										type="number"
										bind:value={maxOccurrences}
										placeholder={$_('events.seriesNew.maxOccurrencesPlaceholder')}
										error={errors.maxOccurrences || ''}
									/>
								</div>
							{/if}
							<label class="flex items-center gap-3 cursor-pointer">
								<input
									type="radio"
									name="endCondition"
									value="date"
									bind:group={endCondition}
									class="text-primary focus:ring-primary/40"
								/>
								<span class="text-sm text-neutral-700">{$_('events.seriesNew.endConditionDate')}</span>
							</label>
							{#if endCondition === 'date'}
								<div class="ml-7 space-y-1">
									<input
										type="date"
										bind:value={recurrenceEnd}
										min={startDate || today}
										class="block w-full rounded-lg border px-3 py-2 text-sm shadow-sm transition-colors focus:outline-none focus:ring-2 focus:ring-offset-0 {errors.recurrenceEnd
											? 'border-error text-error focus:border-error focus:ring-error'
											: 'border-neutral-300 text-neutral-900 focus:border-primary focus:ring-primary'}"
									/>
									{#if errors.recurrenceEnd}
										<p class="text-sm text-error">{errors.recurrenceEnd}</p>
									{/if}
								</div>
							{/if}
						</div>
					</fieldset>
				</div>
			</Card>

			<Card class="mb-6">
				<div class="space-y-6">
					<h2 class="text-lg font-semibold font-display text-neutral-900">{$_('events.seriesNew.rsvpSettingsHeading')}</h2>

					<Select
						label={$_('events.seriesNew.contactRequirementLabel')}
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
						<legend class="text-sm font-medium text-neutral-700 mb-3">{$_('events.seriesNew.guestVisibilityLegend')}</legend>
						<p class="text-xs text-neutral-400 mb-3">{$_('events.seriesNew.guestVisibilityHelper')}</p>
						<div class="space-y-2">
							<label class="flex items-center gap-3 cursor-pointer">
								<input
									type="checkbox"
									bind:checked={showHeadcount}
									class="rounded border-neutral-300 text-primary focus:ring-primary/40"
								/>
								<span class="text-sm text-neutral-700">{$_('events.seriesNew.showHeadcountLabel')}</span>
							</label>
							<label class="flex items-center gap-3 cursor-pointer">
								<input
									type="checkbox"
									bind:checked={showGuestList}
									class="rounded border-neutral-300 text-primary focus:ring-primary/40"
								/>
								<span class="text-sm text-neutral-700">{$_('events.seriesNew.showGuestListLabel')}</span>
							</label>
						</div>
					</fieldset>

					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<Input
							label={$_('events.seriesNew.rsvpOffsetLabel')}
							name="rsvpDeadlineOffsetHours"
							type="number"
							bind:value={rsvpDeadlineOffsetHours}
							placeholder={$_('events.seriesNew.rsvpOffsetPlaceholder')}
							helper={$_('events.seriesNew.rsvpOffsetHelper')}
							error={errors.rsvpDeadlineOffsetHours || ''}
						/>
						<Input
							label={$_('events.seriesNew.maxCapacityLabel')}
							name="maxCapacity"
							type="number"
							bind:value={maxCapacity}
							placeholder={$_('events.seriesNew.maxCapacityPlaceholder')}
							helper={$_('events.seriesNew.maxCapacityHelper')}
							error={errors.maxCapacity || ''}
						/>
					</div>

					<div class="pt-2">
						{#if showRetention}
							<Input
								label={$_('events.seriesNew.retentionLabel')}
								name="retentionDays"
								type="number"
								bind:value={retentionDays}
								helper={$_('events.seriesNew.retentionHelper')}
								error={errors.retentionDays || ''}
							/>
						{:else}
							<p class="text-xs text-neutral-400">
								{$_('events.seriesNew.retentionDefaultNote')}
								<button
									type="button"
									class="text-primary hover:text-primary-hover underline underline-offset-2"
									onclick={() => (showRetention = true)}
								>
									{$_('events.seriesNew.retentionCustomize')}
								</button>
							</p>
						{/if}
					</div>
				</div>
			</Card>

			<div class="flex items-center justify-end gap-3">
				<Button variant="outline" href="/events/series">{$_('events.seriesNew.cancel')}</Button>
				<Button type="submit" loading={submitting}>{$_('events.seriesNew.createSeries')}</Button>
			</div>
		</form>
	</div>
</AppShell>
