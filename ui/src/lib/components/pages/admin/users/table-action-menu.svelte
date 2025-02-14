<script lang="ts">
	import { DeleteUserDialog } from '$lib/components/dialogs';
	import { ActionIcon, DeleteIcon, DeselectIcon } from '$lib/components/icons';
	import RightChevron from '$lib/components/icons/right-chevron.svelte';
	import type { UserModel } from '$lib/models/user';
	import { DropdownMenu } from 'bits-ui';

	// TODO: Support updating role
	// TODO: Support removing session

	type Props = {
		users: Record<string, UserModel>;
		onUpdate: () => void;
		onDelete: () => void;
	};

	let { users = $bindable(), onUpdate, onDelete }: Props = $props();

	let roleDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);
</script>

<DropdownMenu.Root>
	<DropdownMenu.Trigger
		disabled={Object.keys(users).length === 0}
		class="border-background-alt-4 data-[state=open]:border-foreground-alt-2 hover:border-foreground-alt-2 disabled:text-foreground-alt-2 disabled:hover:border-background-alt-4 inline-flex h-10 w-32 items-center justify-between rounded-md border px-2.5 text-sm duration-200 select-none hover:cursor-pointer disabled:cursor-not-allowed [&[data-state=open]>svg]:rotate-90"
	>
		<div class="flex items-center gap-1.5">
			<ActionIcon class="size-4 stroke-[1.5]" />
			<span>Actions</span>
		</div>
		<RightChevron class="stroke-foreground-alt-2 size-4.5 duration-200" />
	</DropdownMenu.Trigger>

	<DropdownMenu.Content
		align="end"
		sideOffset={2}
		class="bg-background border-background-alt-5 data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2 flex w-36 flex-col gap-1 rounded-md border p-1 outline-none select-none data-[side=bottom]:translate-y-1 data-[side=top]:-translate-y-1"
	>
		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				users = {};
			}}
		>
			<DeselectIcon class="size-4 stroke-[1.5]" />
			<span>Deselect all</span>
		</DropdownMenu.Item>

		<!-- <DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				roleDialogOpen = true;
			}}
		>
			<FlagIcon class="size-4 stroke-[1.5]" />
			<span>Set Role</span>
		</DropdownMenu.Item> -->

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

<!-- <EditUserRoleDialog bind:open={roleDialogOpen} {user} successFn={onUpdate} /> -->
<DeleteUserDialog
	bind:open={deleteDialogOpen}
	value={Object.values(users)}
	me={false}
	successFn={onDelete}
/>
