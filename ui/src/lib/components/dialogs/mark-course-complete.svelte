<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
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

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let isPosting = $state(false);

	const mdBreakpoint = +theme.screens.md.replace('rem', '');
	let isDesktop = $derived(remCalc(innerWidth.current ?? 0) > mdBreakpoint);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (open) {
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doMarkComplete(): Promise<void> {
		isPosting = true;

		try {
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
			<span class="text-lg">Mark course as complete?</span>
			<span class="text-foreground-alt-3">This will mark every asset as complete</span>
		</div>
	</Dialog.Alert>
{/snippet}

{#snippet confirmButton()}
	<Button variant="default" class="w-32" disabled={isPosting} onclick={doMarkComplete}>
		{#if isPosting}
			<Spinner class="bg-foreground-alt-1 size-2" />
		{:else}
			Mark Complete
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
					{@render confirmButton()}
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
					{@render confirmButton()}
				</Drawer.Footer>
			</div>
		</Drawer.Content>
	</Drawer.Root>
{/if}
