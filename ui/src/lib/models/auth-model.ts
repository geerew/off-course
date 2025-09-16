import { boolean, object, string, type InferOutput } from 'valibot';

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Signup status schema
export const SignupStatusSchema = object({
	enabled: boolean()
});

export type SignupStatusModel = InferOutput<typeof SignupStatusSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Auth register create schema
export const AuthRegisterSchema = object({
	username: string(),
	password: string()
});

export type AuthRegisterModel = InferOutput<typeof AuthRegisterSchema>;

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// Auth login schema
export const AuthLoginSchema = object({
	username: string(),
	password: string()
});

export type AuthLoginModel = InferOutput<typeof AuthLoginSchema>;
