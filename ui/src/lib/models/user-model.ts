import { array, object, omit, partial, picklist, string, type InferOutput } from 'valibot';
import { BasePaginationSchema, type PaginationReqParams } from './pagination-model';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const UserRoleSchema = picklist(['admin', 'user']);
export type UserRole = InferOutput<typeof UserRoleSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const SelectUserRoles = [
	{ value: 'user', label: 'User' },
	{ value: 'admin', label: 'Admin' }
];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// User
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// User schema
export const UserSchema = object({
	id: string(),
	username: string(),
	displayName: string(),
	role: UserRoleSchema
});

export type UserModel = InferOutput<typeof UserSchema>;
export type UsersModel = UserModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// User create schema
export const UserCreateSchema = object({
	username: string(),
	displayName: string(),
	role: UserRoleSchema,
	password: string()
});

export type UserCreateModel = InferOutput<typeof UserCreateSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// User update schema
export const UserUpdateSchema = partial(
	object({
		...omit(UserCreateSchema, ['username']).entries,
		currentPassword: string()
	})
);

export type UserUpdateModel = InferOutput<typeof UserUpdateSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
// Self
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Self update schema
export const SelfUpdateSchema = partial(
	object({
		...omit(UserCreateSchema, ['username', 'role']).entries,
		currentPassword: string()
	})
);

export type SelfUpdateModel = InferOutput<typeof SelfUpdateSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Self delete schema
export const SelfDeleteSchema = object({
	currentPassword: string()
});

export type SelfDeleteModel = InferOutput<typeof SelfDeleteSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const UserPaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(UserSchema)
});

export type UserPaginationModel = InferOutput<typeof UserPaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type UserReqParams = PaginationReqParams & {
	q?: string;
};
