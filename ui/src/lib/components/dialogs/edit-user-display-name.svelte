<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { UpdateSelf } from '$lib/api/self-api';
	import { UpdateUser } from '$lib/api/user-api';
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { Button, Dialog, Input } from '$lib/components/ui';
	import type { SelfUpdateModel, UserModel, UserUpdateModel } from '$lib/models/user-model';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open?: boolean;
		value: UserModel;
		trigger?: Snippet;
		successFn?: () => void;
	};

	let { open = $bindable(false), value, trigger, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let inputEl = $state<HTMLInputElement>();
	let newValue = $state<string>('');
	let isPosting = $state(false);

	const deletingSelf = value.id === auth?.user?.id;

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			newValue = '';
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doUpdate(e: Event) {
		e.preventDefault();
		isPosting = true;

		try {
			if (deletingSelf) {
				await UpdateSelf({ displayName: newValue } satisfies SelfUpdateModel);
				await auth.me();
			} else {
				await UpdateUser(value.id, { displayName: newValue } satisfies UserUpdateModel);
			}

			open = false;
			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
	}
</script>

<Dialog.Root bind:open {trigger}>
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
		<form
			onsubmit={(e) => {
				doUpdate(e);
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
				<Dialog.CloseButton>Close</Dialog.CloseButton>

				<Button
					type="submit"
					variant="default"
					class="w-24"
					disabled={newValue === '' || isPosting}
				>
					{#if isPosting}
						<Spinner class="bg-background-alt-4  size-2" />
					{:else}
						Update
					{/if}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
