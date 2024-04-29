<!-- TODO: Handle adding a tag that goes off screen (scroll to it?) -->
<script lang="ts">
	import Loading from '$components/generic/loading.svelte';
	import { Badge } from '$components/ui/badge';
	import { Button } from '$components/ui/button';
	import * as Dialog from '$components/ui/dialog';
	import { AddTag, GetTag } from '$lib/api';
	import { cn } from '$lib/utils';
	import axios from 'axios';
	import { Search, Tag, X } from 'lucide-svelte';
	import { createEventDispatcher } from 'svelte';
	import { toast } from 'svelte-sonner';

	// ----------------------
	// Variables
	// ----------------------

	const dispatch = createEventDispatcher();

	let isOpen = false;

	let inputEl: HTMLInputElement;

	let showSpinner = false;

	let toAdd: string[] = [];

	// ----------------------
	// Functions
	// ----------------------

	// Use:action for inputting tags
	const tagInput = (node: HTMLInputElement) => {
		async function handleInput(e: KeyboardEvent) {
			if (e.key === 'Enter') {
				e.preventDefault();

				if (!node.value) return;

				showSpinner = true;

				// Check if the tag already exists in the list
				if (toAdd.find((t) => t.toLowerCase() === node.value.toLowerCase())) {
					toast.error('Tag already in list of tags to add');
					showSpinner = false;
					return;
				}

				// Check if tag already exists in the backend
				try {
					await GetTag(node.value, true);

					toast.error('Tag already exists');
					showSpinner = false;
					return;
				} catch (error) {
					if (!axios.isAxiosError(error) || (error.response && error.response.status !== 404)) {
						toast.error(error instanceof Error ? error.message : String(error));
					}
				}

				toAdd = [...toAdd, node.value.toLowerCase()];
				node.value = '';

				showSpinner = false;
			}
		}

		node.addEventListener('keydown', handleInput);

		return {
			destroy() {
				node.removeEventListener('keydown', handleInput);
			}
		};
	};

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function addTags() {
		try {
			await Promise.all(
				toAdd.map(async (tag) => {
					try {
						await AddTag(tag);
					} catch (error) {
						toast.error('Failed to add tag: ' + tag);
					}
				})
			);
		} catch (error) {
			toast.error(error instanceof Error ? error.message : (error as string));
		}
	}
</script>

<Button
	variant="outline"
	class="bg-primary hover:bg-primary group flex h-8 gap-1.5 hover:brightness-110"
	on:click={() => (isOpen = true)}
>
	<Tag class="size-4" />
	<span>Add Tags</span>
</Button>

<Dialog.Root
	bind:open={isOpen}
	onOpenChange={(open) => {
		if (!open) {
			toAdd = [];
		} else {
			inputEl.focus();
		}
	}}
>
	<Dialog.Content
		class="bg-muted max-w-[27rem] p-0 transition md:max-w-xl [&>button[data-dialog-close]]:hidden"
	>
		<div class="border-alt-1/60 group relative flex flex-row items-center border-b">
			<label class="px-5" for="add-tag-input">
				<Search class="text-muted-foreground size-6" />
			</label>

			<input
				type="text"
				id="add-tag-input"
				bind:this={inputEl}
				use:tagInput
				placeholder="Add tag..."
				class="placeholder-muted-foreground/60 text-foreground h-14 w-full rounded-none border-none bg-inherit px-0 focus-visible:outline-none focus-visible:ring-0"
			/>

			<Loading
				class={cn('absolute right-3 h-auto min-h-0 w-auto p-0', !showSpinner && 'hidden')}
				loaderClass="size-5"
			/>
		</div>

		<div
			class="flex max-h-[20rem] min-h-[7rem] flex-col gap-2 overflow-hidden overflow-y-auto px-4"
		>
			<div class="flex flex-row flex-wrap gap-2.5">
				{#each toAdd as tag}
					<div class="flex flex-row">
						<!-- Tag -->
						<Badge
							class={cn(
								'bg-success text-success-foreground hover:bg-success min-w-0 items-center justify-between gap-1.5 whitespace-nowrap rounded-sm rounded-r-none border-none text-sm'
							)}
						>
							{tag}
						</Badge>

						<!-- Delete button -->
						<Button
							class={cn(
								'hover:bg-destructive bg-success text-success-foreground inline-flex h-auto items-center rounded-l-none rounded-r-sm border-l px-1.5 py-0.5'
							)}
							on:click={() => {
								// When its a newly added tag, just delete it completely
								toAdd = toAdd.filter((t) => t !== tag);
							}}
						>
							<X class="size-3" />
						</Button>
					</div>
				{/each}
			</div>
		</div>

		<Dialog.Footer
			class="border-alt-1/60 h-14 flex-row items-center justify-end gap-2 border-t px-4"
		>
			<Button
				variant="outline"
				class="bg-muted border-alt-1/60 hover:bg-alt-1/60 h-8 px-6"
				on:click={() => {
					isOpen = false;
				}}>Cancel</Button
			>
			<Button
				class="h-8 px-6"
				disabled={toAdd.length === 0}
				on:click={async () => {
					await addTags();
					dispatch('added');
					toAdd = [];
					isOpen = false;
				}}
			>
				Add {toAdd.length > 0 ? `(${toAdd.length})` : ''}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>