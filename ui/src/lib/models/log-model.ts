import { array, number, object, string, unknown, type InferOutput } from 'valibot';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Log schema - Note: updatedAt is not included as logs are immutable
export const LogSchema = object({
	id: string(),
	createdAt: string(),
	level: number(),
	message: string(),
	data: unknown() // JSON map - can contain any values
});

export type LogModel = InferOutput<typeof LogSchema>;
export type LogsModel = LogModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const LogPaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(LogSchema)
});

export type LogPaginationModel = InferOutput<typeof LogPaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type LogReqParams = PaginationReqParams & {
	q?: string;
};
