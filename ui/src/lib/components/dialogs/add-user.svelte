<script lang="ts">
	import { CreateUser } from '$lib/api/user-api';
	import { Spinner } from '$lib/components';
	import { PlusIcon, UserIcon } from '$lib/components/icons';
	import { Button, Dialog, Drawer, Input, PasswordInput, Select } from '$lib/components/ui';
	import { SelectUserRoles, type UserRole } from '$lib/models/user-model';
	import { remCalc } from '$lib/utils';
	import { Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';
	import { innerWidth } from 'svelte/reactivity/window';
	import theme from 'tailwindcss/defaultTheme';

	type Props = {
		successFn?: () => void;
	};

	let { successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let open = $state(false);

	let usernameInputEl = $state<HTMLInputElement>();
	let usernameValue = $state<string>('');

	let displayNameValue = $state<string>('');

	let roleValue: UserRole | '' = $state('');

	let passwordValue = $state('');
	let confirmPasswordValue = $state('');

	const mdBreakpoint = +theme.screens.md.replace('rem', '');
	let isDesktop = $derived(remCalc(innerWidth.current ?? 0) > mdBreakpoint);

	// False when any of the password fields are empty
	let submitDisabled = $derived.by(() => {
		return (
			usernameValue === '' ||
			roleValue === '' ||
			passwordValue === '' ||
			confirmPasswordValue === ''
		);
	});

	let isPosting = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			usernameValue = '';
			displayNameValue = '';
			roleValue = '';
			passwordValue = '';
			confirmPasswordValue = '';
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Focus on the username input when the page loads
	$effect(() => {
		if (usernameInputEl) {
			usernameInputEl.focus();
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function add(event: Event) {
		event.preventDefault();
		isPosting = true;

		if (passwordValue !== confirmPasswordValue) {
			toast.error('Passwords do not match');
			isPosting = false;
			return;
		}

		try {
			await CreateUser({
				username: usernameValue,
				displayName: displayNameValue,
				password: passwordValue,
				role: roleValue === '' ? 'user' : roleValue
			});

			successFn?.();
			toast.success('User created');
		} catch (error) {
			if (error instanceof Error) {
				toast.error(error.message);
			} else {
				toast.error('An error occurred');
			}
		}

		isPosting = false;
		open = false;
	}
</script>

{#snippet trigger()}
	{#if isDesktop}
		<Dialog.Trigger class="flex h-10 w-auto flex-row items-center gap-2 px-5">
			<PlusIcon class="size-5 stroke-[1.5]" />
			Add User
		</Dialog.Trigger>
	{:else}
		<Drawer.Trigger class="flex h-10 w-auto flex-row items-center gap-2 px-5 py-2">
			<PlusIcon class="size-5 stroke-[1.5]" />
			Add User
		</Drawer.Trigger>
	{/if}
{/snippet}

{#snippet contents()}
	<main
		class="flex max-h-[50vh] min-h-[5rem] w-full flex-1 flex-col gap-4 overflow-x-hidden overflow-y-auto py-5"
		data-vaul-no-drag=""
	>
		<!-- Username -->
		<div class="flex place-content-center">
			<div class="flex w-xs flex-col gap-3">
				<div class="text-foreground-alt-3 text-[15px] uppercase">Username</div>
				<Input bind:ref={usernameInputEl} bind:value={usernameValue} name="username" type="text" />
			</div>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Display name -->
		<div class="flex place-content-center">
			<div class="flex w-xs flex-col gap-3">
				<div class="text-foreground-alt-3 text-[15px] uppercase">Display Name</div>
				<Input
					bind:value={displayNameValue}
					name="display name"
					type="text"
					placeholder={displayNameValue === '' ? usernameValue : displayNameValue}
				/>
			</div>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Role -->
		<div class="flex place-content-center">
			<div class="flex w-xs flex-col gap-3">
				<div class="text-foreground-alt-3 text-[15px] uppercase">Role</div>
				<Select
					type="single"
					items={SelectUserRoles}
					bind:value={roleValue}
					placeholder="Select a role"
					contentProps={{ sideOffset: 8, loop: true }}
					contentClass="z-[50]"
				/>
			</div>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Password -->
		<div class="flex place-content-center">
			<div class="flex w-xs flex-col gap-3">
				<div class="text-foreground-alt-3 text-[15px] uppercase">Password</div>
				<PasswordInput bind:value={passwordValue} name="new password" />
			</div>
		</div>

		<div class="flex place-content-center">
			<div class="flex w-xs flex-col gap-3">
				<div class="text-foreground-alt-3 text-[15px] uppercase">Confirm Password</div>
				<PasswordInput bind:value={confirmPasswordValue} name="confirm password" />
			</div>
		</div>
	</main>
{/snippet}

{#if isDesktop}
	<Dialog.Root bind:open {trigger}>
		<Dialog.Content
			class="inline-flex h-[min(calc(100vh-10rem),46rem)] max-w-lg flex-col"
			onOpenAutoFocus={(e) => {
				e.preventDefault();
				usernameInputEl?.focus();
			}}
			onCloseAutoFocus={(e) => {
				e.preventDefault();
			}}
		>
			<Dialog.Header>
				<div class="flex items-center gap-2">
					<UserIcon class="size-5 stroke-2" />
					<span>User Add</span>
				</div>
			</Dialog.Header>

			<form
				onsubmit={add}
				class="flex min-h-[5rem] w-full flex-1 flex-col gap-4 overflow-x-hidden overflow-y-auto pt-5"
			>
				{@render contents()}

				<Dialog.Footer>
					<Dialog.CloseButton>Close</Dialog.CloseButton>

					<Button type="submit" disabled={submitDisabled || isPosting} class="h-10 w-25 py-2">
						{#if isPosting}
							<Spinner class="bg-background-alt-4 size-2" />
						{:else}
							Create
						{/if}
					</Button>
				</Dialog.Footer>
			</form>
		</Dialog.Content>
	</Dialog.Root>
{:else}
	<Drawer.Root bind:open>
		{@render trigger()}

		<Drawer.Content>
			<form onsubmit={add}>
				{@render contents()}

				<Drawer.Footer>
					<Drawer.CloseButton>Close</Drawer.CloseButton>

					<Button type="submit" disabled={submitDisabled || isPosting} class="h-10 w-25 py-2">
						{#if isPosting}
							<Spinner class="bg-background-alt-4 size-2" />
						{:else}
							Create
						{/if}
					</Button>
				</Drawer.Footer>
			</form>
		</Drawer.Content>
	</Drawer.Root>
{/if}
