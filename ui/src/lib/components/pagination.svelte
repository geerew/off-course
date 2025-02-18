<script lang="ts">
	import { SelectPaginationPerPage } from '$lib/models/pagination-model';
	import { Pagination } from 'bits-ui';
	import { LeftChevronIcon, RightChevronIcon } from './icons';
	import { Select } from './ui';

	type Props = {
		count: number;
		page: number;
		onPageChange: () => void;
		perPage: number;
		onPerPageChange: () => void;
	};

	let {
		count,
		page = $bindable(),
		onPageChange,
		perPage = $bindable(),
		onPerPageChange
	}: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let perPageValue = $state(`${perPage}`);
</script>

<Pagination.Root {count} {perPage} bind:page class="flex w-full justify-center" {onPageChange}>
	{#snippet children({ pages, range })}
		<div class="grid w-full flex-1 grow grid-cols-5 items-center justify-between gap-5">
			<Select
				type="single"
				items={SelectPaginationPerPage}
				bind:value={perPageValue}
				contentProps={{ sideOffset: 8, loop: true }}
				contentClass=""
				triggerClass="w-full"
				onValueChange={(v) => {
					perPage = +v;

					// If the current page is greater than the new total pages, set it to the last
					// page
					if (page > Math.ceil(count / perPage)) {
						page = Math.ceil(count / perPage);
					}

					onPerPageChange();
				}}
			/>

			<div class="col-span-3 flex place-content-center items-center gap-5">
				{#if count > perPage}
					<Pagination.PrevButton
						class="hover:bg-background-alt-2 disabled:text-background-alt-6 data-[selected]:text-background data-[selected]:bg-background-primary-alt-1 inline-flex h-10 flex-row items-center justify-center gap-1 rounded-lg pr-2 text-[15px] font-medium duration-200 select-none hover:cursor-pointer disabled:cursor-not-allowed hover:disabled:bg-transparent"
					>
						<LeftChevronIcon class="size-6" />
						<span class="text-xs">PREVIOUS</span>
					</Pagination.PrevButton>

					<div class="flex items-center gap-2.5">
						{#each pages as page (page.key)}
							{#if page.type === 'ellipsis'}
								<div class="text-foreground-alt text-[15px] font-medium select-none">...</div>
							{:else}
								<Pagination.Page
									{page}
									class="hover:bg-background-alt-2 data-[selected]:text-background data-[selected]:bg-background-primary-alt-1 inline-flex size-10 items-center justify-center rounded-lg text-[15px] font-medium duration-200 select-none hover:cursor-pointer"
								>
									{page.value}
								</Pagination.Page>
							{/if}
						{/each}
					</div>

					<Pagination.NextButton
						class="hover:bg-background-alt-2 disabled:text-background-alt-6 data-[selected]:text-background data-[selected]:bg-background-primary-alt-1 inline-flex h-10 flex-row items-center justify-center gap-1 rounded-lg pl-2 text-[15px] font-medium duration-200 select-none hover:cursor-pointer disabled:cursor-not-allowed hover:disabled:bg-transparent"
					>
						<span class="text-xs">NEXT</span>
						<RightChevronIcon class="size-5" />
					</Pagination.NextButton>
				{/if}
			</div>

			<p class="text-muted-foreground text-end text-sm">
				{range.start} - {range.end} / {count}
			</p>
		</div>
	{/snippet}
</Pagination.Root>
