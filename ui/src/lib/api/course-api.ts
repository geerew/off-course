import { APIError } from '$lib/api-error.svelte';
import {
	CoursePaginationSchema,
	CourseTagSchema,
	type CoursePaginationModel,
	type CourseReqParams,
	type CourseTagsModel,
	type CreateCourseModel,
	type CreateCourseTagModel
} from '$lib/models/course-model';
import { buildQueryString } from '$lib/utils';
import { array, safeParse } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Get a paginated list of courses
export async function GetCourses(params?: CourseReqParams): Promise<CoursePaginationModel> {
	const qs = params && buildQueryString(params);

	const response = await fetch('/api/courses' + (qs ? `?${qs}` : ''));

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
	const response = await fetch('/api/courses', {
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
	const response = await fetch(`/api/courses/${id}`, {
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

// Get all course tags
export async function GetCourseTags(id: string): Promise<CourseTagsModel> {
	const response = await fetch(`/api/courses/${id}/tags`);

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
	const response = await fetch(`/api/courses/${id}/tags`, {
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
	const response = await fetch(`/api/courses/${id}/tags/${tag}`, {
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
