<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { api } from '$lib/api/client';
	import { currentUser } from '$lib/stores/auth';
	import { toast } from '$lib/stores/toast';
	import { smsEnabled, supportedLanguages, loadAppConfig } from '$lib/stores/config';
	import { toISOLocal, datetimeLocalToUTC } from '$lib/utils/dates';
	import { getTimezoneOptions, getTimezoneLabel } from '$lib/utils/timezones';
	import type { Event } from '$lib/types';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Textarea from '$lib/components/ui/Textarea.svelte';
	import DateTimePicker from '$lib/components/ui/DateTimePicker.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import { _ } from '$lib/i18n';

	// Auto-fill timezone from profile or browser detection.
	const defaultTz = $currentUser?.timezone
		|| Intl.DateTimeFormat().resolvedOptions().timeZone
		|| '';

	let step = $state(1);
	let submitting = $state(false);

	// Step 1 fields
	let title = $state('');
	let eventDate = $state('');
	let endDate = $state('');
	let location = $state('');
	let timezone = $state(defaultTz);

	// Step 2 fields
	let description = $state('');
	let language = $state('en');
	let contactRequirement = $state('email');
	let showHeadcount = $state(false);
	let showGuestList = $state(false);
	let rsvpDeadline = $state('');
	let maxCapacity = $state('');
	let waitlistEnabled = $state(false);
	let retentionDays = $state('30');
	let showRetention = $state(false);

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

	const languageNames: Record<string, string> = { en: 'English', de: 'Deutsch' };
	const languageOptions = $derived(
		$supportedLanguages.map(code => ({ value: code, label: languageNames[code] || code }))
	);

	onMount(() => {
		loadAppConfig();
	});

	// Validation errors
	let errors: Record<string, string> = $state({});

	const tzOptions = getTimezoneOptions(defaultTz);

	const minDate = toISOLocal(new Date());

	function validateStep1(): boolean {
		errors = {};
		if (!title.trim()) errors.title = $_('events.new.titleRequired');
		if (!eventDate) errors.eventDate = $_('events.new.eventDateRequired');
		if (!timezone) errors.timezone = $_('events.new.timezoneRequired');
		return Object.keys(errors).length === 0;
	}

	function validateStep2(): boolean {
		errors = {};
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
		return Object.keys(errors).length === 0;
	}

	function nextStep() {
		if (step === 1 && validateStep1()) {
			step = 2;
		} else if (step === 2 && validateStep2()) {
			step = 3;
		}
	}

	function prevStep() {
		if (step > 1) step -= 1;
	}

	async function handleSubmit() {
		submitting = true;
		try {
			const body: Record<string, unknown> = {
				title: title.trim(),
				eventDate: eventDate ? datetimeLocalToUTC(eventDate, timezone) : eventDate,
				location: location.trim(),
				timezone,
				description: description.trim(),
				language,
				contactRequirement,
				showHeadcount,
				showGuestList,
				retentionDays: parseInt(retentionDays)
			};
			if (endDate) body.endDate = datetimeLocalToUTC(endDate, timezone);
			if (rsvpDeadline) body.rsvpDeadline = datetimeLocalToUTC(rsvpDeadline, timezone);
			if (maxCapacity) {
				body.maxCapacity = parseInt(maxCapacity);
				body.waitlistEnabled = waitlistEnabled;
			}

			const result = await api.post<{ data: Event }>('/events', body);
			toast.success($_('events.new.createSuccess'));
			goto(`/events/${result.data.id}/invite`);
		} catch (err: unknown) {
			const apiErr = err as { message?: string };
			toast.error(apiErr.message || $_('events.new.createError'));
		} finally {
			submitting = false;
		}
	}
</script>

<svelte:head>
	<title>{$_('events.new.pageTitle')}</title>
</svelte:head>

