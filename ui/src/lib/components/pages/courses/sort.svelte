<script module lang="ts">
	let sortColumns = [
		{ label: 'Title', column: 'courses.title', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Available', column: 'courses.available', asc: 'Ascending', desc: 'Descending' },
		{ label: 'Added', column: 'courses.created_at', asc: 'Oldest', desc: 'Newest' },
		{ label: 'Updated', column: 'courses.updated_at', asc: 'Oldest', desc: 'Newest' },
		{
			label: 'Progress',
			column: 'courses_progress.updated_at',
			asc: 'Oldest',
			desc: 'Newest'
		}
	] as const satisfies SortColumns;

	export type SortColumn = (typeof sortColumns)[number]['column'];
</script>

<script lang="ts">
	import { SortMenu } from '$lib/components';
	import {
		RightChevronIcon,
		SortAscendingIcon,
		SortDescendingIcon,
		TickIcon
	} from '$lib/components/icons';
	import type { SortColumns, SortDirection } from '$lib/types/sort';
	import { cn } from '$lib/utils';
	import { Accordion, Label, RadioGroup, Separator, useId } from 'bits-ui';
	import { tick } from 'svelte';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		type: 'dropdown' | 'accordion';
		value?: string;
		selectedColumn?: SortColumn;
		selectedDirection?: SortDirection;
		disabled?: boolean;
		onApply: () => void;
	};

	let {
		type,
		value = $bindable(''),
		selectedColumn = $bindable(sortColumns[0].column),
		selectedDirection = $bindable('asc'),
		disabled = false,
		onApply
	}: Props = $props();
</script>

{#if type === 'dropdown'}
	<div class="flex h-10 items-center gap-3 rounded-lg">
		<SortMenu
			{disabled}
			columns={sortColumns}
			bind:selectedColumn
			bind:selectedDirection
			onUpdate={async () => {
				await tick();
				await onApply();
			}}
		/>
	</div>
{:else}
	<Accordion.Item value="sort" class="bg-background-alt-1 overflow-hidden rounded-lg">
		<Accordion.Header>
			<Accordion.Trigger
				class={cn(
					'data-[state=open]:border-b-background-primary-alt-1 data-[state=closed]:border-b-background-primary-alt-1 group flex w-full flex-1 select-none items-center justify-between border-b border-transparent px-2.5 py-5 font-medium transition-transform hover:cursor-pointer data-[state=closed]:border-b-2'
				)}
			>
				<div class="flex items-center gap-1.5">
					{#if selectedDirection === 'asc'}
						<SortAscendingIcon class="size-6 stroke-2" />
					{:else}
						<SortDescendingIcon class="size-6 stroke-2" />
					{/if}

					<span class="w-full text-left">
						Sort
						{#if sortColumns.find((column) => column.column === selectedColumn)}
							<span class="text-foreground-alt-3 text-sm">
								({sortColumns.find((column) => column.column === selectedColumn)?.label})
							</span>
						{/if}
					</span>
				</div>

				<RightChevronIcon
					class="size-4.5 stroke-2 transition-transform duration-100 group-data-[state=open]:rotate-90"
				/>
			</Accordion.Trigger>
		</Accordion.Header>

		<Accordion.Content
			class="data-[state=closed]:animate-accordion-up data-[state=open]:animate-accordion-down flex max-h-72 flex-col gap-2 overflow-hidden overflow-y-scroll px-2.5 py-3 text-sm tracking-[-0.01em]"
		>
			<RadioGroup.Root
				bind:value={selectedColumn}
				class="flex flex-col"
				onValueChange={(val) => {
					onApply();
				}}
			>
				{#each sortColumns as column}
					{@const id = useId()}
					<div
						class="hover:bg-background-alt-3 flex flex-row items-center overflow-hidden rounded-md hover:cursor-pointer"
					>
						<RadioGroup.Item
							{id}
							value={column.column}
							class="inline-flex size-3.5 h-full shrink-0 items-center justify-center py-1.5 hover:cursor-pointer"
						>
							{#snippet children({ checked })}
								<div class="inline-flex pl-2.5">
									{#if checked}
										<TickIcon class="size-3.5 stroke-2" />
									{:else}
										<span class="size-3.5"></span>
									{/if}
								</div>
							{/snippet}
						</RadioGroup.Item>

						<Label.Root
							for={id}
							class="inline-flex w-full select-none py-1 pl-3.5 pr-1.5 text-sm hover:cursor-pointer"
						>
							{column.label}
						</Label.Root>
					</div>
				{/each}
			</RadioGroup.Root>

			<Separator.Root class="bg-background-alt-5 h-px w-60 shrink-0" />

			<RadioGroup.Root
				bind:value={selectedDirection}
				class="flex flex-col"
				onValueChange={() => {
					onApply();
				}}
			>
				<!-- Each 2 for asc and desc -->
				{#each ['asc', 'desc'] as direction}
					{@const id = useId()}
					<div
						class="hover:bg-background-alt-3 flex flex-row items-center overflow-hidden rounded-md hover:cursor-pointer"
					>
						<RadioGroup.Item
							{id}
							value={direction}
							class="inline-flex size-3.5 h-full shrink-0 items-center justify-center py-1.5 hover:cursor-pointer"
						>
							{#snippet children({ checked })}
								<div class="inline-flex pl-2.5">
									{#if checked}
										<TickIcon class="size-3.5 stroke-2" />
									{:else}
										<span class="size-3.5"></span>
									{/if}
								</div>
							{/snippet}
						</RadioGroup.Item>

						<Label.Root
							for={id}
							class="inline-flex w-full select-none gap-1.5 py-1 pl-3.5 pr-1.5 text-sm hover:cursor-pointer"
						>
							{#if direction === 'asc'}
								<SortAscendingIcon class="text-foreground-alt-3 size-4" />
								{sortColumns.find((column) => column.column === selectedColumn)?.asc || 'Ascending'}
							{:else}
								<SortDescendingIcon class="text-foreground-alt-3 size-4" />
								{sortColumns.find((column) => column.column === selectedColumn)?.desc ||
									'Descending'}
							{/if}
						</Label.Root>
					</div>
				{/each}
			</RadioGroup.Root>
		</Accordion.Content>
	</Accordion.Item>
{/if}
