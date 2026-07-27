export type MapProvider = 'none' | 'google' | 'osm';

// Returns the map search URL for the given provider/address, or null when
// the organizer has turned map links off for this event.
export function getMapUrl(location: string, provider: MapProvider | undefined): string | null {
	if (!location) return null;
	switch (provider) {
		case 'google':
			return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(location)}`;
		case 'osm':
			return `https://www.openstreetmap.org/search?query=${encodeURIComponent(location)}`;
		default:
			return null;
	}
}
