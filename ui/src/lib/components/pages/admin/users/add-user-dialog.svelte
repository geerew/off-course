<script lang="ts">
	import { CreateUser } from '$lib/api/user-api';
	import { Spinner } from '$lib/components';
	import { PlusIcon, UserIcon } from '$lib/components/icons';
	import { Button, Dialog, Input, InputPassword, Select } from '$lib/components/ui';
	import { SelectUserRoles, type UserRole } from '$lib/models/user-model';
	import { Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';

	type Props = {
		successFn?: () => void;
	};

	let { successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let open = $state(false);

	// Username
	let usernameInputEl = $state<HTMLInputElement>();
	let usernameValue = $state<string>('');

	// Display name
	let displayNameValue = $state<string>('');

	// Roles
	let roleValue: UserRole | '' = $state('');

	// Password
	let passwordValue = $state('');
	let confirmPasswordValue = $state('');

	// False when any of the password fields are empty
	let submitDisabled = $derived.by(() => {
		return (
			usernameValue === '' ||
			roleValue === '' ||
			passwordValue === '' ||
			confirmPasswordValue === ''
		);
	});

	// True when a request is being made
	let isPosting = $state(false);

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

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Focus on the username input when the page loads
	$effect(() => {
		if (usernameInputEl) {
			usernameInputEl.focus();
		}
	});
</script>

<Dialog
	bind:open
	triggerClass="flex h-10 w-auto flex-row items-center gap-2 px-5"
	contentClass="inline-flex h-[min(calc(100vh-10rem),46rem)] max-w-lg flex-col"
	onOpenChange={() => {
		usernameValue = '';
		displayNameValue = '';
		roleValue = '';
		passwordValue = '';
		confirmPasswordValue = '';
		isPosting = false;
	}}
	contentProps={{
		onOpenAutoFocus: (e) => {
			e.preventDefault();
			usernameInputEl?.focus();
		},
		onCloseAutoFocus: (e) => {
			e.preventDefault();
		}
	}}
>
	{#snippet trigger()}
		<PlusIcon class="size-5 stroke-[1.5]" />
		Add User
	{/snippet}

	{#snippet content()}
		<header
			class="border-background-alt-3 bg-background-alt-2 flex h-16 shrink-0 items-center border-b px-3 text-base font-medium"
		>
			<div class="flex items-center gap-2">
				<UserIcon class="size-5 stroke-2" />
				<span>User Add</span>
			</div>
		</header>

		<main
			class="flex min-h-[5rem] w-full flex-1 flex-col gap-4 overflow-x-hidden overflow-y-auto py-5"
		>
			<!-- Username -->
			<div class="flex place-content-center">
				<div class="flex w-xs flex-col gap-3">
					<div class="text-foreground-alt-2 text-[15px] uppercase">Username</div>
					<Input
						bind:ref={usernameInputEl}
						bind:value={usernameValue}
						name="username"
						type="text"
					/>
				</div>
			</div>

			<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

			<!-- Display name -->
			<div class="flex place-content-center">
				<div class="flex w-xs flex-col gap-3">
					<div class="text-foreground-alt-2 text-[15px] uppercase">Display Name</div>
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
					<div class="text-foreground-alt-2 text-[15px] uppercase">Role</div>
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
					<div class="text-foreground-alt-2 text-[15px] uppercase">Password</div>
					<InputPassword bind:value={passwordValue} name="new password" />
				</div>
			</div>

			<div class="flex place-content-center">
				<div class="flex w-xs flex-col gap-3">
					<div class="text-foreground-alt-2 text-[15px] uppercase">Confirm Password</div>
					<InputPassword bind:value={confirmPasswordValue} name="confirm password" />
				</div>
			</div>
		</main>
	{/snippet}

	{#snippet action()}
		<Button onclick={add} disabled={submitDisabled || isPosting} class="h-10 w-25 py-2">
			{#if !isPosting}
				Create
			{:else}
				<Spinner class="bg-foreground-alt-3 size-2" />
			{/if}
		</Button>
	{/snippet}
</Dialog>
