<script lang="ts">
	import {
		DeleteUserDialog,
		EditUserRoleDialog,
		RevokeUserSessionsDialog
	} from '$lib/components/dialogs';
	import {
		ActionIcon,
		DeleteIcon,
		DeselectAllIcon,
		FlagIcon,
		RightChevronIcon,
		SessionIcon
	} from '$lib/components/icons';
	import { Dropdown } from '$lib/components/ui';
	import type { UserModel } from '$lib/models/user-model';

	type Props = {
		users: Record<string, UserModel>;
		onUpdate: () => void;
		onDelete: () => void;
	};

	let { users = $bindable(), onUpdate, onDelete }: Props = $props();

	let roleDialogOpen = $state(false);
	let revokeSessionsDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);
</script>

<Dropdown.Root>
	<Dropdown.Trigger
		class="w-32 [&[data-state=open]>svg]:rotate-90"
		disabled={Object.keys(users).length === 0}
	>
		<div class="flex items-center gap-1.5">
			<ActionIcon class="size-4 stroke-[1.5]" />
			<span>Actions</span>
		</div>
		<RightChevronIcon class="stroke-foreground-alt-3 size-4.5 duration-200" />
	</Dropdown.Trigger>

	<Dropdown.Content class="w-42">
		<Dropdown.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				users = {};
			}}
		>
			<DeselectAllIcon class="size-4 stroke-[1.5]" />
			<span>Deselect all</span>
		</Dropdown.Item>

		<Dropdown.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				roleDialogOpen = true;
			}}
		>
			<FlagIcon class="size-4 stroke-[1.5]" />
			<span>Update Role</span>
		</Dropdown.Item>

		<Dropdown.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				revokeSessionsDialogOpen = true;
			}}
		>
			<SessionIcon class="size-4 stroke-[1.5]" />
			<span>Revoke Sessions</span>
		</Dropdown.Item>

		<Dropdown.Separator />

		<Dropdown.CautionItem
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete</span>
		</Dropdown.CautionItem>
	</Dropdown.Content>
</Dropdown.Root>

<EditUserRoleDialog bind:open={roleDialogOpen} value={Object.values(users)} successFn={onUpdate} />
<RevokeUserSessionsDialog bind:open={revokeSessionsDialogOpen} value={Object.values(users)} />
<DeleteUserDialog bind:open={deleteDialogOpen} value={Object.values(users)} successFn={onDelete} />
