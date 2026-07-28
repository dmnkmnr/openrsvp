<script lang="ts">
	import { page } from '$app/stores';
	import { api } from '$lib/api/client';
	import { toast } from '$lib/stores/toast';
	import type { CSVImportRow, CSVPreviewResponse, CSVImportResult, ApiError } from '$lib/types';
	import AppShell from '$lib/components/layout/AppShell.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Spinner from '$lib/components/ui/Spinner.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import { _, locale } from '$lib/i18n';

	const MAX_FILE_SIZE = 1024 * 1024; // 1MB

	const eventId = $derived($page.params.eventId);

	type Step = 'upload' | 'preview' | 'confirm' | 'result';

	let step = $state<Step>('upload');
	let selectedFile = $state<File | null>(null);
	let preview = $state<CSVPreviewResponse | null>(null);
	let result = $state<CSVImportResult | null>(null);
	let uploading = $state(false);
	let importing = $state(false);
	let error = $state('');
	let sendInvitations = $state(false);
	let dragging = $state(false);
	let fileInput = $state<HTMLInputElement | null>(null);

	const validRows = $derived(
		preview ? preview.rows.filter((r) => !r.error && !r.duplicate) : []
	);

	const stepNumber = $derived(
		step === 'upload' ? 1 : step === 'preview' ? 2 : step === 'confirm' ? 3 : 4
	);

	function validateFile(file: File): string | null {
		if (!file.name.toLowerCase().endsWith('.csv') && file.type !== 'text/csv') {
			return $_('events.import.invalidFileType');
		}
		if (file.size > MAX_FILE_SIZE) {
			const sizeMB = (file.size / (1024 * 1024)).toFixed(1);
			return $_('events.import.fileTooLarge', { values: { size: sizeMB } });
		}
		return null;
	}

	function handleFileSelect(file: File) {
		const validationError = validateFile(file);
		if (validationError) {
			error = validationError;
			selectedFile = null;
			return;
		}
		error = '';
		selectedFile = file;
		uploadFile(file);
	}

	function handleInputChange(e: Event) {
		const input = e.target as HTMLInputElement;
		const file = input.files?.[0];
		if (file) {
			handleFileSelect(file);
		}
	}

	function handleDragOver(e: DragEvent) {
		e.preventDefault();
		dragging = true;
	}

	function handleDragLeave(e: DragEvent) {
		e.preventDefault();
		dragging = false;
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragging = false;
		const file = e.dataTransfer?.files?.[0];
		if (file) {
			handleFileSelect(file);
		}
	}

	async function uploadFile(file: File) {
		uploading = true;
		error = '';
		preview = null;
		result = null;

		try {
			const res = await api.uploadCSV<{ data: CSVPreviewResponse }>(
				`/rsvp/event/${eventId}/import/preview`,
				file
			);
			preview = res.data;
			step = 'preview';
		} catch (err) {
			const apiErr = err as ApiError;
			error = apiErr.message || $_('events.import.parseError');
			step = 'upload';
		} finally {
			uploading = false;
		}
	}

	function proceedToConfirm() {
		step = 'confirm';
	}

	function backToPreview() {
		step = 'preview';
	}

	async function confirmImport() {
		if (!preview) return;
		importing = true;
		error = '';

		try {
			const res = await api.post<{ data: CSVImportResult }>(
				`/rsvp/event/${eventId}/import`,
				{
					rows: validRows,
					sendInvitations
				}
			);
			result = res.data;
			step = 'result';
			toast.success($_('events.import.importedToast', { values: { count: res.data.imported, plural: res.data.imported !== 1 ? 's' : '' } }));
		} catch (err) {
			const apiErr = err as ApiError;
			error = apiErr.message || $_('events.import.importError');
		} finally {
			importing = false;
		}
	}

	function startOver() {
		step = 'upload';
		selectedFile = null;
		preview = null;
		result = null;
		error = '';
		sendInvitations = false;
		if (fileInput) {
			fileInput.value = '';
		}
	}

	async function downloadTemplate() {
		try {
			const response = await fetch(`/api/v1/rsvp/import/template?lang=${$locale ?? 'en'}`);
			if (!response.ok) throw new Error('Failed to download template');
			const blob = await response.blob();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'guest_import_template.csv';
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
		} catch {
			toast.error($_('events.import.templateDownloadError'));
		}
	}
