<script lang="ts">
	import { Button, Dialog, Select } from '$lib/components/ui';
	import { SelectRoles, type UserModel, type UserRole } from '$lib/models/user';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import Spinner from '../spinner.svelte';

	type Props = {
		open?: boolean;
		user: UserModel;
		trigger?: Snippet;
		triggerClass?: string;
		successFn?: () => void;
	};

	let {
		open = $bindable(false),
		user = $bindable(),
		trigger,
		triggerClass,
		successFn
	}: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let inputEl = $state<HTMLInputElement>();
	let isPosting = $state(false);

	let roleValue: UserRole = $state(user.role);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function update() {
		isPosting = true;

		const response = await fetch(`/api/users/${user.id}`, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				role: roleValue
			})
		});

		if (response.ok) {
			user.role = roleValue;
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
				type="single"
				items={SelectRoles}
				bind:value={roleValue}
				contentProps={{ sideOffset: 8, loop: true }}
				contentClass="z-50"
			/>
		</div>
	{/snippet}

	{#snippet action()}
		<Button disabled={isPosting} class="w-24" onclick={update}>
			{#if !isPosting}
				Update
			{:else}
				<Spinner class="bg-foreground-alt-3 size-2" />
			{/if}
		</Button>
	{/snippet}
</Dialog>
