<script lang="ts">
	import { GetCourses } from '$lib/api/course-api';
	import { LogoIcon, WarningIcon } from '$lib/components/icons';
	import Spinner from '$lib/components/spinner.svelte';
	import { Badge, Button } from '$lib/components/ui';
	import type { CourseReqParams, CoursesModel } from '$lib/models/course-model';
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

	let loadPromise = $state(fetcher(false));

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Stop the scan monitor when the component is destroyed
	$effect(() => {
		return () => scanMonitor.clearAll();
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
	async function fetcher(append: boolean): Promise<void> {
		if (!paginationPerPage) {
			setPaginationPerPage(remCalc(window.innerWidth));
		}

		try {
			const sort = 'sort:"courses.title asc"';
			const q = filterValue ? `${filterValue} ${sort}` : sort;
			const courseReqParams: CourseReqParams = {
				q,
				withUserProgress: true,
				page: paginationPage,
				perPage: paginationPerPage
			};

			const data = await GetCourses(courseReqParams);
			paginationTotal = data.totalItems;

			if (append) {
				courses.push(...data.items);
			} else {
				courses = data.items;
			}

			// TODO change to use maintenance.svelte.ts
			// const coursesToTrack = courses.filter(
			// 	(course) => course.scanStatus === 'processing' || course.scanStatus === 'waiting'
			// );

			// scanMonitor.trackCourses(coursesToTrack);
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
						<!-- <FilterBar
							bind:value={filterValue}
							disabled={!filterAppliedValue && courses.length === 0}
							{filterOptions}
							onApply={async () => {
								if (filterValue !== filterAppliedValue) {
									filterAppliedValue = filterValue;
									paginationPage = 1;
									loadPromise = fetcher(false);
								}
							}}
						/> -->
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
							<div class="grid grid-cols-1 items-stretch gap-5 md:grid-cols-2 lg:grid-cols-3">
								{#each courses as course}
									<Button
										href={`/course/${course.id}`}
										variant="ghost"
										class="group border-background-alt-3 flex h-full flex-col items-start rounded-lg border p-2 text-start whitespace-normal"
									>
										<div class="relative aspect-[16/9] max-h-40 w-full overflow-hidden rounded-lg">
											{#if course.hasCard}
												<Avatar.Root class="h-full w-full">
													<Avatar.Image
														src={`/api/courses/${course.id}/card`}
														class="h-full w-full object-cover"
														data-card={course.hasCard}
													/>
													<Avatar.Fallback
														class="bg-background-alt-2 flex h-full w-full items-center justify-center"
													>
														<LogoIcon class="fill-background-alt-3 size-15 md:size-20" />
													</Avatar.Fallback>
												</Avatar.Root>
											{:else}
												<div
													class="bg-background-alt-2 z-1 flex h-full w-full items-center justify-center rounded-lg"
												>
													<LogoIcon class="fill-background-alt-3 size-15 md:size-20" />
												</div>
											{/if}
										</div>

										<div class="flex w-full flex-1 flex-col justify-between gap-3 pt-2">
											<!-- Title -->
											<div class="mb-3">
												<span
													class="group-hover:text-background-primary transition-colors duration-150"
												>
													{course.title}
												</span>
											</div>

											<div class="flex items-start justify-between">
												<!-- Progress -->
												<div class="flex gap-2">
													{#if course.progress?.started}
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
												</div>

												<div class="flex gap-2 font-medium">
													{#if course.initialScan !== undefined && !course.initialScan}
														<Badge class="bg-background-warning text-foreground-alt-1"
															>Initial Scan</Badge
														>
													{:else if course.maintenance}
														<Badge class="bg-background-warning text-foreground-alt-1"
															>Maintenance</Badge
														>
													{:else if !course.available}
														<Badge class="bg-background-error text-foreground-alt-1"
															>Unavailable</Badge
														>
													{/if}
												</div>
											</div>
										</div>
									</Button>
								{/each}
							</div>

							{#if paginationTotal && paginationTotal > courses.length}
								<div class="flex w-full justify-center pt-5">
									<Button
										variant="default"
										class="w-full px-4 text-base"
										disabled={loadingMore}
										onclick={async () => {
											paginationPage += 1;
											loadingMore = true;
											await fetcher(true);
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
