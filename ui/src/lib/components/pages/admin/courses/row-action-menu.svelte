<script lang="ts">
	import { goto } from '$app/navigation';
	import type { APIError } from '$lib/api-error.svelte';
	import { StartScan } from '$lib/api/scan-api';
	import { DeleteCourseDialog, EditCourseTagsDialog } from '$lib/components/dialogs';
	import { DeleteIcon, DotsIcon, OverviewIcon, ScanIcon, TagIcon } from '$lib/components/icons';
	import Dropdown from '$lib/components/ui/dropdown.svelte';
	import type { CourseModel } from '$lib/models/course-model';
	import { scanMonitor } from '$lib/scans.svelte';
	import { DropdownMenu } from 'bits-ui';
	import { toast } from 'svelte-sonner';

	type Props = {
		course: CourseModel;
		onScan: () => void;
		onDelete: () => void;
	};

	let { course, onScan, onDelete }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let tagsDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doScan() {
		try {
			await StartScan({ courseId: course.id });
			scanMonitor.trackCourses(course);
			onScan();
		} catch (error) {
			toast.error('Failed to start the scan ' + (error as APIError).message);
		}
	}
</script>

<Dropdown
	triggerClass="hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 rounded-lg border-none"
	contentClass="w-42 p-1"
>
	{#snippet trigger()}
		<DotsIcon class="size-5 stroke-[1.5]" />
	{/snippet}

	{#snippet content()}
		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 data-disabled:text-foreground-alt-3 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none disabled:opacity-50 data-disabled:cursor-default data-disabled:hover:bg-transparent"
			disabled={!course.available}
			onclick={async () => {
				if (!course.available) return;
				goto(`/course/${course.id}`);
			}}
		>
			<OverviewIcon class="size-4 stroke-[1.5]" />
			<span>Overview</span>
		</DropdownMenu.Item>

		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 data-disabled:text-foreground-alt-3 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none disabled:opacity-50 data-disabled:cursor-default data-disabled:hover:bg-transparent"
			onclick={async () => {
				doScan();
			}}
		>
			<ScanIcon class="size-4 stroke-[1.5]" />
			<span>Scan</span>
		</DropdownMenu.Item>

		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				tagsDialogOpen = true;
			}}
		>
			<TagIcon class="size-4 stroke-[1.5]" />
			<span>Edit Tags</span>
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

<EditCourseTagsDialog bind:open={tagsDialogOpen} value={course} />
<DeleteCourseDialog bind:open={deleteDialogOpen} value={course} successFn={onDelete} />
