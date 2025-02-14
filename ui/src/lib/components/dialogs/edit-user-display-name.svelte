<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { Button, Dialog, Input } from '$lib/components/ui';
	import type { UserModel } from '$lib/models/user';
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

	let {
		open = $bindable(false),
		user = $bindable(),
		me,
		trigger,
		triggerClass,
		successFn
	}: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let inputEl = $state<HTMLInputElement>();
	let newValue = $state<string>('');
	let isPosting = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function update() {
		isPosting = true;

		let api = `/api/users/${user.id}`;
		if (me) {
			api = '/api/auth/me';
		}

		const response = await fetch(api, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				displayName: newValue
			})
		});

		if (response.ok) {
			if (me) {
				await auth.me();
			}
			user.displayName = newValue;
			open = false;

			successFn?.();
		} else {
			const data = await response.json();
			toast.error(data.message);
			isPosting = false;
		}
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
				placeholder={user.displayName}
			/>
		</div>
	{/snippet}

	{#snippet action()}
		<Button disabled={newValue === '' || isPosting} class="w-24" onclick={update}>
			{#if !isPosting}
				Update
			{:else}
				<Spinner class="bg-foreground-alt-3 size-2" />
			{/if}
		</Button>
	{/snippet}
</Dialog>
