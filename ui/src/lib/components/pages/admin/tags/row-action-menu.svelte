<script lang="ts">
	import { DeleteTagDialog, EditTagNameDialog } from '$lib/components/dialogs';
	import { DeleteIcon, DotsIcon, EditIcon } from '$lib/components/icons';
	import Dropdown from '$lib/components/ui/dropdown.svelte';
	import type { TagModel } from '$lib/models/tag-model';
	import { DropdownMenu } from 'bits-ui';

	type Props = {
		tag: TagModel;
		onDelete: () => void;
	};

	let { tag = $bindable(), onDelete }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let tagDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);
</script>

<Dropdown
	triggerClass="hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 rounded-lg border-none"
	contentClass="w-38 p-1 text-sm"
	portalProps={{ disabled: false }}
>
	{#snippet trigger()}
		<DotsIcon class="size-5 stroke-[1.5]" />
	{/snippet}

	{#snippet content()}
		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				tagDialogOpen = true;
			}}
		>
			<EditIcon class="size-4 stroke-[1.5]" />
			<span>Rename</span>
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-background-alt-3 h-px w-full" />

		<DropdownMenu.Item
			class="text-foreground-error hover:text-foreground hover:bg-background-error inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete tag</span>
		</DropdownMenu.Item>
	{/snippet}
</Dropdown>

<EditTagNameDialog bind:open={tagDialogOpen} bind:value={tag} />
<DeleteTagDialog bind:open={deleteDialogOpen} value={tag} successFn={onDelete} />
