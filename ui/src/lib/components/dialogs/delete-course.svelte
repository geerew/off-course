<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { DeleteCourse } from '$lib/api/course-api';
	import { Spinner } from '$lib/components';
	import { AlertDialog, Button } from '$lib/components/ui';
	import type { CourseModel, CoursesModel } from '$lib/models/course-model';
	import { type Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';

	type Props = {
		open?: boolean;
		value: CourseModel | CoursesModel;
		trigger?: Snippet;
		triggerClass?: string;
		successFn?: () => void;
	};

	let { open = $bindable(false), value, trigger, triggerClass, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let isPosting = $state(false);
	const isArray = Array.isArray(value);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doDelete(): Promise<void> {
		isPosting = true;

		try {
			if (isArray) {
				await Promise.all(Object.values(value).map((u) => DeleteCourse(u.id)));
				toast.success('Selected courses deleted');
			} else {
				await DeleteCourse(value.id);
			}

			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
		open = false;
	}
</script>

<AlertDialog
	bind:open
	contentProps={{
		interactOutsideBehavior: 'close'
	}}
	{trigger}
	{triggerClass}
>
	{#snippet description()}
		<div class="text-foreground-alt-1 flex flex-col gap-2 text-center">
			{#if isArray && Object.values(value).length > 1}
				<span class="text-lg">Are you sure you want to delete these courses?</span>
			{:else}
				<span class="text-lg">Are you sure you want to delete this course?</span>
			{/if}
			<span class="text-foreground-alt-2">All associated data will be deleted</span>
		</div>
	{/snippet}

	{#snippet action()}
		<Button
			disabled={isPosting}
			onclick={doDelete}
			class="bg-background-error disabled:bg-background-error/80 enabled:hover:bg-background-error-alt-1 text-foreground-alt-1 enabled:hover:text-foreground w-24"
		>
			{#if isPosting}
				<Spinner class="bg-foreground-alt-1 size-2" />
			{:else}
				Delete
			{/if}
		</Button>
	{/snippet}
</AlertDialog>
