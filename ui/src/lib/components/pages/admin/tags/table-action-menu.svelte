<script lang="ts">
	import { DeleteTagDialog } from '$lib/components/dialogs';
	import { ActionIcon, DeleteIcon, DeselectIcon } from '$lib/components/icons';
	import RightChevron from '$lib/components/icons/right-chevron.svelte';
	import { Dropdown } from '$lib/components/ui';
	import type { TagModel } from '$lib/models/tag-model';
	import { DropdownMenu } from 'bits-ui';

	type Props = {
		tags: Record<string, TagModel>;
		onDelete: () => void;
	};

	let { tags = $bindable(), onDelete }: Props = $props();

	let deleteDialogOpen = $state(false);
</script>

<Dropdown
	triggerProps={{ disabled: Object.keys(tags).length === 0 }}
	triggerClass="w-32 [&[data-state=open]>svg]:rotate-90"
	contentClass="w-42 p-1"
>
	{#snippet trigger()}
		<div class="flex items-center gap-1.5">
			<ActionIcon class="size-4 stroke-[1.5]" />
			<span>Actions</span>
		</div>
		<RightChevron class="stroke-foreground-alt-3 size-4.5 duration-200" />
	{/snippet}

	{#snippet content()}
		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 text-sm duration-200 select-none"
			onclick={() => {
				tags = {};
			}}
		>
			<DeselectIcon class="size-4 stroke-[1.5]" />
			<span>Deselect all</span>
		</DropdownMenu.Item>

		<DropdownMenu.Separator class="bg-background-alt-3 h-px w-full" />

		<DropdownMenu.Item
			class="text-foreground-error hover:text-foreground hover:bg-background-error inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 text-sm duration-200 select-none"
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete</span>
		</DropdownMenu.Item>
	{/snippet}
</Dropdown>

<DeleteTagDialog bind:open={deleteDialogOpen} value={Object.values(tags)} successFn={onDelete} />
