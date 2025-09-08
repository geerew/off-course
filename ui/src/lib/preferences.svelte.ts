import { PersistedState } from 'runed';
import { MediaPreferenceDefault, type MediaPreferencesModel } from './models/media-model';

export const mediaPreferences = new PersistedState<MediaPreferencesModel>(
	'media_preferences',
	MediaPreferenceDefault
);
