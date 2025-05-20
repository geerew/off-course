<script lang="ts">
	import {
		DeleteUserDialog,
		EditUserPasswordDialog,
		EditUserRoleDialog,
		RevokeUserSessionsDialog
	} from '$lib/components/dialogs';
	import { DeleteIcon, DotsIcon, FlagIcon, SecureIcon, SessionIcon } from '$lib/components/icons';
	import { Dropdown } from '$lib/components/ui';
	import type { UserModel } from '$lib/models/user-model';
	import { cn } from '$lib/utils';

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

<Dropdown.Root>
	<Dropdown.Trigger
		class={cn(
			'hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 rounded-lg border-none',
			triggerClass
		)}
	>
		<DotsIcon class="size-5 stroke-[1.5]" />
	</Dropdown.Trigger>

	<Dropdown.Content class="w-42">
		<Dropdown.Item
			onclick={() => {
				roleDialogOpen = true;
			}}
		>
			<FlagIcon class="size-4 stroke-[1.5]" />
			<span>Update Role</span>
		</Dropdown.Item>

		<Dropdown.Item
			onclick={() => {
				passwordDialogOpen = true;
			}}
		>
			<SecureIcon class="size-4 stroke-[1.5]" />
			<span>Update Password</span>
		</Dropdown.Item>

		<Dropdown.Item
			onclick={() => {
				revokeSessionsDialogOpen = true;
			}}
		>
			<SessionIcon class="size-4 stroke-[1.5]" />
			<span>Revoke Sessions</span>
		</Dropdown.Item>

		<Dropdown.Separator />

		<Dropdown.CautionItem
			class="text-foreground-error hover:text-foreground hover:bg-background-error"
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete User</span>
		</Dropdown.CautionItem>
	</Dropdown.Content>
</Dropdown.Root>

<EditUserRoleDialog bind:open={roleDialogOpen} value={user} successFn={onUpdate} />
<EditUserPasswordDialog bind:open={passwordDialogOpen} value={user} />
<RevokeUserSessionsDialog bind:open={revokeSessionsDialogOpen} value={user} />
<DeleteUserDialog bind:open={deleteDialogOpen} value={user} successFn={onDelete} />
