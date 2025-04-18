import { array, boolean, number, object, picklist, string, type InferOutput } from 'valibot';
import { BaseSchema } from './base-model';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset type schema
const AssetTypeSchema = picklist(['video', 'html', 'pdf']);
export type AssetType = InferOutput<typeof AssetTypeSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type ChapteredAssets = Record<string, AssetsModel>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AssetProgressSchema = object({
	videoPos: number(),
	completed: boolean(),
	completedAt: string()
});

export type AssetProgressModel = InferOutput<typeof AssetProgressSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset schema
export const AssetSchema = object({
	...BaseSchema.entries,
	title: string(),
	prefix: number(),
	chapter: string(),
	path: string(),
	assetType: AssetTypeSchema,
	progress: AssetProgressSchema
});

export type AssetModel = InferOutput<typeof AssetSchema>;
export type AssetsModel = AssetModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AssetPaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(AssetSchema)
});

export type AssetPaginationModel = InferOutput<typeof AssetPaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type AssetReqParams = PaginationReqParams & {
	q?: string;
};
