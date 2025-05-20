<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { StartScan } from '$lib/api/scan-api';
	import { DeleteCourseDialog, EditCourseTagsDialog } from '$lib/components/dialogs';
	import {
		ActionIcon,
		DeleteIcon,
		DeselectAllIcon,
		RightChevronIcon,
		ScanIcon,
		TagIcon
	} from '$lib/components/icons';
	import { Dropdown } from '$lib/components/ui';
	import type { CourseModel } from '$lib/models/course-model';
	import { scanMonitor } from '$lib/scans.svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		courses: Record<string, CourseModel>;
		onScan: () => void;
		onDelete: () => void;
	};

	let { courses = $bindable(), onScan, onDelete }: Props = $props();

	let tagsDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doScan() {
		try {
			const coursesToScan = Object.values(courses);
			await Promise.all(coursesToScan.map((c) => StartScan({ courseId: c.id })));
			scanMonitor.trackCourses(coursesToScan);
			toast.success('Scanning started for selected courses');
			onScan();
		} catch (error) {
			toast.error('Failed to start the scans ' + (error as APIError).message);
		}
	}
</script>

<Dropdown.Root>
	<Dropdown.Trigger
		class="w-32 [&[data-state=open]>svg]:rotate-90"
		disabled={Object.keys(courses).length === 0}
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
				courses = {};
			}}
		>
			<DeselectAllIcon class="size-4 stroke-[1.5]" />
			<span>Deselect all</span>
		</Dropdown.Item>

		<Dropdown.Item
			class="text-foreground-alt-1 hover:text-foreground data-disabled:text-foreground-alt-3 hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none disabled:opacity-50 data-disabled:cursor-default data-disabled:hover:bg-transparent"
			onclick={async () => {
				doScan();
			}}
		>
			<ScanIcon class="size-4 stroke-[1.5]" />
			<span>Scan</span>
		</Dropdown.Item>

		<Dropdown.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={() => {
				tagsDialogOpen = true;
			}}
		>
			<TagIcon class="size-4 stroke-[1.5]" />
			<span>Add Tags</span>
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

<EditCourseTagsDialog bind:open={tagsDialogOpen} value={Object.values(courses)} />

<DeleteCourseDialog
	bind:open={deleteDialogOpen}
	value={Object.values(courses)}
	successFn={onDelete}
/>
