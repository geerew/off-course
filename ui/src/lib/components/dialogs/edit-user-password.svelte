<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { UpdateSelf } from '$lib/api/self-api';
	import { UpdateUser } from '$lib/api/user-api';
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { Button, Dialog, InputPassword } from '$lib/components/ui';
	import type { UserModel } from '$lib/models/user-model';
	import { Separator } from 'bits-ui';
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

	let firstInputEl = $state<HTMLInputElement>();
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isPosting = $state(false);

	let deletingSelf = value.id === auth?.user?.id;

	let passwordSubmitDisabled = $derived.by(() => {
		return (deletingSelf && currentPassword === '') || newPassword === '' || confirmPassword === '';
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doUpdate() {
		isPosting = true;

		try {
			if (deletingSelf) {
				if (newPassword !== confirmPassword) {
					toast.error('Passwords do not match');
					isPosting = false;
					return;
				}

				await UpdateSelf({ currentPassword, password: newPassword });
			} else {
				await UpdateUser(value.id, { password: newPassword });
			}

			successFn?.();
			toast.success('Password changed');
		} catch (error) {
			toast.error((error as APIError).message);
		}

		open = false;
		isPosting = false;
	}
</script>

<Dialog.Root
	bind:open
	onOpenChange={() => {
		currentPassword = '';
		newPassword = '';
		confirmPassword = '';
		isPosting = false;
	}}
	{trigger}
>
	<Dialog.Content
		class="max-w-sm"
		interactOutsideBehavior="close"
		onOpenAutoFocus={(e) => {
			e.preventDefault();
			firstInputEl?.focus();
		}}
		onCloseAutoFocus={(e) => {
			e.preventDefault();
		}}
	>
		<form onsubmit={doUpdate}>
			<main class="flex flex-col gap-4 p-5">
				{#if deletingSelf}
					<div class="flex flex-col gap-2.5">
						<div>Current Password:</div>
						<InputPassword
							bind:ref={firstInputEl}
							bind:value={currentPassword}
							name="current password"
						/>
					</div>

					<Separator.Root class="bg-background-alt-3 mt-2 h-px w-full shrink-0" />

					<div class="flex flex-col gap-2.5">
						<div>New Password:</div>
						<InputPassword bind:value={newPassword} name="new password" />
					</div>
				{:else}
					<div class="flex flex-col gap-2.5">
						<div>New Password:</div>
						<InputPassword bind:ref={firstInputEl} bind:value={newPassword} name="new password" />
					</div>
				{/if}

				<div class="flex flex-col gap-2.5">
					<div>Confirm Password:</div>
					<InputPassword bind:value={confirmPassword} name="confirm password" />
				</div>
			</main>

			<Dialog.Footer>
				<Dialog.CloseButton />
				<Button type="submit" disabled={passwordSubmitDisabled || isPosting} class="w-24">
					{#if !isPosting}
						Update
					{:else}
						<Spinner class="bg-foreground-alt-3 size-2" />
					{/if}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
