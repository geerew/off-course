import { APIError } from '$lib/api-error.svelte';
import {
	TagPaginationSchema,
	TagSchema,
	type TagCreateModel,
	type TagModel,
	type TagPaginationModel,
	type TagReqParams,
	type TagUpdateModel
} from '$lib/models/tag-model';
import { buildQueryString } from '$lib/utils';
import { safeParse } from 'valibot';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a paginated list of tags
export async function GetTags(params?: TagReqParams): Promise<TagPaginationModel> {
	const qs = params && buildQueryString(params);

	const response = await apiFetch('/api/tags' + (qs ? `?${qs}` : ''));

	if (response.ok) {
		const data = (await response.json()) as TagPaginationModel;
		const result = safeParse(TagPaginationSchema, data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a tag (by name)
export async function GetTag(name: string): Promise<TagModel> {
	const response = await apiFetch(`/api/tags/${name}`);

	if (response.ok) {
		const data = (await response.json()) as TagModel;
		const result = safeParse(TagSchema, data);

		if (!result.success) throw new Error('Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a list of tag names
export async function GetTagNames(params?: TagReqParams): Promise<string[]> {
	const qs = params && buildQueryString(params);

	const response = await apiFetch('/api/tags/names' + (qs ? `?${qs}` : ''));

	if (response.ok) {
		const data = (await response.json()) as string[];
		return data;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create a tag
export async function CreateTag(data: TagCreateModel): Promise<void> {
	const response = await apiFetch('/api/tags', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update a tag
export async function UpdateTag(tagId: string, data: TagUpdateModel): Promise<void> {
	const response = await apiFetch(`/api/tags/${tagId}`, {
		method: 'PUT',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify(data)
	});

	if (!response.ok) {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete a tag
export async function DeleteTag(tagId: string): Promise<void> {
	const response = await apiFetch(`/api/tags/${tagId}`, {
		method: 'DELETE',
		headers: {
			'Content-Type': 'application/json'
		}
	});

	if (!response.ok) {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}
