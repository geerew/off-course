<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import {
		DeleteUserDialog,
		EditUserDisplayNameDialog,
		EditUserPasswordDialog
	} from '$lib/components/dialogs';
	import { EditIcon } from '$lib/components/icons';
	import { Separator } from 'bits-ui';
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
					<EditUserDisplayNameDialog
						user={auth.user}
						me={true}
						triggerClass="text-foreground-alt-2 bg-transparent hover:bg-transparent w-4.5 hover:text-foreground-alt-1 py-0 mb-0.5 cursor-pointer duration-200"
					>
						{#snippet trigger()}
							<EditIcon class="size-4.5 stroke-2" />
						{/snippet}
					</EditUserDisplayNameDialog>
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
				<EditUserPasswordDialog user={auth.user} me={true}>
					{#snippet trigger()}
						Change Password
					{/snippet}
				</EditUserPasswordDialog>
			</div>

			<Separator.Root class="bg-background-alt-3 my-2 h-px w-full shrink-0" />

			<!-- Delete account -->
			<div class="flex flex-col gap-3">
				<div class="text-foreground-alt-2 text-[15px] uppercase">Delete Account</div>
				<DeleteUserDialog user={auth.user} me={true}>
					{#snippet trigger()}
						Delete Account
					{/snippet}
				</DeleteUserDialog>
			</div>
		</div>
	</div>
{/if}
