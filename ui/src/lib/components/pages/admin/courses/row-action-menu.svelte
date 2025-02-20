<script lang="ts">
	import { StartScan } from '$lib/api/scan-api';
	import { DeleteCourseDialog } from '$lib/components/dialogs';
	import { DeleteIcon, DotsIcon, ScanIcon } from '$lib/components/icons';
	import Dropdown from '$lib/components/ui/dropdown.svelte';
	import type { CourseModel } from '$lib/models/course-model';
	import { DropdownMenu } from 'bits-ui';
	import { toast } from 'svelte-sonner';

	type Props = {
		course: CourseModel;
		onScan: () => void;
		onDelete: () => void;
	};

	let { course, onScan, onDelete }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let deleteDialogOpen = $state(false);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doScan() {
		try {
			await StartScan({ courseId: course.id });
			onScan();
		} catch (error) {
			toast.error('Failed to start the scan ' + (error as Error).message);
		}
	}
</script>

<Dropdown
	triggerClass="hover:bg-background-alt-3 data-[state=open]:bg-background-alt-3 rounded-lg border-none"
	contentClass="w-42"
>
	{#snippet trigger()}
		<DotsIcon class="size-5 stroke-[1.5]" />
	{/snippet}

	{#snippet content()}
		<DropdownMenu.Item
			class="text-foreground-alt-1 hover:text-foreground hover:bg-background-alt-2 inline-flex w-full cursor-pointer items-center gap-2.5 rounded-md px-1 py-1 duration-200 select-none"
			onclick={async () => {
				doScan();
			}}
		>
			<ScanIcon class="size-4 stroke-[1.5]" />
			<span>Scan</span>
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

<!-- <EditUserRoleDialog bind:open={roleDialogOpen} value={user} successFn={onUpdate} />
<EditUserPasswordDialog bind:open={passwordDialogOpen} value={user} />
<RevokeUserSessionsDialog bind:open={revokeSessionsDialogOpen} value={user} /> -->
<DeleteCourseDialog bind:open={deleteDialogOpen} value={course} successFn={onDelete} />
