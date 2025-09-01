import {
	array,
	boolean,
	number,
	object,
	optional,
	picklist,
	string,
	type InferOutput
} from 'valibot';
import { BaseSchema } from './base-model';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Asset type schema
const AssetTypeSchema = picklist(['video', 'html', 'pdf', 'markdown', 'text']);
export type AssetType = InferOutput<typeof AssetTypeSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const AssetVideoMetadataSchema = object({
	duration: number(),
	width: number(),
	height: number(),
	codec: string(),
	resolution: string()
});

export type AssetVideoMetadataModel = InferOutput<typeof AssetVideoMetadataSchema>;

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
	courseId: string(),
	lessonId: string(),
	title: string(),
	prefix: number(),
	subPrefix: optional(number()),
	subTitle: optional(string()),
	module: string(),
	path: string(),
	assetType: AssetTypeSchema,
	videoMetadata: optional(AssetVideoMetadataSchema),
	progress: optional(AssetProgressSchema)
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
