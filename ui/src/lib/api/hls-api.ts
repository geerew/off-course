import { APIError } from '$lib/api-error.svelte';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// HLS Qualities Response
export interface HLSQualitiesResponse {
	qualities: string[];
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get available HLS qualities for a video
export async function GetHLSQualities(assetId: string): Promise<string[]> {
	const response = await apiFetch(`/api/hls/${assetId}/qualities`);

	if (response.ok) {
		const data = (await response.json()) as HLSQualitiesResponse;
		return data.qualities;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.error || 'Failed to get HLS qualities');
	}
}
