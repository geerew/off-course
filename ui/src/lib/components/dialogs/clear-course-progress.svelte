<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { DeleteCourseProgress } from '$lib/api/course-api';
	import { Button, Dialog, Drawer } from '$lib/components/ui';
	import type { CourseModel } from '$lib/models/course-model';
	import { remCalc } from '$lib/utils';
	import type { Snippet } from 'svelte';
	import { toast } from 'svelte-sonner';
	import { innerWidth } from 'svelte/reactivity/window';
	import theme from 'tailwindcss/defaultTheme';
	import { Spinner } from '..';

	type Props = {
		open?: boolean;
		course: CourseModel;
		trigger?: Snippet;
		successFn?: () => void;
	};

	let { open = $bindable(false), course, trigger, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let isPosting = $state(false);

	const mdBreakpoint = +theme.screens.md.replace('rem', '');
	let isDesktop = $derived(remCalc(innerWidth.current ?? 0) > mdBreakpoint);

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
			await DeleteCourseProgress(course.id);

			successFn?.();
		} catch (error) {
			toast.error((error as APIError).message);
		}

		isPosting = false;
		open = false;
	}
</script>

{#snippet alertContents()}
	<Dialog.Alert>
		<div class="text-foreground-alt-1 flex flex-col gap-2 text-center">
			<span class="text-lg">Are you sure you want to delete your progress for this course?</span>
		</div>
	</Dialog.Alert>
{/snippet}

{#snippet deleteButton()}
	<Button variant="destructive" class=" w-24" disabled={isPosting} onclick={doDelete}>
		{#if isPosting}
			<Spinner class="bg-foreground-alt-1 size-2" />
		{:else}
			Delete
		{/if}
	</Button>
{/snippet}

{#if isDesktop}
	<Dialog.Root bind:open {trigger}>
		<Dialog.Content interactOutsideBehavior="close" class="w-lg">
			<div class="bg-background-alt-1 overflow-hidden rounded-lg">
				{@render alertContents()}

				<Dialog.Footer>
					<Dialog.CloseButton>Close</Dialog.CloseButton>
					{@render deleteButton()}
				</Dialog.Footer>
			</div>
		</Dialog.Content>
	</Dialog.Root>
{:else}
	<Drawer.Root bind:open>
		{@render trigger?.()}
		<Drawer.Content class="bg-background-alt-2" handleClass="bg-background-alt-4">
			<div class="bg-background-alt-1 overflow-hidden rounded-lg">
				{@render alertContents()}

				<Drawer.Footer>
					<Drawer.CloseButton>Close</Drawer.CloseButton>
					{@render deleteButton()}
				</Drawer.Footer>
			</div>
		</Drawer.Content>
	</Drawer.Root>
{/if}
