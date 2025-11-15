import { boolean, number, object, string, type InferOutput } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const MediaPreferenceDefault: MediaPreferencesModel = {
	autoplay: false,
	autoloadNext: false,
	playbackRate: 1.0,
	volume: 1.0,
	muted: false,
	quality: 'original'
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const MediaPreferences = object({
	autoplay: boolean(),
	autoloadNext: boolean(),
	playbackRate: number(),
	volume: number(),
	muted: boolean(),
	quality: string()
});

export type MediaPreferencesModel = InferOutput<typeof MediaPreferences>;
