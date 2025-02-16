<script lang="ts">
	import { DeleteUserDialog } from '$lib/components/dialogs';
	import { ActionIcon, DeleteIcon, DeselectIcon } from '$lib/components/icons';
	import RightChevron from '$lib/components/icons/right-chevron.svelte';
	import { Dropdown } from '$lib/components/ui';
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

<Dropdown
	triggerProps={{ disabled: Object.keys(users).length === 0 }}
	triggerClass="[&[data-state=open]>svg]:rotate-90 w-32"
>
	{#snippet trigger()}
		<div class="flex items-center gap-1.5">
			<ActionIcon class="size-4 stroke-[1.5]" />
			<span>Actions</span>
		</div>
		<RightChevron class="stroke-foreground-alt-2 size-4.5 duration-200" />
	{/snippet}

	{#snippet content()}
		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				users = {};
			}}
		>
			<DeselectIcon class="size-4 stroke-[1.5]" />
			<span>Deselect all</span>
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
	{/snippet}
</Dropdown>

<!-- <EditUserRoleDialog bind:open={roleDialogOpen} {user} successFn={onUpdate} /> -->
<DeleteUserDialog
	bind:open={deleteDialogOpen}
	value={Object.values(users)}
	me={false}
	successFn={onDelete}
/>
