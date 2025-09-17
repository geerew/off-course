import { array, object, picklist, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const ScanStatusSchema = picklist(['waiting', 'processing', '']);
export type ScanStatus = InferOutput<typeof ScanStatusSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan create schema
export const ScanCreateSchema = object({
	courseId: string()
});

export type ScanCreateModel = InferOutput<typeof ScanCreateSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Scan schema
export const ScanSchema = object({
	...BaseSchema.entries,
	courseId: string(),
	coursePath: string(),
	status: ScanStatusSchema
});

export type ScanModel = InferOutput<typeof ScanSchema>;
export type ScansModel = ScanModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const ScanPaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(ScanSchema)
});

export type ScanPaginationModel = InferOutput<typeof ScanPaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type ScanReqParams = PaginationReqParams & {
	q?: string;
};
