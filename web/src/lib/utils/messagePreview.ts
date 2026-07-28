import { formatDateTime } from './dates';

// Replaces {variableName} placeholders with representative sample values so
// character/SMS counters reflect the length of the actual sent message,
// not the length of the placeholder token itself. Unknown placeholders are
// left as-is.
export function interpolatePreview(text: string, variables: Record<string, string>): string {
	return text.replace(/\{(\w+)\}/g, (match, key: string) =>
		Object.prototype.hasOwnProperty.call(variables, key) ? variables[key] : match
	);
}

// Representative sample values for {guestName}, {eventTitle}, {eventDate},
// {location}, {rsvpStatus}, {rsvpLink} -- used to preview the length of an
// organizer-composed message before it's sent to any specific guest.
export function sampleMessageVariables(params: {
	locale: string;
	origin: string;
	eventTitle?: string;
	eventDate?: string;
	timezone?: string;
	location?: string;
}): Record<string, string> {
	const isDe = params.locale === 'de';
	return {
		guestName: isDe ? 'Max Mustermann' : 'John Doe',
		eventTitle: params.eventTitle || (isDe ? 'Sommerfest' : 'Summer Party'),
		eventDate: params.eventDate
			? formatDateTime(params.eventDate, params.timezone)
			: isDe
				? '12. Juni 2026 um 14:00 Uhr'
				: 'June 12, 2026 at 2:00 PM',
		location: params.location || (isDe ? 'Musterstraße 1, Berlin' : '123 Main St, Springfield'),
		rsvpStatus: isDe ? 'Zugesagt' : 'Attending',
		rsvpLink: `${params.origin}/r/AbCdEfGhIjKl`
	};
}
