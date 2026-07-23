<script lang="ts">
	import { locale, _, SUPPORTED_LOCALES } from '$lib/i18n';

	const labels: Record<string, string> = { en: 'English', de: 'Deutsch' };
</script>

<select
	bind:value={$locale}
	aria-label={$_('nav.language')}
	class="rounded-md border-none bg-transparent p-2 text-sm text-neutral-400 hover:text-neutral-700 focus:outline-none focus:ring-2 focus:ring-primary transition-colors duration-short ease-out"
>
	{#each SUPPORTED_LOCALES as l (l)}
		<option value={l}>{labels[l]}</option>
	{/each}
</select>
