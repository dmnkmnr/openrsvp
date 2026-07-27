<script lang="ts">
	import { _ } from '$lib/i18n';

	let { location }: { location: string } = $props();

	const googleUrl = $derived(`https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(location)}`);
	const osmUrl = $derived(`https://www.openstreetmap.org/search?query=${encodeURIComponent(location)}`);
</script>

<span class="inline-flex items-center gap-2 text-xs">
	<a
		href={googleUrl}
		target="_blank"
		rel="noopener noreferrer"
		class="text-primary hover:text-primary-hover underline underline-offset-2"
	>
		{$_('common.openInGoogleMaps')}
	</a>
	<span class="text-neutral-300">·</span>
	<a
		href={osmUrl}
		target="_blank"
		rel="noopener noreferrer"
		class="text-primary hover:text-primary-hover underline underline-offset-2"
	>
		{$_('common.openInOpenStreetMap')}
	</a>
</span>
