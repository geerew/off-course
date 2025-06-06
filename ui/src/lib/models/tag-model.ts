import { array, number, object, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag course tag schema
export const TagCourseTagSchema = object({
	id: string(),
	title: string()
});

export type TagCourseTagModel = InferOutput<typeof TagCourseTagSchema>;
export type TagCourseTagsModel = TagCourseTagModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag schema
export const TagSchema = object({
	...BaseSchema.entries,
	tag: string(),
	courseCount: number()
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
	q?: string;
	filter?: string;
};
