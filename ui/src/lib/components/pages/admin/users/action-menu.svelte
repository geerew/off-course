<script lang="ts">
	import {
		DeleteUserDialog,
		EditUserPasswordDialog,
		EditUserRoleDialog
	} from '$lib/components/dialogs';
	import { DeleteIcon, DotsIcon, FlagIcon, SecureIcon } from '$lib/components/icons';
	import type { UserModel } from '$lib/models/user';
	import { DropdownMenu } from 'bits-ui';

	type Props = {
		user: UserModel;
		refreshFn: () => void;
	};

	let { user = $bindable(), refreshFn }: Props = $props();

	let roleDialogOpen = $state(false);
	let passwordDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger
		class="hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 flex h-8 w-8 place-items-center items-center justify-center rounded-lg hover:cursor-pointer"
	>
		<DotsIcon class="size-5 stroke-[1.5]" />
	</DropdownMenu.Trigger>
	<DropdownMenu.Content
		align="end"
		sideOffset={2}
		class="bg-background border-background-alt-5 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2 flex w-36 flex-col gap-1 rounded-md border p-1 outline-none select-none data-[side=bottom]:translate-y-1 data-[side=top]:-translate-y-1"
	>
		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				roleDialogOpen = true;
			}}
		>
			<FlagIcon class="size-4 stroke-[1.5]" />
			<span>Set Role</span>
		</DropdownMenu.Item>

		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				passwordDialogOpen = true;
			}}
		>
			<SecureIcon class="size-4 stroke-[1.5]" />
			<span>Set Password</span>
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-background-alt-3 h-px w-full" />

		<DropdownMenu.Item
			class="text-foreground-error hover:text-foreground hover:bg-background-error inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete</span>
		</DropdownMenu.Item>
	</DropdownMenu.Content>
</DropdownMenu.Root>

<EditUserRoleDialog bind:open={roleDialogOpen} bind:user successFn={refreshFn} />
<EditUserPasswordDialog bind:open={passwordDialogOpen} {user} me={false} />
<DeleteUserDialog bind:open={deleteDialogOpen} {user} me={false} successFn={refreshFn} />
