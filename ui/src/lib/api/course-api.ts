import { APIError } from '$lib/api-error.svelte';
import {
	AssetPaginationSchema,
	type AssetModel,
	type AssetPaginationModel,
	type AssetProgressModel,
	type AssetReqParams
} from '$lib/models/asset-model';
import {
	CoursePaginationSchema,
	CourseSchema,
	CourseTagSchema,
	type CourseModel,
	type CoursePaginationModel,
	type CourseReqParams,
	type CourseTagsModel,
	type CreateCourseModel,
	type CreateCourseTagModel
} from '$lib/models/course-model';
import { ModulesSchema, type ModulesModel, type ModulesReqParams } from '$lib/models/module-model';
import { buildQueryString } from '$lib/utils';
import { array, safeParse } from 'valibot';
import { apiFetch } from './fetch';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a paginated list of courses
export async function GetCourses(params?: CourseReqParams): Promise<CoursePaginationModel> {
	const qs = params && buildQueryString(params);

	const response = await apiFetch('/api/courses' + (qs ? `?${qs}` : ''));

	if (response.ok) {
		const data = (await response.json()) as CoursePaginationModel;
		const result = safeParse(CoursePaginationSchema, data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create a course
export async function CreateCourse(data: CreateCourseModel): Promise<void> {
	const response = await apiFetch('/api/courses', {
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

// Delete a course
export async function DeleteCourse(id: string): Promise<void> {
	const response = await apiFetch(`/api/courses/${id}`, {
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete course progress (for a specific user)
export async function DeleteCourseProgress(id: string): Promise<void> {
	const response = await apiFetch(`/api/courses/${id}/progress`, {
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a course
export async function GetCourse(id: string): Promise<CourseModel> {
	const response = await apiFetch(`/api/courses/${id}`);

	if (response.ok) {
		const data = (await response.json()) as CourseModel;
		const result = safeParse(CourseSchema, data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get all course tags
export async function GetCourseTags(id: string): Promise<CourseTagsModel> {
	const response = await apiFetch(`/api/courses/${id}/tags`);

	if (response.ok) {
		const data = (await response.json()) as CourseTagsModel;
		const result = safeParse(array(CourseTagSchema), data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create a course tag
export async function CreateCourseTag(id: string, data: CreateCourseTagModel): Promise<void> {
	const response = await apiFetch(`/api/courses/${id}/tags`, {
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

// Delete a course tag
export async function DeleteCourseTag(id: string, tag: string): Promise<void> {
	const response = await apiFetch(`/api/courses/${id}/tags/${tag}`, {
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

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a paginated list of assets for a course
export async function GetCourseAssets(
	id: string,
	params?: AssetReqParams
): Promise<AssetPaginationModel> {
	const qs = params && buildQueryString(params);
	const response = await apiFetch(`/api/courses/${id}/assets` + (qs ? `?${qs}` : ''));

	if (response.ok) {
		const data = (await response.json()) as AssetPaginationModel;
		const result = safeParse(AssetPaginationSchema, data);

		if (!result.success) throw new APIError(response.status, 'Invalid response from the server');
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get all the assets for a course
export async function GetAllCourseAssets(
	courseId: string,
	params?: AssetReqParams
): Promise<AssetModel[]> {
	let assets: AssetModel[] = [];
	let page = 1;
	let totalPages = 1;

	do {
		try {
			const response = await GetCourseAssets(courseId, { ...params, page, perPage: 100 });

			if (response.totalItems > 0) {
				assets.push(...response.items);
				totalPages = response.totalPages;
				page++;
			} else {
				break;
			}
		} catch (error) {
			throw error;
		}
	} while (page <= totalPages);

	return assets;
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update a course assets progress
export async function UpdateCourseAssetProgress(
	courseId: string,
	assetId: string,
	data: AssetProgressModel
): Promise<void> {
	const response = await apiFetch(`/api/courses/${courseId}/assets/${assetId}/progress`, {
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

// Serve a course asset
export async function ServeCourseAsset(courseId: string, assetId: string): Promise<string> {
	const response = await apiFetch(`/api/courses/${courseId}/assets/${assetId}/serve`);

	if (response.ok) {
		return await response.text();
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Serve the description of a course asset
export async function ServeCourseAssetDescription(
	courseId: string,
	assetId: string
): Promise<string> {
	const response = await apiFetch(`/api/courses/${courseId}/assets/${assetId}/description`);

	if (response.ok) {
		return await response.text();
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get the structured modules (chapters/lessons) for a course
export async function GetCourseModules(
	courseId: string,
	params?: ModulesReqParams
): Promise<ModulesModel> {
	const qs = params ? buildQueryString(params) : '';
	const response = await apiFetch(`/api/courses/${courseId}/modules${qs ? `?${qs}` : ''}`);

	if (response.ok) {
		const data = (await response.json()) as ModulesModel;
		const result = safeParse(ModulesSchema, data);

		if (!result.success) {
			throw new APIError(response.status, 'Invalid response from the server');
		}
		return result.output;
	} else {
		const data = await response.json();
		throw new APIError(response.status, data.message || 'Unknown error');
	}
}
