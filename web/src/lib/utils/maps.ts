export type MapProvider = 'none' | 'google' | 'osm' | 'custom';

// Returns the map link URL for the given provider/address, or null when
// the organizer has turned map links off for this event. For "custom", the
// organizer's own URL is returned as-is (empty/missing yields null).
export function getMapUrl(
	location: string,
	provider: MapProvider | undefined,
	customUrl?: string
): string | null {
	switch (provider) {
		case 'google':
			return location ? `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(location)}` : null;
		case 'osm':
			return location ? `https://www.openstreetmap.org/search?query=${encodeURIComponent(location)}` : null;
		case 'custom':
			return customUrl || null;
		default:
			return null;
	}
}
