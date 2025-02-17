import { array, object, omit, partial, picklist, string, type InferOutput } from 'valibot';
import { BasePaginationSchema, type PaginationReqParams } from './pagination';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

const UserRoleSchema = picklist(['admin', 'user']);
export type UserRole = InferOutput<typeof UserRoleSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const SelectUserRoles = [
	{ value: 'user', label: 'User' },
	{ value: 'admin', label: 'Admin' }
];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// User schema
export const UserSchema = object({
	id: string(),
	username: string(),
	displayName: string(),
	role: UserRoleSchema
});

export type UserModel = InferOutput<typeof UserSchema>;
export type UsersModel = UserModel[];

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Create user schema
export const CreateUserSchema = object({
	username: string(),
	displayName: string(),
	role: UserRoleSchema,
	password: string()
});

export type CreateUserModel = InferOutput<typeof CreateUserSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update user schema
export const UpdateUserSchema = partial(
	object({
		...omit(CreateUserSchema, ['username']).entries,
		currentPassword: string()
	})
);

export type UpdateUserModel = InferOutput<typeof UpdateUserSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Update self schema
export const UpdateSelfSchema = partial(
	object({
		...omit(CreateUserSchema, ['role']).entries,
		currentPassword: string()
	})
);

export type UpdateSelfModel = InferOutput<typeof UpdateSelfSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Delete self schema
export const DeleteSelfSchema = object({
	currentPassword: string()
});

export type DeleteSelfModel = InferOutput<typeof DeleteSelfSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export const UserPaginationSchema = object({
	...BasePaginationSchema.entries,
	items: array(UserSchema)
});

export type UserPaginationModel = InferOutput<typeof UserPaginationSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

export type UserReqParams = PaginationReqParams & {
	orderBy?: string;
};
