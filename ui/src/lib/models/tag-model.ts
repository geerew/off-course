import { array, number, object, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag schema
export const TagSchema = object({
	...BaseSchema.entries,
	tag: string(),
	courseCount: number()
});

export type TagModel = InferOutput<typeof TagSchema>;
export type TagsModel = TagModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Tag create schema
export const TagCreateSchema = object({
	tag: string()
});

export type TagCreateModel = InferOutput<typeof TagCreateSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update tag schema
export const TagUpdateSchema = object({
	tag: string()
});

export type TagUpdateModel = InferOutput<typeof TagUpdateSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const TagPaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(TagSchema)
});

export type TagPaginationModel = InferOutput<typeof TagPaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type TagReqParams = PaginationReqParams & {
	q?: string;
	filter?: string;
};
