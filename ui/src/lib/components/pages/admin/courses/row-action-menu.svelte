<script lang="ts">
	import { goto } from '$app/navigation';
	import type { APIError } from '$lib/api-error.svelte';
	import { StartScan } from '$lib/api/scan-api';
	import { DeleteCourseDialog, EditCourseTagsDialog } from '$lib/components/dialogs';
	import { DeleteIcon, DotsIcon, OverviewIcon, ScanIcon, TagIcon } from '$lib/components/icons';
	import { Dropdown } from '$lib/components/ui';
	import type { CourseModel } from '$lib/models/course-model';
	import type { ScanCreateModel } from '$lib/models/scan-model';
	import { toast } from 'svelte-sonner';

	type Props = {
		course: CourseModel;
		onScan?: () => void;
		onDelete: () => void;
	};

	let { course, onScan, onDelete }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let tagsDialogOpen = $state(false);
	let deleteDialogOpen = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doScan() {
		try {
			await StartScan({ courseId: course.id } satisfies ScanCreateModel);
			onScan?.();
		} catch (error) {
			toast.error('Failed to start the scan ' + (error as APIError).message);
		}
	}
</script>

<Dropdown.Root>
	<Dropdown.Trigger
		class="hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 w-auto rounded-lg border-none"
	>
		<DotsIcon class="size-5 stroke-[1.5]" />
	</Dropdown.Trigger>

	<Dropdown.Content class="w-38">
		<Dropdown.Item
			onclick={async () => {
				goto(`/course/${course.id}`);
			}}
		>
			<OverviewIcon class="size-4 stroke-[1.5]" />
			<span>Overview</span>
		</Dropdown.Item>

		<Dropdown.Item
			onclick={async () => {
				doScan();
			}}
		>
			<ScanIcon class="size-4 stroke-[1.5]" />
			<span>Scan</span>
		</Dropdown.Item>

		<Dropdown.Item
			onclick={() => {
				tagsDialogOpen = true;
			}}
		>
			<TagIcon class="size-4 stroke-[1.5]" />
			<span>Edit Tags</span>
		</Dropdown.Item>

		<Dropdown.Separator />

		<Dropdown.CautionItem
			onclick={() => {
				deleteDialogOpen = true;
			}}
		>
			<DeleteIcon class="size-4 stroke-[1.5]" />
			<span>Delete Course</span>
		</Dropdown.CautionItem>
	</Dropdown.Content>
</Dropdown.Root>

<EditCourseTagsDialog bind:open={tagsDialogOpen} value={course} />
<DeleteCourseDialog bind:open={deleteDialogOpen} value={course} successFn={onDelete} />
