<script lang="ts">
	import { UpdateSelf, UpdateUser } from '$lib/api/users';
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { Button, Dialog, Input } from '$lib/components/ui';
	import type { UserModel } from '$lib/models/user';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open?: boolean;
		value: UserModel;
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

<Dialog
	bind:open
	onOpenChange={() => {
		newValue = '';
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
			<div>Display Name:</div>
			<Input
				bind:ref={inputEl}
				bind:value={newValue}
				name="display name"
				type="text"
				placeholder={value.displayName}
			/>
		</div>
	{/snippet}

	{#snippet action()}
		<Button disabled={newValue === '' || isPosting} class="w-24" onclick={doUpdate}>
			{#if !isPosting}
				Update
			{:else}
				<Spinner class="bg-foreground-alt-3 size-2" />
			{/if}
		</Button>
	{/snippet}
</Dialog>
