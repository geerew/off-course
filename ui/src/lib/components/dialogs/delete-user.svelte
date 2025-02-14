<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { AlertDialog, Badge, Button, InputPassword } from '$lib/components/ui';
	import type { UserModel } from '$lib/models/user';
	import { Separator } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open?: boolean;
		user: UserModel;
		me: boolean;
		trigger?: Snippet;
		triggerClass?: string;
		successFn?: () => void;
	};

	let { open = $bindable(false), user, me, trigger, triggerClass, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let currentInputEl = $state<HTMLInputElement>();
	let currentPassword = $state('');
	let isPosting = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function deleteUser() {
		isPosting = true;

		let api = `/api/users/${user.id}`;
		let body = JSON.stringify('');

		if (me) {
			api = '/api/auth/me';
			body = JSON.stringify({
				current_password: currentPassword
			});
		}

		const response = await fetch(api, {
			method: 'DELETE',
			headers: {
				'Content-Type': 'application/json'
			},
			body
		});

		if (response.ok) {
			if (me) {
				auth.empty();
				window.location.href = '/auth/login';
			} else {
				toast.success('User deleted');
			}

			successFn?.();
		} else {
			const data = await response.json();
			toast.error(`${data.message}`);
		}

		isPosting = false;
		open = false;
	}
</script>

<AlertDialog
	bind:open
	onOpenChange={() => {
		currentPassword = '';
		isPosting = false;
	}}
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
			{#if me}
				<span class="text-lg">Are you sure you want to delete your account?</span>
			{:else}
				<span class="text-lg">Are you sure you want to delete this user?</span>

				<span>
					<Badge class="bg-background-error text-foreground-alt-1 text-sm">{user.username}</Badge>
				</span>
			{/if}
			<span class="text-foreground-alt-2">All associated data will be deleted</span>
		</div>

		{#if me}
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
			disabled={(me && currentPassword === '') || isPosting}
			onclick={deleteUser}
			class="bg-background-error disabled:bg-background-error/80 enabled:hover:bg-background-error-alt-1 text-foreground-alt-1 enabled:hover:text-foreground w-24"
		>
			{#if !isPosting}
				Delete
			{:else}
				<Spinner class="bg-foreground-alt-1 size-2" />
			{/if}
		</Button>
	{/snippet}
</AlertDialog>
