<script lang="ts">
	import { UpdateSelf, UpdateUser } from '$lib/api/user-api';
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { Button, Dialog, Input } from '$lib/components/ui';
	import type { UserModel } from '$lib/models/user-model';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open?: boolean;
		value: UserModel;
		trigger?: Snippet;
		successFn?: () => void;
	};

	let { open = $bindable(false), value = $bindable(), trigger, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let inputEl = $state<HTMLInputElement>();
	let newValue = $state<string>('');
	let isPosting = $state(false);

	const deletingSelf = value.id === auth?.user?.id;

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doUpdate() {
		isPosting = true;

		try {
			if (deletingSelf) {
				await UpdateSelf({ displayName: newValue });
				await auth.me();
			} else {
				await UpdateUser(value.id, { displayName: newValue });
			}

			value.displayName = newValue;
			open = false;

			successFn?.();
		} catch (error) {
			toast.error((error as Error).message);
		}

		isPosting = false;
	}
</script>

<Dialog.Root
	bind:open
	onOpenChange={() => {
		newValue = '';
		isPosting = false;
	}}
	{trigger}
>
	<Dialog.Content
		class="max-w-xs"
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
			<div>Display Name:</div>
			<Input
				bind:ref={inputEl}
				bind:value={newValue}
				name="display name"
				type="text"
				placeholder={value.displayName}
			/>
		</main>

		<Dialog.Footer>
			<Dialog.CloseButton />

			<Button disabled={newValue === '' || isPosting} class="w-24" onclick={doUpdate}>
				{#if !isPosting}
					Update
				{:else}
					<Spinner class="bg-foreground-alt-3 size-2" />
				{/if}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
