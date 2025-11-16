<script lang="ts">
	import type { APIError } from '$lib/api-error.svelte';
	import { DeleteTag } from '$lib/api/tag-api';
	import { Spinner } from '$lib/components';
	import { Button, Dialog, Drawer } from '$lib/components/ui';
	import type { TagModel, TagsModel } from '$lib/models/tag-model';
	import { remCalc } from '$lib/utils';
	import { toast } from 'svelte-sonner';
	import { innerWidth } from 'svelte/reactivity/window';
	import theme from 'tailwindcss/defaultTheme';

	type Props = {
		open?: boolean;
		value: TagModel | TagsModel;
		successFn?: () => void;
	};

	let { open = $bindable(false), value, successFn }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let isPosting = $state(false);
	const isArray = Array.isArray(value);

	const mdBreakpoint = +theme.screens.md.replace('rem', '');
	let isDesktop = $derived(remCalc(innerWidth.current ?? 0) > mdBreakpoint);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	$effect(() => {
		if (open) {
			isPosting = false;
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function doDelete(): Promise<void> {
		isPosting = true;

		try {
			if (isArray) {
				await Promise.all(Object.values(value).map((t) => DeleteTag(t.id)));
				toast.success('Selected tags deleted');
			} else {
				await DeleteTag(value.id);
			}

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
			{#if isArray && Object.values(value).length > 1}
				<span class="text-lg">Are you sure you want to delete these tags?</span>
			{:else}
				<span class="text-lg">Are you sure you want to delete this tag?</span>
			{/if}
		</div>
	</Dialog.Alert>
{/snippet}

{#snippet deleteButton()}
	<Button variant="destructive" class="w-24" disabled={isPosting} onclick={doDelete}>
		{#if isPosting}
			<Spinner class="bg-foreground-alt-1 size-2" />
		{:else}
			Delete
		{/if}
	</Button>
{/snippet}

{#if isDesktop}
	<Dialog.Root bind:open>
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
