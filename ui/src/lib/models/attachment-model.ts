import { array, object, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';
import { BasePaginationSchema } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment schema
export const AttachmentSchema = object({
	...BaseSchema.entries,
	title: string(),
	path: string()
});

export type AttachmentModel = InferOutput<typeof AttachmentSchema>;
export type AttachmentsModel = AttachmentModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AttachmentPaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(AttachmentSchema)
});

export type AttachmentPaginationModel = InferOutput<typeof AttachmentPaginationSchema>;
