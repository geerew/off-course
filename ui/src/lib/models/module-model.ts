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
import { AssetSchema } from './asset-model';
import { AttachmentSchema } from './attachment-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Description type schema
const DescriptionTypeSchema = picklist(['markdown', 'text']);
export type DescriptionType = InferOutput<typeof DescriptionTypeSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Lesson schema
export const LessonSchema = object({
	prefix: number(),
	title: string(),
	hasDescription: boolean(),
	descriptionType: optional(DescriptionTypeSchema),
	assets: array(AssetSchema),
	attachments: array(AttachmentSchema),
	completed: boolean(),
	startedAssetCount: number(),
	completedAssetCount: number(),
	totalVideoDuration: number()
});

export type LessonModel = InferOutput<typeof LessonSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Module schema
export const ModuleSchema = object({
	module: string(),
	index: number(),
	lessons: array(LessonSchema)
});

export type ModuleModel = InferOutput<typeof ModuleSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Modules schema
export const ModulesSchema = object({
	modules: array(ModuleSchema)
});
export type ModulesModel = InferOutput<typeof ModulesSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type ModulesReqParams = {
	withProgress?: boolean;
};
