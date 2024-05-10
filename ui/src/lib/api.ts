import { PUBLIC_BACKEND } from '$env/static/public';
import { FileSystemSchema, type FileSystem } from '$lib/types/fileSystem';
import {
	AssetSchema,
	CourseSchema,
	CourseTagSchema,
	ScanSchema,
	TagSchema,
	type Asset,
	type AssetsGetParams,
	type Course,
	type CourseTag,
	type CoursesGetParams,
	type Scan,
	type Tag,
	type TagGetParams,
	type TagsGetParams
} from '$lib/types/models';
import { PaginationSchema, type Pagination } from '$lib/types/pagination';
import axios from 'axios';
import { array, safeParse } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const FS_API =
	process.env.NODE_ENV === 'production' ? '/api/filesystem' : `${PUBLIC_BACKEND}/api/filesystem`;

export const COURSE_API =
	process.env.NODE_ENV === 'production' ? '/api/courses' : `${PUBLIC_BACKEND}/api/courses`;

export const ASSET_API =
	process.env.NODE_ENV === 'production' ? '/api/assets' : `${PUBLIC_BACKEND}/api/assets`;

export const ATTACHMENT_API =
	process.env.NODE_ENV === 'production' ? '/api/attachments' : `${PUBLIC_BACKEND}/api/attachments`;

export const TAGS_API =
	process.env.NODE_ENV === 'production' ? '/api/tags' : `${PUBLIC_BACKEND}/api/tags`;

