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

const UserRoleSchema = picklist(['admin', 'user']);
export type UserRole = InferOutput<typeof UserRoleSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const SelectUserRoles = [
	{ value: 'user', label: 'User' },
	{ value: 'admin', label: 'Admin' }
];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const CourseProgressSchema = object({
	started: boolean(),
	startedAt: string(),
	percent: number(),
	completedAt: string(),
	progressUpdatedAt: string()
});

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course schema
export const CourseSchema = object({
	...BaseSchema.entries,
	title: string(),
	path: optional(string()),
	hasCard: boolean(),
	available: boolean(),
	duration: number(),
	initialScan: optional(boolean()),
	maintenance: boolean(),
	progress: optional(CourseProgressSchema)
});

export type CourseModel = InferOutput<typeof CourseSchema>;
export type CoursesModel = CourseModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create course schema
export const CreateCourseSchema = object({
	title: string(),
	path: string()
});

export type CreateCourseModel = InferOutput<typeof CreateCourseSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create course tag schema
export const CreateCourseTagSchema = object({
	tag: string()
});

export type CreateCourseTagModel = InferOutput<typeof CreateCourseTagSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Course tag schema
export const CourseTagSchema = object({
	id: string(),
	tag: string()
});

export type CourseTagModel = InferOutput<typeof CourseTagSchema>;
export type CourseTagsModel = CourseTagModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const CoursePaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(CourseSchema)
});

export type CoursePaginationModel = InferOutput<typeof CoursePaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type CourseReqParams = PaginationReqParams & {
	q?: string;
	available?: boolean;
};
