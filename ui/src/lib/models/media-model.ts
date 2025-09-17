import { boolean, number, object, type InferOutput } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const MediaPreferenceDefault: MediaPreferencesModel = {
	autoplay: false,
	autoloadNext: false,
	playbackRate: 1.0,
	volume: 1.0,
	muted: false
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const MediaPreferences = object({
	autoplay: boolean(),
	autoloadNext: boolean(),
	playbackRate: number(),
	volume: number(),
	muted: boolean()
});

export type MediaPreferencesModel = InferOutput<typeof MediaPreferences>;