export const SCAN_API =
	process.env.NODE_ENV === 'production' ? '/api/scans' : `${PUBLIC_BACKEND}/api/scans`;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// FileSystem
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get information about a directory
//
// When the path is empty, the available drives are returned. When the path is populated, the
// directories and files for this path are returned
export const GetFileSystem = async (path?: string): Promise<FileSystem> => {
	try {
		let query = FS_API;
		if (path) query += `/${window.btoa(encodeURIComponent(path))}`;

		const response = await axios.get<FileSystem>(query);
		const result = safeParse(FileSystemSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve file system: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Courses
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a paginated list of courses. Use `GetAllCourses()` to get an unpaginated list of
// courses
export const GetCourses = async (params?: CoursesGetParams): Promise<Pagination> => {
	try {
		const response = await axios.get<Pagination>(COURSE_API, { params });
		const result = safeParse(PaginationSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve courses: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get all courses (not paginated). This calls GetCourses(...) until all courses are
// are returned
export const GetAllCourses = async (params?: CoursesGetParams): Promise<Course[]> => {
	let allCourses: Course[] = [];
	let page = 1;
	let totalPages = 1;

	do {
		try {
			const data = await GetCourses({ ...params, page, perPage: 100 });

			if (data.totalItems > 0) {
				allCourses = [...allCourses, ...(data.items as Course[])];
				totalPages = data.totalPages;
				page++;
			} else {
				break;
			}
		} catch (error) {
			if (axios.isAxiosError(error)) {
				throw error;
			} else {
				throw new Error(`Failed to fetch all courses: ${error}`);
			}
		}
	} while (page <= totalPages);

	return allCourses;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a course by ID
export const GetCourse = async (id: string): Promise<Course> => {
	try {
		const response = await axios.get<Course>(`${COURSE_API}/${id}`);
		const result = safeParse(CourseSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve course: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Gets the course id from the search params and queries the api for the course
export async function GetCourseFromParams(params: URLSearchParams): Promise<Course> {
	const id = params && params.get('id');
	if (!id) throw new Error('Missing course ID');

	return GetCourse(id);
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Create a course
export const AddCourse = async (title: string, path: string): Promise<Course> => {
	try {
		const response = await axios.post<Course>(
			COURSE_API,
			{ title, path },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		);
		const result = safeParse(CourseSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to create course: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PUT - Update a course
export const UpdateCourse = async (course: Course): Promise<Course> => {
	try {
		const response = await axios.put<Course>(`${COURSE_API}/${course.id}`, course, {
			headers: {
				'content-type': 'application/json'
			}
		});
		const result = safeParse(CourseSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to update course: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DELETE - Delete a course
export const DeleteCourse = async (id: string): Promise<boolean> => {
	try {
		await axios.delete(`${COURSE_API}/${id}`);
		return true;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to delete course: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Course Tags
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a list of tags for a course
export const GetCourseTags = async (courseId: string): Promise<CourseTag[]> => {
	try {
		const response = await axios.get<CourseTag[]>(`${COURSE_API}/${courseId}/tags`);
		const result = safeParse(array(CourseTagSchema), response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve course tags: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Add a tag to a course. The tag will be created if it does not exist
export const AddCourseTag = async (courseId: string, tag: string): Promise<CourseTag> => {
	try {
		const response = await axios.post<CourseTag>(
			`${COURSE_API}/${courseId}/tags/`,
			{ tag },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		);
		const result = safeParse(CourseTagSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to add course tag: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DELETE - Delete a course tag
export const DeleteCourseTag = async (courseId: string, tagId: string): Promise<boolean> => {
	try {
		await axios.delete(`${COURSE_API}/${courseId}/tags/${tagId}`);
		return true;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to delete course tag: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Assets
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a paginated list of assets for a course. Use `GetAllCourseAssets()` to get an
// unpaginated list of assets for a course
//
// Requires a course ID
export const GetCourseAssets = async (
	courseId: string,
	params?: AssetsGetParams
): Promise<Pagination> => {
	try {
		const response = await axios.get<Pagination>(`${COURSE_API}/${courseId}/assets`, { params });
		const result = safeParse(PaginationSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to get course assets: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get all assets (not paginated) for a course. This calls GetAssets(...) until all assets
// are returned
export const GetAllCourseAssets = async (
	courseId: string,
	params?: AssetsGetParams
): Promise<Asset[]> => {
	let allAssets: Asset[] = [];
	let page = 1;
	let totalPages = 1;

	do {
		try {
			const response = await GetCourseAssets(courseId, { ...params, page, perPage: 100 });

			if (response.totalItems > 0) {
				const result = safeParse(PaginationSchema, response);
				if (!result.success) throw new Error('Invalid response');

				allAssets = [...allAssets, ...(response.items as Asset[])];
				totalPages = response.totalPages;
				page++;
			} else {
				break;
			}
		} catch (error) {
			if (axios.isAxiosError(error)) {
				throw error;
			} else {
				throw new Error(`Failed to get all course assets: ${error}`);
			}
		}
	} while (page <= totalPages);

	return allAssets;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PUT - Update an asset
export const UpdateAsset = async (asset: Asset): Promise<Asset> => {
	try {
		const response = await axios.put<Asset>(`${ASSET_API}/${asset.id}`, asset, {
			headers: {
				'content-type': 'application/json'
			}
		});

		const parseResult = safeParse(AssetSchema, response.data);

		if (!parseResult.success) throw new Error('Invalid response from server');
		return parseResult.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to update asset: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Scans
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a scan by course ID
export const GetScan = async (courseId: string): Promise<Scan> => {
	try {
		const response = await axios.get<Scan>(`${SCAN_API}/${courseId}`);
		const result = safeParse(ScanSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to get scan: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Create a scan for a course
export const AddScan = async (courseId: string): Promise<Scan> => {
	try {
		const response = await axios.post<Scan>(
			SCAN_API,
			{ courseId },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		);

		const result = safeParse(ScanSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to add scan job: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Tags
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a tag by ID or name
export const GetTag = async (idOrName: string, params?: TagGetParams): Promise<Tag> => {
	try {
		const response = await axios.get<Tag>(`${TAGS_API}/${idOrName}`, { params });
		const result = safeParse(TagSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to get tag: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get a paginated list of tags
export const GetTags = async (params?: TagsGetParams): Promise<Pagination> => {
	try {
		const response = await axios.get<Pagination>(TAGS_API, { params });
		const result = safeParse(PaginationSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to retrieve tags: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// GET - Get all tags (not paginated). This calls GetTags(...) until all tags are
// are returned
export const GetAllTags = async (params?: CoursesGetParams): Promise<Tag[]> => {
	let allTags: Tag[] = [];
	let page = 1;
	let totalPages = 1;

	do {
		try {
			const data = await GetTags({ ...params, page, perPage: 100 });

			if (data.totalItems > 0) {
				allTags = [...allTags, ...(data.items as Tag[])];
				totalPages = data.totalPages;
				page++;
			} else {
				break;
			}
		} catch (error) {
			if (axios.isAxiosError(error)) {
				throw error;
			} else {
				throw new Error(`Failed to fetch all tags: ${error}`);
			}
		}
	} while (page <= totalPages);

	return allTags;
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// POST - Create a tag
export const AddTag = async (tag: string): Promise<Tag> => {
	try {
		const response = await axios.post<Tag>(
			TAGS_API,
			{ tag },
			{
				headers: {
					'content-type': 'application/json'
				}
			}
		);

		const result = safeParse(TagSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to add tag: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// PUT - Update a tag
export const UpdateTag = async (tag: Tag): Promise<Tag> => {
	try {
		const response = await axios.put<Tag>(`${TAGS_API}/${tag.id}`, tag, {
			headers: {
				'content-type': 'application/json'
			}
		});
		const result = safeParse(TagSchema, response.data);

		if (!result.success) throw new Error('Invalid response from server');
		return result.output;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to update tag: ${error}`);
		}
	}
};

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// DELETE - Delete a tag
export const DeleteTag = async (tagId: string): Promise<boolean> => {
	try {
		await axios.delete(`${TAGS_API}/${tagId}`);
		return true;
	} catch (error) {
		if (axios.isAxiosError(error)) {
			throw error;
		} else {
			throw new Error(`Failed to delete course tag: ${error}`);
		}
	}
};
