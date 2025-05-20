<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { UpdateSelf } from '$lib/api/self-api';
	import { UpdateUser } from '$lib/api/user-api';
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { Button, Dialog, Drawer, PasswordInput } from '$lib/components/ui';
	import type { UserModel } from '$lib/models/user-model';
	import { remCalc } from '$lib/utils';
	import { Separator } from 'bits-ui';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { innerWidth } from 'svelte/reactivity/window';
	import theme from 'tailwindcss/defaultTheme';

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

	let updatingSelf = value.id === auth?.user?.id;

	let passwordSubmitDisabled = $derived.by(() => {
		return (updatingSelf && currentPassword === '') || newPassword === '' || confirmPassword === '';
	});

	const mdBreakpoint = +theme.screens.md.replace('rem', '');
	let isDesktop = $derived(remCalc(innerWidth.current ?? 0) > mdBreakpoint);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			currentPassword = '';
			newPassword = '';
			confirmPassword = '';
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doUpdate() {
		isPosting = true;

		try {
			if (updatingSelf) {
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

	$inspect(isDesktop);
	console.log('updateSelf', updatingSelf);
</script>

{#snippet contents()}
	<main class="flex flex-col gap-4 p-5">
		{#if updatingSelf}
			<div class="flex flex-col gap-2.5">
				<div>Current Password:</div>
				<PasswordInput
					bind:ref={firstInputEl}
					bind:value={currentPassword}
					name="current password"
				/>
			</div>

			<Separator.Root class="bg-background-alt-3 mt-2 h-px w-full shrink-0" />

			<div class="flex flex-col gap-2.5">
				<div>New Password:</div>
				<PasswordInput bind:value={newPassword} name="new password" />
			</div>
		{:else}
			<div class="flex flex-col gap-2.5">
				<div>New Password:</div>
				<PasswordInput bind:ref={firstInputEl} bind:value={newPassword} name="new password" />
			</div>
		{/if}

		<div class="flex flex-col gap-2.5">
			<div>Confirm Password:</div>
			<PasswordInput bind:value={confirmPassword} name="confirm password" />
		</div>
	</main>
{/snippet}

{#snippet action()}
	<Button type="submit" disabled={passwordSubmitDisabled || isPosting} class="w-24">
		{#if isPosting}
			<Spinner class="bg-background-alt-4  size-2" />
		{:else}
			Update
		{/if}
	</Button>
{/snippet}

{#if isDesktop}
	<Dialog.Root bind:open {trigger}>
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
				{@render contents()}

				<Dialog.Footer>
					<Dialog.CloseButton>Close</Dialog.CloseButton>
					{@render action()}
				</Dialog.Footer>
			</form>
		</Dialog.Content>
	</Dialog.Root>
{:else}
	<Drawer.Root bind:open>
		{@render trigger?.()}

		<Drawer.Content>
			<div class="flex max-w-sm place-self-center">
				{@render contents()}
			</div>

			<Drawer.Footer>
				<Drawer.CloseButton>Close</Drawer.CloseButton>
				{@render action()}
			</Drawer.Footer>
		</Drawer.Content>
	</Drawer.Root>
{/if}
