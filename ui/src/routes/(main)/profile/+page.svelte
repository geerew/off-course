<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { Spinner } from '$lib/components';
	import { EditIcon } from '$lib/components/icons';
	import { AlertDialog, Button, Dialog, Input, InputPassword } from '$lib/components/ui';
	import { Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';

	// Dialog controls
	let dialogDisplayName = $state(false);
	let dialogPassword = $state(false);
	let dialogDeleteAccount = $state(false);

	// Input elements for focus
	let displayNameInputEl = $state<HTMLInputElement>();
	let passwordCurrentEl = $state<HTMLInputElement>();

	let displayNameValue = $state<string>('');

	// Password fields
	let currentPasswordValue = $state('');
	let newPasswordValue = $state('');
	let confirmPasswordValue = $state('');

	// False when any of the password fields are empty
	let passwordSubmitDisabled = $derived.by(() => {
		return currentPasswordValue === '' || newPasswordValue === '' || confirmPasswordValue === '';
	});

	// True when a request is being made
	let isPosting = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Send a DELETE request to delete the account
	async function submitDeleteAccount(event: Event) {
		event.preventDefault();
		isPosting = true;

		const response = await fetch('/api/auth/me', {
			method: 'DELETE',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				current_password: currentPasswordValue
			})
		});

		if (response.ok) {
			auth.empty();
			window.location.href = '/auth/login';
		} else {
			const data = await response.json();
			toast.error(`${data.message}`);
			isPosting = false;
		}
	}

	// Send a PUT request to update the display name
	async function submitDisplayNameForm(event: Event) {
		event.preventDefault();
		isPosting = true;

		const response = await fetch('/api/auth/me', {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				display_name: displayNameValue
			})
		});

		if (response.ok) {
			await auth.me();
			dialogDisplayName = false;
		} else {
			const data = await response.json();
			toast.error(data.message);
			isPosting = false;
		}
	}

	// Send a PUT request to update the password
	async function submitPasswordForm(event: Event) {
		event.preventDefault();
		isPosting = true;

		if (newPasswordValue !== confirmPasswordValue) {
			toast.error('Passwords do not match');
			isPosting = false;
			return;
		}

		const response = await fetch('/api/auth/me', {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				current_password: currentPasswordValue,
				password: newPasswordValue
			})
		});

		if (response.ok) {
			await auth.me();
			dialogPassword = false;
		} else {
			const data = await response.json();
			toast.error(data.message);
			isPosting = false;
		}
	}
</script>

