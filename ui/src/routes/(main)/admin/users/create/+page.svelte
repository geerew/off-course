<script lang="ts">
	import { goto } from '$app/navigation';
	import { CreateUser } from '$lib/api/users';
	import { Spinner } from '$lib/components';
	import { BackArrowIcon } from '$lib/components/icons';
	import { Input, InputPassword, Select, SubmitButton } from '$lib/components/ui';
	import type { UserRole } from '$lib/models/user';
	import { cn } from '$lib/utils';
	import { Button, Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';

	// Username
	let usernameInputEl = $state<HTMLInputElement>();
	let usernameValue = $state<string>('');

	// Display name
	let displayNameInputEl = $state<HTMLInputElement>();
	let displayNameValue = $state<string>('');

	// Roles
	let roles = [
		{ value: 'user', label: 'User' },
		{ value: 'admin', label: 'Admin' }
	];
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

	async function submitForm(event: Event) {
		event.preventDefault();

		// sleep for 10 second
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

			toast.success(`${roleValue === 'admin' ? 'Admin' : 'Basic'} user created`);
			goto('/admin/users');
		} catch (error) {
			if (error instanceof Error) {
				toast.error(error.message);
			} else {
				toast.error('An error occurred');
			}
		}

		isPosting = false;
	}

	// Focus on the username input when the page loads
	$effect(() => {
		if (usernameInputEl) {
			usernameInputEl.focus();
		}
	});
</script>

<div class="flex w-full place-content-center">
	<form onsubmit={submitForm} class="flex w-[22rem] flex-col gap-6 pt-1">
		<!-- Username -->
		<div class="flex place-content-center">
			<div class="flex w-xs flex-col gap-3">
				<div class="text-foreground-alt-2 text-[15px] uppercase">Username</div>
				<Input bind:ref={usernameInputEl} bind:value={usernameValue} name="username" type="text" />
			</div>
		</div>

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Display name -->
		<div class="flex place-content-center">
			<div class="flex w-xs flex-col gap-3">
				<div class="text-foreground-alt-2 text-[15px] uppercase">Display Name</div>
				<Input
					bind:ref={displayNameInputEl}
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
					items={roles}
					bind:value={roleValue}
					placeholder="Select a role"
					contentProps={{ loop: true }}
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

		<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

		<!-- Back / submit buttons  -->
		<div class="flex place-content-center">
			<div class="flex w-xs justify-end gap-6">
				<Button.Root
					href="/admin/users"
					class={cn(
						'bg-background-alt-4 hover:bg-background-alt-5 inline-flex h-10 flex-row items-center gap-2.5 rounded-md px-5 hover:cursor-pointer',
						isPosting &&
							'bg-background-alt-3 hover:bg-background-alt-3 text-foreground-alt-2 hover:cursor-not-allowed'
					)}
					aria-label="Toggle password visibility"
					onclick={(e) => {
						if (isPosting) e.preventDefault();
					}}
				>
					<BackArrowIcon class="size-5 stroke-[1.5]" />
					Back
				</Button.Root>

				<SubmitButton type="submit" disabled={submitDisabled || isPosting} class="h-10 py-2">
					{#if !isPosting}
						Create
					{:else}
						<Spinner class="bg-foreground-alt-3 size-2" />
					{/if}
				</SubmitButton>
			</div>
		</div>
	</form>
</div>