<AppShell>
	<div class="max-w-3xl mx-auto">
		<div class="mb-8">
			<a href="/events" class="text-sm text-primary hover:text-primary-hover">&larr; {$_('events.new.backToEvents')}</a>
			<h1 class="mt-2 text-2xl font-bold font-display text-neutral-900">{$_('events.new.heading')}</h1>
		</div>

		<!-- Step indicator -->
		<div class="mb-8">
			<div class="flex items-center">
				{#each [1, 2, 3] as s}
					<div class="flex items-center {s < 3 ? 'flex-1' : ''}">
						<div
							class="flex h-8 w-8 items-center justify-center rounded-full text-sm font-medium {s <= step
								? 'bg-primary text-white'
								: 'bg-neutral-200 text-neutral-600'}"
						>
							{s}
						</div>
						{#if s < 3}
							<div class="flex-1 mx-2 h-0.5 {s < step ? 'bg-primary' : 'bg-neutral-200'}"></div>
						{/if}
					</div>
				{/each}
			</div>
			<div class="flex justify-between mt-2 text-xs text-neutral-500">
				<span>{$_('events.new.stepDetails')}</span>
				<span>{$_('events.new.stepDescription')}</span>
				<span>{$_('events.new.stepReview')}</span>
			</div>
		</div>

		<Card>
			{#if step === 1}
				<div class="space-y-6">
					<h2 class="text-lg font-semibold font-display text-neutral-900">{$_('events.new.step1Heading')}</h2>

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
							min={minDate}
							error={errors.eventDate || ''}
							required
						/>
						<DateTimePicker
							label={$_('events.new.endDateLabel')}
							name="endDate"
							bind:value={endDate}
							min={eventDate || minDate}
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
				</div>

			{:else if step === 2}
				<div class="space-y-6">
					<h2 class="text-lg font-semibold font-display text-neutral-900">{$_('events.new.step2Heading')}</h2>

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
							min={minDate}
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
				</div>

			{:else if step === 3}
				<div class="space-y-6">
					<h2 class="text-lg font-semibold font-display text-neutral-900">{$_('events.new.step3Heading')}</h2>

					<dl class="divide-y divide-neutral-200">
						<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
							<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewTitle')}</dt>
							<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{title}</dd>
						</div>
						<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
							<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewEventDate')}</dt>
							<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{eventDate || $_('events.new.reviewNotSet')}</dd>
						</div>
						{#if endDate}
							<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
								<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewEndDate')}</dt>
								<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{endDate}</dd>
							</div>
						{/if}
						<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
							<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewLocation')}</dt>
							<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{location || $_('events.new.reviewNotSpecified')}</dd>
						</div>
						<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
							<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewTimezone')}</dt>
							<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{getTimezoneLabel(timezone)}</dd>
						</div>
						{#if description}
							<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
								<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewDescription')}</dt>
								<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0 whitespace-pre-wrap">{description}</dd>
							</div>
						{/if}
						{#if language !== 'en'}
							<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
								<dt class="text-sm font-medium text-neutral-500">{$_('events.new.languageLabel')}</dt>
								<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{languageOptions.find(o => o.value === language)?.label}</dd>
							</div>
						{/if}
						{#if contactRequirement !== 'email_or_phone'}
							<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
								<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewContactRequirement')}</dt>
								<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{contactRequirementOptions.find(o => o.value === contactRequirement)?.label}</dd>
							</div>
						{/if}
						{#if showHeadcount || showGuestList}
							<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
								<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewGuestVisibility')}</dt>
								<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">
									{#if showHeadcount && showGuestList}
										{$_('events.new.visibilityBoth')}
									{:else if showHeadcount}
										{$_('events.new.visibilityHeadcount')}
									{:else}
										{$_('events.new.visibilityGuestList')}
									{/if}
								</dd>
							</div>
						{/if}
						{#if rsvpDeadline}
							<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
								<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewRsvpDeadline')}</dt>
								<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{rsvpDeadline}</dd>
							</div>
						{/if}
						{#if maxCapacity}
							<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
								<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewMaxAttendees')}</dt>
								<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{maxCapacity}{#if waitlistEnabled} {$_('events.new.waitlistEnabledNote')}{/if}</dd>
							</div>
						{/if}
						{#if retentionDays !== '30'}
							<div class="py-3 sm:grid sm:grid-cols-3 sm:gap-4">
								<dt class="text-sm font-medium text-neutral-500">{$_('events.new.reviewRetention')}</dt>
								<dd class="mt-1 text-sm text-neutral-900 sm:col-span-2 sm:mt-0">{retentionDays} {$_('events.new.daysSuffix')}</dd>
							</div>
						{/if}
					</dl>
				</div>
			{/if}

			<!-- Navigation buttons -->
			<div class="mt-8 flex items-center justify-between border-t border-neutral-200 pt-6">
				<div>
					{#if step > 1}
						<Button variant="outline" onclick={prevStep}>{$_('events.new.back')}</Button>
					{/if}
				</div>
				<div>
					{#if step < 3}
						<Button onclick={nextStep}>{$_('events.new.next')}</Button>
					{:else}
						<Button onclick={handleSubmit} loading={submitting}>{$_('events.new.createEvent')}</Button>
					{/if}
				</div>
			</div>
		</Card>
	</div>
</AppShell>
