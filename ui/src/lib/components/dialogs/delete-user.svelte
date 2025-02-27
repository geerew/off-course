<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { DeleteSelf } from '$lib/api/self-api';
	import { DeleteUser } from '$lib/api/user-api';
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { AlertDialog, Button, InputPassword } from '$lib/components/ui';
	import type { UserModel, UsersModel } from '$lib/models/user-model';
	import { Separator } from 'bits-ui';
	import { type Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open?: boolean;
		value: UserModel | UsersModel;
		trigger?: Snippet;
		triggerClass?: string;
		successFn?: () => void;
	};

	let { open = $bindable(false), value, trigger, triggerClass, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let currentInputEl = $state<HTMLInputElement>();
	let currentPassword = $state('');
	let isPosting = $state(false);

	const isArray = Array.isArray(value);
	const deletingSelf = isArray ? false : value.id === auth?.user?.id;

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			currentPassword = '';
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doDelete(): Promise<void> {
		isPosting = true;

		try {
			if (isArray) {
				await Promise.all(Object.values(value).map((u) => DeleteUser(u.id)));
				toast.success('Selected users deleted');
			} else {
				if (deletingSelf) {
					await DeleteSelf({ currentPassword });
					auth.empty();
					window.location.href = '/auth/login';
				} else {
					await DeleteUser(value.id);
				}
			}

			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
		open = false;
	}
</script>

<AlertDialog
	bind:open
	contentProps={{
		interactOutsideBehavior: 'close',
		onOpenAutoFocus: (e) => {
			e.preventDefault();
			currentInputEl?.focus();
		},
		onCloseAutoFocus: (e) => {
			e.preventDefault();
		}
	}}
	{trigger}
	{triggerClass}
>
	{#snippet description()}
		<div class="text-foreground-alt-1 flex flex-col gap-2 text-center">
			{#if deletingSelf}
				<span class="text-lg">Are you sure you want to delete your account?</span>
			{:else if isArray && Object.values(value).length > 1}
				<span class="text-lg"> Are you sure you want to delete these users? </span>
			{:else}
				<span class="text-lg">Are you sure you want to delete this user?</span>
			{/if}
			<span class="text-foreground-alt-2">All associated data will be deleted</span>
		</div>

		{#if deletingSelf}
			<Separator.Root class="bg-background-alt-3 mt-2 h-px w-full shrink-0" />

			<div class="flex flex-col gap-2.5 px-2.5">
				<div>Confirm Password:</div>
				<InputPassword
					bind:ref={currentInputEl}
					bind:value={currentPassword}
					name="current password"
				/>
			</div>
		{/if}
	{/snippet}

	{#snippet action()}
		<Button
			disabled={(deletingSelf && currentPassword === '') || isPosting}
			onclick={doDelete}
			class="bg-background-error disabled:bg-background-error/80 enabled:hover:bg-background-error-alt-1 text-foreground-alt-1 enabled:hover:text-foreground w-24"
		>
			{#if isPosting}
				<Spinner class="bg-foreground-alt-1 size-2" />
			{:else}
				Delete
			{/if}
		</Button>
	{/snippet}
</AlertDialog>
