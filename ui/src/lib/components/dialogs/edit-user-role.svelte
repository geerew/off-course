<script lang="ts">
	import { UpdateUser } from '$lib/api/user-api';
	import { Button, Dialog, Select } from '$lib/components/ui';
	import {
		SelectUserRoles,
		type UserModel,
		type UserRole,
		type UsersModel
	} from '$lib/models/user-model';
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
	let roleValue = $state<UserRole>();

	const isArray = Array.isArray(value);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doUpdate(): Promise<void> {
		isPosting = true;

		try {
			if (isArray) {
				await Promise.all(Object.values(value).map((u) => UpdateUser(u.id, { role: roleValue })));
				toast.success('Selected users updated');
			} else {
				await UpdateUser(value.id, { role: roleValue });
			}

			successFn?.();
		} catch (error) {
			toast.error((error as Error).message);
		}

		isPosting = false;
		open = false;
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
	contentClass="w-80"
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
		<Button disabled={isPosting || !roleValue} class="w-24" onclick={doUpdate}>
			{#if !isPosting}
				Update
			{:else}
				<Spinner class="bg-foreground-alt-3 size-2" />
			{/if}
		</Button>
	{/snippet}
</Dialog>
