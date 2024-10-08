<script lang="ts">
	import { Icons } from '$components/icons';
	import { Badge } from '$components/ui/badge';
	import { Button } from '$components/ui/button';
	import Separator from '$components/ui/separator/separator.svelte';
	import type { LogLevel } from '$lib/types/models';
	import { cn } from '$lib/utils';
	import { createEventDispatcher } from 'svelte';
	import LevelFilter from './level-filter.svelte';
	import MessageFilter from './message-filter.svelte';
	import TypeFilter from './type-filter.svelte';

	// ----------------------
	// Variables
	// ----------------------
	export let filterMessages: string[];
	export let filterLevels: LogLevel[];
	export let filterTypes: string[];

	const dispatchEvent = createEventDispatcher();

	// ----------------------
	// Reactive
	// ----------------------
	$: isFiltering = filterMessages.length > 0 || filterLevels.length > 0 || filterTypes.length > 0;
</script>

<!-- Filters -->
<div class="flex w-full flex-col gap-5 md:flex-row">
	<MessageFilter
		on:change={(e) => {
			filterMessages = [...filterMessages, e.detail];
			dispatchEvent('filterMessages', filterMessages);
		}}
	/>

	<div class="flex gap-2.5 md:gap-5">
		<LevelFilter
			bind:filterLevels
			on:change={() => {
				dispatchEvent('filterLevels', filterLevels);
			}}
		/>

		<TypeFilter
			bind:filterTypes
			on:change={() => {
				dispatchEvent('filterTypes', filterTypes);
			}}
		/>
	</div>
</div>

{#if isFiltering}
	<div class="flex flex-col gap-4 border-b border-alt-1/60 pb-5">
		<div class="flex flex-row items-center gap-2.5 text-sm text-primary">
			<Icons.Filter class="size-4" />
			<span class="tracking-wide">ACTIVE FILTERS</span>
		</div>

		<div class="flex flex-row items-center gap-2">
			<!-- Messages -->
			{#each filterMessages as message}
				<div class="flex flex-row" data-message={message}>
					<Badge
						class={cn(
							'h-6 min-w-0 items-center justify-between gap-2 whitespace-nowrap rounded-sm rounded-r-none bg-alt-1/60 text-foreground hover:bg-alt-1/60'
						)}
					>
						<Icons.Text class="size-3" />
						<span>{message}</span>
					</Badge>

					<Button
						class={cn(
							'inline-flex h-6 items-center rounded-l-none rounded-r-sm border-l bg-alt-1/60 px-1.5 py-0.5 duration-200 hover:bg-destructive'
						)}
						on:click={() => {
							filterMessages = filterMessages.filter((t) => t !== message);
							filterMessages = [...filterMessages];
							dispatchEvent('filterMessages', filterMessages);
						}}
					>
						<Icons.X class="size-3" />
					</Button>
				</div>
			{/each}

			<!-- Levels -->
			{#if filterLevels.length > 0}
				{#if filterMessages.length > 0}
					<Separator orientation="vertical" class="h-6 bg-alt-1" />
				{/if}
				{#each filterLevels as level}
					<div class="flex flex-row" data-level={level}>
						<Badge
							class={cn(
								'h-6 min-w-0 items-center justify-between gap-2 whitespace-nowrap rounded-sm rounded-r-none bg-alt-1/60 text-foreground hover:bg-alt-1/60'
							)}
						>
							<Icons.Scroll class="size-3" />
							<span>{level}</span>
						</Badge>

						<Button
							class={cn(
								'inline-flex h-6 items-center rounded-l-none rounded-r-sm border-l bg-alt-1/60 px-1.5 py-0.5 duration-200 hover:bg-destructive'
							)}
							on:click={() => {
								filterLevels = filterLevels.filter((t) => t !== level);
								filterLevels = [...filterLevels];
								dispatchEvent('filterLevels', filterLevels);
							}}
						>
							<Icons.X class="size-3" />
						</Button>
					</div>
				{/each}
			{/if}

			<!-- Types -->
			{#if filterTypes.length > 0}
				{#if filterMessages.length > 0 || filterLevels.length > 0}
					<Separator orientation="vertical" class="h-6 bg-alt-1" />
				{/if}
				{#each filterTypes as t}
					<div class="flex flex-row" data-type={t}>
						<Badge
							class={cn(
								'h-6 min-w-0 items-center justify-between gap-2 whitespace-nowrap rounded-sm rounded-r-none bg-alt-1/60 text-foreground hover:bg-alt-1/60'
							)}
						>
							<Icons.Text class="size-3" />
							<span>{t}</span>
						</Badge>

						<Button
							class={cn(
								'inline-flex h-6 items-center rounded-l-none rounded-r-sm border-l bg-alt-1/60 px-1.5 py-0.5 duration-200 hover:bg-destructive'
							)}
							on:click={() => {
								filterTypes = filterTypes.filter((type) => type !== t);
								filterTypes = [...filterTypes];
								dispatchEvent('filterTypes', filterTypes);
							}}
						>
							<Icons.X class="size-3" />
						</Button>
					</div>
				{/each}
			{/if}

			<Button
				class={cn(
					'inline-flex h-6 items-center rounded-lg bg-primary px-2.5 py-0.5 duration-200 hover:bg-primary hover:brightness-110'
				)}
				on:click={() => {
					filterMessages = [];
					filterLevels = [];
					filterTypes = [];

					dispatchEvent('clear');
				}}
			>
				Clear all
			</Button>
		</div>
	</div>
{/if}
