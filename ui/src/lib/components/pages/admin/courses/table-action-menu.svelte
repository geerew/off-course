<script lang="ts">
	import { DeleteCourseDialog, EditCourseTagsDialog } from '$lib/components/dialogs';
	import { ActionIcon, DeleteIcon, DeselectIcon, TagIcon } from '$lib/components/icons';
	import RightChevron from '$lib/components/icons/right-chevron.svelte';
	import { Dropdown } from '$lib/components/ui';
	import type { CourseModel } from '$lib/models/course-model';
	import { DropdownMenu } from 'bits-ui';

	type Props = {
		courses: Record<string, CourseModel>;
		onDelete: () => void;
	};

	let { courses = $bindable(), onDelete }: Props = $props();

	let tagsDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);
</script>

<Dropdown
	triggerProps={{ disabled: Object.keys(courses).length === 0 }}
	triggerClass="w-32 [&[data-state=open]>svg]:rotate-90"
	contentClass="w-42"
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
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 text-sm duration-200 select-none"
			onclick={() => {
				courses = {};
			}}
		>
			<DeselectIcon class="size-4 stroke-[1.5]" />
			<span>Deselect all</span>
		</DropdownMenu.Item>

		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				tagsDialogOpen = true;
			}}
		>
			<TagIcon class="size-4 stroke-[1.5]" />
			<span>Add Tags</span>
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

<EditCourseTagsDialog bind:open={tagsDialogOpen} value={Object.values(courses)} />

<DeleteCourseDialog
	bind:open={deleteDialogOpen}
	value={Object.values(courses)}
	successFn={onDelete}
/>
