<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { Button, Dialog, InputPassword } from '$lib/components/ui';
	import type { UserModel } from '$lib/models/user';
	import { Separator } from 'bits-ui';
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

	let { open = $bindable(false), user, me, trigger, triggerClass, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let firstInputEl = $state<HTMLInputElement>();
	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isPosting = $state(false);

	let passwordSubmitDisabled = $derived.by(() => {
		return (me && currentPassword === '') || newPassword === '' || confirmPassword === '';
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function update() {
		isPosting = true;

		if (newPassword !== confirmPassword) {
			toast.error('Passwords do not match');
			isPosting = false;
			return;
		}

		let api = `/api/users/${user.id}`;
		let body = JSON.stringify({ password: newPassword });

		if (me) {
			api = '/api/auth/me';
			body = JSON.stringify({
				currentPassword: currentPassword,
				password: newPassword
			});
		}

		const response = await fetch(api, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body
		});

		if (response.ok) {
			if (me) {
				await auth.me();
			} else {
				toast.success('Password changed');
			}

			successFn?.();
			open = false;
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
		currentPassword = '';
		newPassword = '';
		confirmPassword = '';
		isPosting = false;
	}}
	contentProps={{
		interactOutsideBehavior: 'close',
		onOpenAutoFocus: (e) => {
			e.preventDefault();
			firstInputEl?.focus();
		},
		onCloseAutoFocus: (e) => {
			e.preventDefault();
		}
	}}
	{trigger}
	{triggerClass}
>
	{#snippet content()}
		<div class="flex flex-col gap-4 p-5">
			{#if me}
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
		</div>
	{/snippet}

	{#snippet action()}
		<Button disabled={passwordSubmitDisabled || isPosting} class="w-24" onclick={update}>
			{#if !isPosting}
				Update
			{:else}
				<Spinner class="bg-foreground-alt-3 size-2" />
			{/if}
		</Button>
	{/snippet}
</Dialog>