</script>

<svelte:head>
	<title>{$_('events.import.pageTitle')}</title>
</svelte:head>

<AppShell>
	<div class="mb-6">
		<a href="/events/{eventId}" class="text-sm text-primary hover:text-primary-hover font-medium">
			&larr; {$_('events.import.backToEvent')}
		</a>
	</div>

	<!-- Step indicator -->
	<div class="mb-8">
		<nav aria-label={$_('events.import.progressLabel')}>
			<ol class="flex items-center gap-2 text-sm">
				{#each [{ num: 1, label: $_('events.import.stepUpload') }, { num: 2, label: $_('events.import.stepPreview') }, { num: 3, label: $_('events.import.stepConfirm') }, { num: 4, label: $_('events.import.stepResults') }] as s}
					<li class="flex items-center gap-2">
						<span
							class="inline-flex h-6 w-6 items-center justify-center rounded-full text-xs font-medium
								{stepNumber > s.num
									? 'bg-primary text-white'
									: stepNumber === s.num
										? 'bg-primary-lighter text-primary ring-2 ring-primary'
										: 'bg-neutral-100 text-neutral-400'}"
						>
							{#if stepNumber > s.num}
								<svg class="h-3.5 w-3.5" viewBox="0 0 20 20" fill="currentColor">
									<path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
								</svg>
							{:else}
								{s.num}
							{/if}
						</span>
						<span
							class="font-medium {stepNumber >= s.num ? 'text-neutral-900' : 'text-neutral-400'}"
						>
							{s.label}
						</span>
					</li>
					{#if s.num < 4}
						<li class="flex-1 mx-1">
							<div class="h-px {stepNumber > s.num ? 'bg-primary' : 'bg-neutral-200'}"></div>
						</li>
					{/if}
				{/each}
			</ol>
		</nav>
	</div>

	<!-- Error banner -->
	{#if error}
		<div class="mb-6 rounded-lg bg-error-light border border-error px-4 py-3 text-sm text-error flex items-start gap-2">
			<svg class="h-5 w-5 text-error shrink-0 mt-0" viewBox="0 0 20 20" fill="currentColor">
				<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
			</svg>
			<span>{error}</span>
		</div>
	{/if}

	<!-- Step 1: Upload -->
	{#if step === 'upload'}
		<Card>
			{#snippet header()}
				<div class="flex items-center justify-between">
					<h1 class="text-xl font-bold font-display text-neutral-900">{$_('events.import.uploadTitle')}</h1>
					<button
						onclick={downloadTemplate}
						class="text-sm text-primary hover:text-primary font-medium"
					>
						{$_('events.import.downloadTemplate')}
					</button>
				</div>
			{/snippet}

			<div class="space-y-6">
				<p class="text-sm text-neutral-600">
					{$_('events.import.uploadBody', { values: { name: $_('events.import.nameColumn') } })}
				</p>

				<!-- Drag & drop zone -->
				<div
					role="button"
					tabindex="0"
					aria-label={$_('events.import.dropZoneLabel')}
					class="relative rounded-lg border-2 border-dashed p-8 text-center transition-colors
						{dragging
							? 'border-primary-light bg-primary-lighter'
							: 'border-neutral-300 hover:border-neutral-400 hover:bg-neutral-50'}"
					ondragover={handleDragOver}
					ondragleave={handleDragLeave}
					ondrop={handleDrop}
					onclick={() => fileInput?.click()}
					onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); fileInput?.click(); } }}
				>
					{#if uploading}
						<div class="flex flex-col items-center gap-3">
							<Spinner size="lg" class="text-primary" />
							<p class="text-sm text-neutral-600">{$_('events.import.parsingFile')}</p>
						</div>
					{:else}
						<div class="flex flex-col items-center gap-3">
							<svg class="h-10 w-10 text-neutral-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5" />
							</svg>
							<div>
								<p class="text-sm font-medium text-neutral-700">
									{#if dragging}
										{$_('events.import.dropHere')}
									{:else}
										{$_('events.import.dragDropHere')}
									{/if}
								</p>
								<p class="mt-1 text-xs text-neutral-500">{$_('events.import.clickToBrowse')}</p>
							</div>
						</div>
					{/if}
					<input
						bind:this={fileInput}
						type="file"
						accept=".csv,text/csv"
						onchange={handleInputChange}
						class="hidden"
						aria-hidden="true"
					/>
				</div>
			</div>
		</Card>

	<!-- Step 2: Preview -->
	{:else if step === 'preview' && preview}
		<Card>
			{#snippet header()}
				<div class="flex items-center justify-between">
					<h1 class="text-xl font-bold font-display text-neutral-900">{$_('events.import.previewTitle')}</h1>
					<span class="text-sm text-neutral-500">
						{selectedFile?.name}
					</span>
				</div>
			{/snippet}

			<div class="space-y-6">
				<!-- Summary stats -->
				<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
					<div class="rounded-lg bg-neutral-50 border border-neutral-200 p-3 text-center">
						<p class="text-2xl font-bold text-neutral-900">{preview.totalRows}</p>
						<p class="text-xs text-neutral-500 mt-1">{$_('events.import.totalRows')}</p>
					</div>
					<div class="rounded-lg bg-success-light border border-success p-3 text-center">
						<p class="text-2xl font-bold text-success">{preview.validRows}</p>
						<p class="text-xs text-success mt-1">{$_('events.import.valid')}</p>
					</div>
					<div class="rounded-lg bg-error-light border border-error p-3 text-center">
						<p class="text-2xl font-bold text-error">{preview.errorRows}</p>
						<p class="text-xs text-error mt-1">{$_('events.import.errors')}</p>
					</div>
					<div class="rounded-lg bg-warning-light border border-warning p-3 text-center">
						<p class="text-2xl font-bold text-warning">{preview.duplicates}</p>
						<p class="text-xs text-warning mt-1">{$_('events.import.duplicates')}</p>
					</div>
				</div>

				<!-- Preview table -->
				<div class="border border-neutral-200 rounded-lg overflow-hidden">
					<div class="overflow-x-auto max-h-96">
						<table class="w-full text-sm">
							<thead class="bg-neutral-50 sticky top-0">
								<tr>
									<th class="px-4 py-2.5 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">{$_('events.import.statusHeader')}</th>
									<th class="px-4 py-2.5 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">{$_('events.import.nameHeader')}</th>
									<th class="px-4 py-2.5 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">{$_('events.import.emailHeader')}</th>
									<th class="px-4 py-2.5 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">{$_('events.import.phoneHeader')}</th>
									<th class="px-4 py-2.5 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">{$_('events.import.dietaryHeader')}</th>
									<th class="px-4 py-2.5 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">{$_('events.import.plusOnesHeader')}</th>
									<th class="px-4 py-2.5 text-left text-xs font-medium text-neutral-500 uppercase tracking-wider">{$_('events.import.childrenHeader')}</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-neutral-100">
								{#each preview.rows as row}
									<tr class="{row.error ? 'bg-error-light' : row.duplicate ? 'bg-warning-light' : 'hover:bg-neutral-50'}">
										<td class="px-4 py-2.5 whitespace-nowrap">
											{#if row.error}
												<Badge variant="error">{$_('events.import.error')}</Badge>
											{:else if row.duplicate}
												<Badge variant="warning">{$_('events.import.duplicate')}</Badge>
											{:else}
												<Badge variant="success">{$_('events.import.valid')}</Badge>
											{/if}
										</td>
										<td class="px-4 py-2.5 text-neutral-900 font-medium">{row.name || '—'}</td>
										<td class="px-4 py-2.5 text-neutral-600">{row.email || '—'}</td>
										<td class="px-4 py-2.5 text-neutral-600">{row.phone || '—'}</td>
										<td class="px-4 py-2.5 text-neutral-600">{row.dietaryNotes || '—'}</td>
										<td class="px-4 py-2.5 text-neutral-600">{row.plusOnes || 0}</td>
										<td class="px-4 py-2.5 text-neutral-600">{row.plusOnesChildren || 0}</td>
									</tr>
									{#if row.error}
										<tr class="bg-error-light">
											<td colspan="7" class="px-4 py-1.5 text-xs text-error italic">
												{row.error}
											</td>
										</tr>
									{/if}
								{/each}
							</tbody>
						</table>
					</div>
				</div>

				<!-- Actions -->
				<div class="flex items-center justify-between">
					<Button variant="outline" onclick={startOver}>
						{$_('events.import.chooseDifferentFile')}
					</Button>
					{#if preview.validRows > 0}
						<Button onclick={proceedToConfirm}>
							{$_('events.import.continueWith', { values: { count: preview.validRows, plural: preview.validRows !== 1 ? 's' : '' } })}
						</Button>
					{:else}
						<p class="text-sm text-error font-medium">
							{$_('events.import.noValidRows')}
						</p>
					{/if}
				</div>
			</div>
		</Card>

	<!-- Step 3: Confirm -->
	{:else if step === 'confirm' && preview}
		<Card>
			{#snippet header()}
				<h1 class="text-xl font-bold font-display text-neutral-900">{$_('events.import.confirmTitle')}</h1>
			{/snippet}

			<div class="space-y-6">
				<div class="rounded-lg bg-primary-lighter border border-primary-light p-4">
					<p class="text-sm text-primary">
						{$_('events.import.confirmBody', { values: { count: preview.validRows, plural: preview.validRows !== 1 ? 's' : '' } })}
						{#if preview.errorRows > 0}
							{$_('events.import.errorsSkipped', { values: { count: preview.errorRows, plural: preview.errorRows !== 1 ? 's' : '' } })}
						{/if}
						{#if preview.duplicates > 0}
							{$_('events.import.duplicatesSkipped', { values: { count: preview.duplicates, plural: preview.duplicates !== 1 ? 's' : '' } })}
						{/if}
					</p>
				</div>

				<div class="border border-neutral-200 rounded-lg p-4">
					<label class="flex items-start gap-3 cursor-pointer">
						<input
							type="checkbox"
							bind:checked={sendInvitations}
							class="mt-0.5 rounded border-neutral-300 text-primary focus:ring-primary/40"
						/>
						<div>
							<span class="text-sm font-medium text-neutral-900">{$_('events.import.sendInvitationsLabel')}</span>
							<p class="text-xs text-neutral-500 mt-0.5">
								{$_('events.import.sendInvitationsHelper')}
							</p>
						</div>
					</label>
				</div>

				<div class="flex items-center justify-between">
					<Button variant="outline" onclick={backToPreview}>
						{$_('events.import.backToPreview')}
					</Button>
					<Button onclick={confirmImport} loading={importing}>
						{#if importing}
							{$_('events.import.importing')}
						{:else}
							{$_('events.import.importButton', { values: { count: preview.validRows, plural: preview.validRows !== 1 ? 's' : '' } })}
						{/if}
					</Button>
				</div>
			</div>
		</Card>

	<!-- Step 4: Results -->
	{:else if step === 'result' && result}
		<Card>
			{#snippet header()}
				<h1 class="text-xl font-bold font-display text-neutral-900">{$_('events.import.resultTitle')}</h1>
			{/snippet}

			<div class="space-y-6">
				<div class="rounded-lg bg-success-light border border-success p-6">
					<div class="flex items-center gap-2 mb-4">
						<svg class="h-6 w-6 text-success" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
						</svg>
						<h3 class="text-lg font-semibold font-display text-success">{$_('events.import.importSuccessTitle')}</h3>
					</div>
					<div class="grid grid-cols-2 sm:grid-cols-5 gap-4">
						<div class="text-center">
							<p class="text-2xl font-bold text-success">{result.imported}</p>
							<p class="text-xs text-success mt-1">{$_('events.import.imported')}</p>
						</div>
						<div class="text-center">
							<p class="text-2xl font-bold text-neutral-600">{result.skipped}</p>
							<p class="text-xs text-neutral-500 mt-1">{$_('events.import.skipped')}</p>
						</div>
						<div class="text-center">
							<p class="text-2xl font-bold text-error">{result.failed}</p>
							<p class="text-xs text-error mt-1">{$_('events.import.failed')}</p>
						</div>
						<div class="text-center">
							<p class="text-2xl font-bold text-warning">{result.duplicates}</p>
							<p class="text-xs text-warning mt-1">{$_('events.import.duplicates')}</p>
						</div>
						<div class="text-center">
							<p class="text-2xl font-bold text-info">{result.invited}</p>
							<p class="text-xs text-info mt-1">{$_('events.import.invited')}</p>
						</div>
					</div>
				</div>

				<div class="flex items-center gap-3">
					<Button href="/events/{eventId}">{$_('events.import.backToEventButton')}</Button>
					<Button variant="outline" onclick={startOver}>{$_('events.import.importMore')}</Button>
				</div>
			</div>
		</Card>
	{/if}
</AppShell>
