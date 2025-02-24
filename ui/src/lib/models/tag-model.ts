import { array, number, object, optional, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const CourseTagSchema = object({
	id: string(),
	tag: string(),
	courseId: string(),
	title: string()
});

export type CourseTagModel = InferOutput<typeof CourseTagSchema>;
export type CourseTagsModel = CourseTagModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag schema
export const TagSchema = object({
	...BaseSchema.entries,
	tag: string(),
	courseCount: number(),
	courses: optional(array(CourseTagSchema))
});

export type TagModel = InferOutput<typeof TagSchema>;
export type TagsModel = TagModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create tag schema
export const CreateTagSchema = object({
	tag: string()
});

export type CreateTagModel = InferOutput<typeof CreateTagSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update tag schema
export const UpdateTagSchema = object({
	tag: string()
});

export type UpdateTagModel = InferOutput<typeof UpdateTagSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const TagPaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(TagSchema)
});

export type TagPaginationModel = InferOutput<typeof TagPaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type TagReqParams = PaginationReqParams & {
	orderBy?: string;
};
