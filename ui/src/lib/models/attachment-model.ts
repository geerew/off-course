import { object, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Attachment schema
export const AttachmentSchema = object({
	...BaseSchema.entries,
	lessonId: string(),
	title: string(),
	path: string()
});

export type AttachmentModel = InferOutput<typeof AttachmentSchema>;
export type AttachmentsModel = AttachmentModel[];
