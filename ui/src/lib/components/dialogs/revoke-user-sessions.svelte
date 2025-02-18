<script lang="ts">
	import { RevokeUserSessions } from '$lib/api/user-api';
	import { Spinner } from '$lib/components';
	import { AlertDialog, Badge, Button } from '$lib/components/ui';
	import type { UserModel, UsersModel } from '$lib/models/user-model';
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

	let isPosting = $state(false);
	const multipleUsers = Array.isArray(value);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function revokeUsers(): Promise<void> {
		isPosting = true;

		try {
			if (multipleUsers) {
				await Promise.all(Object.values(value).map((u) => revokeSession(u)));
				toast.success('Selected users deleted');
			} else {
				await revokeSession(value);
			}

			successFn?.();
		} catch (error) {
			toast.error((error as Error).message);
		}

		isPosting = false;
		open = false;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function revokeSession(user: UserModel): Promise<void> {
		await RevokeUserSessions(user.id);
	}
</script>

<AlertDialog
	bind:open
	onOpenChange={() => {
		isPosting = false;
	}}
	contentProps={{
		interactOutsideBehavior: 'close'
	}}
	{trigger}
	{triggerClass}
>
	{#snippet description()}
		<div class="text-foreground-alt-1 flex flex-col gap-2 text-center">
			{#if multipleUsers}
				<span class="text-lg">
					Are you sure you want to continue revoking all sessions for the selected users?
				</span>
				<span>
					<Badge class="bg-background-error text-foreground-alt-1 text-sm">
						{Object.values(value).length} user{Object.values(value).length > 1 ? 's' : ''} selected
					</Badge>
				</span>
			{:else}
				<span class="text-lg"
					>Are you sure you want to continue revoking all sessions for this user?</span
				>
				<span>
					<Badge class="bg-background-error text-foreground-alt-1 text-sm">
						{value.username}
					</Badge>
				</span>
			{/if}
		</div>
	{/snippet}

	{#snippet action()}
		<Button
			disabled={isPosting}
			onclick={revokeUsers}
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
