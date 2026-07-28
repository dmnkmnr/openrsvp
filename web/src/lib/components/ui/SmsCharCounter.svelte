<script lang="ts">
	import { _ } from '$lib/i18n';
	import { interpolatePreview } from '$lib/utils/messagePreview';

	// Reference length for one SMS segment. Real segment limits depend on
	// encoding (153 chars/segment for GSM-7 concatenated SMS, 67 for
	// Unicode), so this is a simple, deliberately round estimate -- not an
	// exact billing figure.
	const SMS_SEGMENT_LENGTH = 150;

	// `variables` holds representative sample values for any {placeholder}
	// tokens in `text`, so the count reflects the length of the substituted
	// value rather than the length of the placeholder name itself.
	let { text, variables = {} }: { text: string; variables?: Record<string, string> } = $props();

	const resolvedText = $derived(interpolatePreview(text, variables));
	const charCount = $derived(resolvedText.length);
	const smsCount = $derived(charCount === 0 ? 0 : Math.ceil(charCount / SMS_SEGMENT_LENGTH));
</script>

<p class="text-xs text-neutral-400 mt-1">
	{$_('common.charSmsCounter', { values: { count: charCount, sms: smsCount } })}
</p>
