import { array, boolean, number, object, picklist, string, type InferOutput } from 'valibot';
import { AssetSchema } from './asset-model';
import { AttachmentSchema } from './attachment-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Description type schema
const DescriptionTypeSchema = picklist(['markdown', 'text']);
export type DescriptionType = InferOutput<typeof DescriptionTypeSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Lesson schema
export const LessonSchema = object({
	id: string(),
	courseId: string(),
	prefix: number(),
	title: string(),
	started: boolean(),
	completed: boolean(),
	assetsCompleted: number(),
	totalVideoDuration: number(),
	assets: array(AssetSchema),
	attachments: array(AttachmentSchema)
});

export type LessonModel = InferOutput<typeof LessonSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Module schema
export const ModuleSchema = object({
	prefix: number(),
	module: string(),
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
	withUserProgress?: boolean;
};
