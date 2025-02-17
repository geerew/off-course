<script lang="ts">
	import { UpdateUser } from '$lib/api/users';
	import { Button, Dialog, Select } from '$lib/components/ui';
	import {
		SelectUserRoles,
		type UserModel,
		type UserRole,
		type UsersModel
	} from '$lib/models/user';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import Spinner from '../spinner.svelte';

	type Props = {
		open?: boolean;
		value: UserModel | UsersModel;
		trigger?: Snippet;
		triggerClass?: string;
		successFn?: () => void;
	};

	let {
		open = $bindable(false),
		value = $bindable(),
		trigger,
		triggerClass,
		successFn
	}: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let inputEl = $state<HTMLInputElement>();
	let isPosting = $state(false);
	let roleValue: UserRole | undefined = $state();

	const multipleUsers = Array.isArray(value);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function updateUsers(): Promise<void> {
		isPosting = true;

		try {
			if (multipleUsers) {
				await Promise.all(Object.values(value).map((u) => doUpdate(u)));
				toast.success('Selected users updated');
			} else {
				await doUpdate(value);
			}

			successFn?.();
		} catch (error) {
			toast.error((error as Error).message);
		}

		isPosting = false;
		open = false;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doUpdate(user: UserModel): Promise<void> {
		isPosting = true;

		if (!roleValue) {
			toast.error('Role is required');
			isPosting = false;
			return;
		}

		try {
			await UpdateUser(user.id, { role: roleValue });
			open = false;
			successFn?.();
		} catch (error) {
			toast.error((error as Error).message);
		}

		isPosting = false;
	}
</script>

<Dialog
	bind:open
	onOpenChange={() => {
		roleValue = undefined;
		isPosting = false;
	}}
	contentProps={{
		interactOutsideBehavior: 'close',
		onOpenAutoFocus: (e) => {
			e.preventDefault();
			inputEl?.focus();
		},
		onCloseAutoFocus: (e) => {
			e.preventDefault();
		}
	}}
	{trigger}
	{triggerClass}
>
	{#snippet content()}
		<div class="flex flex-col gap-2.5 p-5">
			<div>Update Role:</div>
			<Select
				placeholder="Select Role"
				type="single"
				items={SelectUserRoles}
				bind:value={roleValue}
				contentProps={{ sideOffset: 8, loop: true }}
				contentClass="z-50"
			/>
		</div>
	{/snippet}

	{#snippet action()}
		<Button disabled={isPosting || !roleValue} class="w-24" onclick={updateUsers}>
			{#if !isPosting}
				Update
			{:else}
				<Spinner class="bg-foreground-alt-3 size-2" />
			{/if}
		</Button>
	{/snippet}
</Dialog>