{#if auth.user !== null}
	<div class="container-px py-8">
		<div class="mx-auto flex max-w-2xl flex-col place-content-center items-start gap-5">
			<!-- Username -->
			<div class="flex flex-col gap-3">
				<div class="text-foreground-alt-2 text-[15px] uppercase">Username</div>
				<span class="text-background-primary text-2xl">{auth.user.username}</span>
			</div>

			<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

			<!-- Display name -->
			<div class="flex flex-col gap-3">
				<div class="flex flex-row items-center gap-3">
					<div class="text-foreground-alt-2 text-[15px] uppercase">Display Name</div>
					<Dialog
						bind:open={dialogDisplayName}
						onOpenChange={() => {
							displayNameValue = '';
							isPosting = false;
						}}
						contentProps={{
							interactOutsideBehavior: 'close',
							onOpenAutoFocus: (e) => {
								e.preventDefault();
								displayNameInputEl?.focus();
							},
							onCloseAutoFocus: (e) => {
								e.preventDefault();
							}
						}}
						triggerClass="text-foreground-alt-2 bg-transparent hover:bg-transparent w-4.5 hover:text-foreground-alt-1 py-0 mb-0.5 cursor-pointer duration-200"
					>
						{#snippet trigger()}
							<EditIcon class="size-4.5 stroke-2" />
						{/snippet}

						{#snippet content()}
							<div class="flex flex-col gap-2.5 p-5">
								<div>Display Name:</div>
								<Input
									bind:ref={displayNameInputEl}
									bind:value={displayNameValue}
									name="display name"
									type="text"
									placeholder={auth?.user?.displayName}
								/>
							</div>
						{/snippet}

						{#snippet action()}
							<Button
								disabled={displayNameValue === '' || isPosting}
								class="w-24"
								onclick={submitDisplayNameForm}
							>
								{#if !isPosting}
									Update
								{:else}
									<Spinner class="bg-foreground-alt-3 size-2" />
								{/if}
							</Button>
						{/snippet}
					</Dialog>
				</div>
				<span class="text-background-primary text-2xl">{auth.user.displayName}</span>
			</div>

			<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

			<!-- Role -->
			<div class="flex flex-col gap-3">
				<div class="text-foreground-alt-2 text-[15px] uppercase">Role</div>
				<span class="text-background-primary text-2xl">{auth.isAdmin ? 'Admin' : 'User'}</span>
			</div>

			<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

			<!-- Password -->
			<div class="flex flex-col gap-3">
				<div class="text-foreground-alt-2 text-[15px] uppercase">Password</div>
				<Dialog
					bind:open={dialogPassword}
					onOpenChange={() => {
						currentPasswordValue = '';
						newPasswordValue = '';
						confirmPasswordValue = '';
						isPosting = false;
					}}
					contentProps={{
						interactOutsideBehavior: 'close',
						onOpenAutoFocus: (e) => {
							e.preventDefault();
							passwordCurrentEl?.focus();
						},
						onCloseAutoFocus: (e) => {
							e.preventDefault();
						}
					}}
				>
					{#snippet trigger()}
						Change Password
					{/snippet}

					{#snippet content()}
						<div class="flex flex-col gap-4 p-5">
							<div class="flex flex-col gap-2.5">
								<div>Current Password:</div>
								<InputPassword
									bind:ref={passwordCurrentEl}
									bind:value={currentPasswordValue}
									name="current password"
								/>
							</div>

							<Separator.Root class="bg-background-alt-3 mt-2 h-px w-full shrink-0" />

							<div class="flex flex-col gap-2.5">
								<div>New Password:</div>
								<InputPassword bind:value={newPasswordValue} name="new password" />
							</div>

							<div class="flex flex-col gap-2.5">
								<div>Confirm Password:</div>
								<InputPassword bind:value={confirmPasswordValue} name="confirm password" />
							</div>
						</div>
					{/snippet}

					{#snippet action()}
						<Button
							disabled={passwordSubmitDisabled || isPosting}
							class="w-24"
							onclick={submitPasswordForm}
						>
							{#if !isPosting}
								Update
							{:else}
								<Spinner class="bg-foreground-alt-3 size-2" />
							{/if}
						</Button>
					{/snippet}
				</Dialog>
			</div>

			<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

			<!-- Delete account -->
			<div class="flex flex-col gap-3">
				<div class="text-foreground-alt-2 text-[15px] uppercase">Delete Account</div>
				<AlertDialog
					bind:open={dialogDeleteAccount}
					onOpenChange={() => {
						currentPasswordValue = '';
						isPosting = false;
					}}
					contentProps={{
						interactOutsideBehavior: 'close',
						onOpenAutoFocus: (e) => {
							e.preventDefault();
							passwordCurrentEl?.focus();
						},
						onCloseAutoFocus: (e) => {
							e.preventDefault();
						}
					}}
				>
					{#snippet trigger()}
						Delete Account
					{/snippet}

					{#snippet description()}
						<div class="text-foreground-alt-1 flex flex-col gap-2 text-center">
							<span class="text-lg">Are you sure you want to delete your account?</span>
							<span class="text-foreground-alt-2">All associated data will be deleted</span>
						</div>

						<Separator.Root class="bg-background-alt-3 mt-2 h-px w-full shrink-0" />

						<div class="flex flex-col gap-2.5 px-2.5">
							<div>Confirm Password:</div>
							<InputPassword
								bind:ref={passwordCurrentEl}
								bind:value={currentPasswordValue}
								name="current password"
							/>
						</div>
					{/snippet}

					{#snippet action()}
						<Button
							disabled={currentPasswordValue === '' || isPosting}
							onclick={submitDeleteAccount}
							class="bg-background-error disabled:bg-background-error/80 enabled:hover:bg-background-error-alt-1 text-foreground-alt-1 enabled:hover:text-foreground w-24"
						>
							{#if !isPosting}
								Delete
							{:else}
								<Spinner class="bg-foreground-alt-1 size-2" />
							{/if}
						</Button>
					{/snippet}
				</AlertDialog>
			</div>
		</div>
	</div>
{/if}
