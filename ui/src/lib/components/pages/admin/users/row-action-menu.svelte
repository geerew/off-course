<script lang="ts">
	import {
		DeleteUserDialog,
		EditUserPasswordDialog,
		EditUserRoleDialog,
		RevokeUserSessionsDialog
	} from '$lib/components/dialogs';
	import { DeleteIcon, DotsIcon, FlagIcon, SecureIcon, SessionIcon } from '$lib/components/icons';
	import Dropdown from '$lib/components/ui/dropdown.svelte';
	import type { UserModel } from '$lib/models/user-model';
	import { cn } from '$lib/utils';
	import { DropdownMenu } from 'bits-ui';

	type Props = {
		user: UserModel;
		triggerClass?: string;
		onUpdate: () => void;
		onDelete: () => void;
	};

	let { user, triggerClass, onUpdate, onDelete }: Props = $props();

	let roleDialogOpen = $state(false);
	let passwordDialogOpen = $state(false);
	let revokeSessionsDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);
</script>

<Dropdown
	triggerClass={cn(
		'hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 rounded-lg border-none',
		triggerClass
	)}
	contentClass="w-42 p-1 text-sm"
	portalProps={{ disabled: false }}
>
	{#snippet trigger()}
		<DotsIcon class="size-5 stroke-[1.5]" />
	{/snippet}

	{#snippet content()}
		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				roleDialogOpen = true;
			}}
		>
			<FlagIcon class="size-4 stroke-[1.5]" />
			<span>Update Role</span>
		</DropdownMenu.Item>

		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				passwordDialogOpen = true;
			}}
		>
			<SecureIcon class="size-4 stroke-[1.5]" />
			<span>Update Password</span>
		</DropdownMenu.Item>

		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				revokeSessionsDialogOpen = true;
			}}
		>
			<SessionIcon class="size-4 stroke-[1.5]" />
			<span>Revoke Sessions</span>
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-background-alt-3 h-px w-full" />

		<DropdownMenu.Item
			class="text-foreground-error hover:text-foreground hover:bg-background-error inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete User</span>
		</DropdownMenu.Item>
	{/snippet}
</Dropdown>

<EditUserRoleDialog bind:open={roleDialogOpen} value={user} successFn={onUpdate} />
<EditUserPasswordDialog bind:open={passwordDialogOpen} value={user} />
<RevokeUserSessionsDialog bind:open={revokeSessionsDialogOpen} value={user} />
<DeleteUserDialog bind:open={deleteDialogOpen} value={user} successFn={onDelete} />
