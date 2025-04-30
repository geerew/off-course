<script lang="ts">
	import { GetCourses } from '$lib/api/course-api';
	import { FilterBar } from '$lib/components';
	import { LogoIcon, WarningIcon } from '$lib/components/icons';
	import Spinner from '$lib/components/spinner.svelte';
	import { Badge, Button } from '$lib/components/ui';
	import type { CoursesModel } from '$lib/models/course-model';
	import { scanMonitor } from '$lib/scans.svelte';
	import { cn, remCalc } from '$lib/utils';
	import { Avatar } from 'bits-ui';
	import theme from 'tailwindcss/defaultTheme';

	let courses: CoursesModel = $state([]);

	let filterValue = $state('');
	let filterAppliedValue = $state('');
	let filterOptions = {
		available: ['true', 'false'],
		tag: [],
		progress: ['not started', 'started', 'completed']
	};

	let paginationPage = $state(1);
	let paginationPerPage = $state<number>();
	let paginationTotal = $state<number>();

	let loadingMore = $state(false);

	let loadPromise = $state(fetchCourses(false, true));

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Stop the scan monitor when the component is destroyed
	$effect(() => {
		return () => scanMonitor.stop();
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Determine the number of courses to load base on the screen size
	$effect(() => {
		setPaginationPerPage(remCalc(window.innerWidth));
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the pagination perPage size based on the screen size
	function setPaginationPerPage(windowWidth: number) {
		paginationPerPage =
			windowWidth >= +theme.screens.lg.replace('rem', '')
				? 15
				: windowWidth >= +theme.screens.md.replace('rem', '')
					? 10
					: 8;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch courses
	async function fetchCourses(doScan: boolean, append: boolean): Promise<void> {
		if (!paginationPerPage) {
			setPaginationPerPage(remCalc(window.innerWidth));
		}

		try {
			if (doScan) await scanMonitor.fetch();

			const sort = 'sort:"courses.title asc"';
			const q = filterValue ? `${filterValue} ${sort}` : sort;

			const data = await GetCourses({
				q,
				page: paginationPage,
				perPage: paginationPerPage
			});
			paginationTotal = data.totalItems;

			if (append) {
				courses.push(...data.items);
			} else {
				courses = data.items;
			}
		} catch (error) {
			throw error;
		}
	}
</script>

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-7xl flex-col gap-6 px-5 py-10">
		<div class="flex w-full place-content-center">
			<div class="flex w-full flex-col gap-8">
				<div class="flex w-full flex-row items-center justify-between gap-5">
					<div class="flex max-w-[40rem] flex-1">
						<FilterBar
							bind:value={filterValue}
							disabled={!filterAppliedValue && courses.length === 0}
							{filterOptions}
							onApply={async () => {
								if (filterValue !== filterAppliedValue) {
									filterAppliedValue = filterValue;
									paginationPage = 1;
									loadPromise = fetchCourses(false, false);
								}
							}}
						/>
					</div>

					{#if courses.length > 0}
						<div class="flex flex-row justify-end">
							<Badge class="text-sm">
								{paginationTotal} courses
							</Badge>
						</div>
					{/if}
				</div>

				{#await loadPromise}
					<div class="flex justify-center pt-10">
						<Spinner class="bg-foreground-alt-3 size-4" />
					</div>
				{:then _}
					<div>
						{#if courses.length === 0}
							<div class="flex w-full flex-col items-center gap-2 pt-5">
								<div class="flex flex-col items-center gap-2">
									<div>No courses</div>

									{#if filterAppliedValue}
										<div class="text-foreground-alt-3">Try adjusting your filters</div>
									{/if}
								</div>
							</div>
						{:else}
							<div class="flex flex-col gap-5">
								<div class="grid-col-1 grid gap-5 md:grid-cols-2 lg:grid-cols-3">
									{#each courses as course}
										<Button
											href={`/course/${course.id}`}
											class={cn(
												'bg-background-alt-1 text-foreground-alt-1 hover:bg-background-alt-1 group flex h-auto flex-row items-start gap-1.5 overflow-hidden rounded-lg text-start duration-200 md:flex-col',
												!course.available && 'cursor-default'
											)}
											onclick={(e) => {
												if (!course.available) e.preventDefault();
											}}
										>
											<div class="h-px min-h-40 w-full md:min-h-35">
												<Avatar.Root class="h-full w-full">
													<Avatar.Image
														src={`/api/courses/${course.id}/card`}
														class="h-full w-full object-cover"
													/>

													<Avatar.Fallback
														class="bg-background-alt-2 flex h-full w-full place-content-center"
													>
														<LogoIcon class="fill-background-alt-3 w-12" />
													</Avatar.Fallback>
												</Avatar.Root>
											</div>

											<div class="flex h-full w-full flex-col justify-between gap-2 p-2.5">
												<span
													class={cn(
														'font-semibold duration-150',
														course.available
															? 'group-hover:text-background-primary'
															: 'text-foreground-alt-3'
													)}
												>
													{course.title}
												</span>

												<!-- Progress -->
												<div class="flex justify-end">
													{#if course.progress.percent > 0}
														<Badge
															class={cn(
																'text-foreground-alt-2',
																course.progress.percent === 100 &&
																	'bg-background-success text-foreground'
															)}
														>
															{course.progress.percent === 100
																? 'Completed'
																: course.progress.percent + '%'}
														</Badge>
													{/if}

													{#if !course.available}
														<Badge class="bg-background-error">unavailable</Badge>
													{/if}
												</div>
											</div>
										</Button>
									{/each}
								</div>

								{#if paginationTotal && paginationTotal > courses.length}
									<div class="flex w-full justify-center pt-5">
										<Button
											disabled={loadingMore}
											class="px-4 py-2 text-base font-semibold"
											onclick={async () => {
												paginationPage += 1;
												loadingMore = true;
												await fetchCourses(false, true);
												loadingMore = false;
											}}
										>
											{#if loadingMore}
												<Spinner class="bg-background-alt-4 size-4" />
											{:else}
												Load more
											{/if}
										</Button>
									</div>
								{/if}
							</div>
						{/if}
					</div>
				{:catch error}
					<div class="flex w-full flex-col items-center gap-2 pt-10">
						<WarningIcon class="text-foreground-error size-10" />
						<span class="text-lg">Failed to fetch courses: {error.message}</span>
					</div>
				{/await}
			</div>
		</div>
	</div>
</div>
