<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { UpdateUser } from '$lib/api/user-api';
	import { Button, Dialog, Select } from '$lib/components/ui';
	import {
		SelectUserRoles,
		type UserModel,
		type UserRole,
		type UsersModel,
		type UserUpdateModel
	} from '$lib/models/user-model';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import Spinner from '../spinner.svelte';

	type Props = {
		open?: boolean;
		value: UserModel | UsersModel;
		trigger?: Snippet;
		successFn?: () => void;
	};

	let { open = $bindable(false), value, trigger, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let inputEl = $state<HTMLInputElement>();
	let isPosting = $state(false);
	let roleValue = $state<UserRole>();

	const isArray = Array.isArray(value);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			roleValue = undefined;
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doUpdate(): Promise<void> {
		isPosting = true;

		try {
			if (isArray) {
				await Promise.all(
					Object.values(value).map((u) =>
						UpdateUser(u.id, { role: roleValue } satisfies UserUpdateModel)
					)
				);
				toast.success('Users role updated');
			} else {
				await UpdateUser(value.id, { role: roleValue } satisfies UserUpdateModel);
				toast.success('Updated role');
			}

			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
		open = false;
	}
</script>

<Dialog.Root bind:open {trigger}>
	<Dialog.Content
		class="w-80"
		interactOutsideBehavior="close"
		onOpenAutoFocus={(e) => {
			e.preventDefault();
			inputEl?.focus();
		}}
		onCloseAutoFocus={(e) => {
			e.preventDefault();
		}}
	>
		<main class="flex flex-col gap-2.5 p-5">
			<div>Update Role:</div>
			<Select
				placeholder="Select Role"
				type="single"
				items={SelectUserRoles}
				bind:value={roleValue}
				contentProps={{ sideOffset: 8, loop: true }}
				contentClass="z-50"
			/>
		</main>

		<Dialog.Footer>
			<Dialog.CloseButton>Close</Dialog.CloseButton>
			<Button variant="default" class="w-24" disabled={isPosting || !roleValue} onclick={doUpdate}>
				{#if isPosting}
					<Spinner class="bg-background-alt-4  size-2" />
				{:else}
					Update
				{/if}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
