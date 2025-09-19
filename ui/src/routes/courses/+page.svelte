<script lang="ts">
	import { GetCourses } from '$lib/api/course-api';
	import { LogoIcon, WarningIcon } from '$lib/components/icons';
	import Filter from '$lib/components/pages/courses/filter.svelte';
	import Spinner from '$lib/components/spinner.svelte';
	import { Badge, Button } from '$lib/components/ui';
	import type { CourseReqParams, CoursesModel } from '$lib/models/course-model';
	import { scanMonitor } from '$lib/scans.svelte';
	import { cn, remCalc } from '$lib/utils';
	import { Avatar } from 'bits-ui';
	import theme from 'tailwindcss/defaultTheme';

	let courses: CoursesModel = $state([]);

	let filterValue = $state(`sort:"courses.title asc"`);
	let filterApplied = $state(false);

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
			const courseReqParams: CourseReqParams = {
				q: filterValue,
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
		} catch (error) {
			throw error;
		}
	}
</script>

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-7xl flex-col gap-6 px-5 py-10">
		<div class="flex w-full place-content-center">
			<div class="flex w-full flex-col gap-8">
				<Filter
					bind:filter={filterValue}
					onApply={async () => {
						filterApplied = true;
						paginationPage = 1;
						await fetcher(false);
					}}
				/>

				<div class="flex w-full justify-start">
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

								{#if filterApplied}
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
										class="group border-background-alt-3 flex h-full flex-row items-stretch gap-3 rounded-lg border p-2 text-start whitespace-normal md:flex-col"
									>
										<!-- Media -->
										<div
											class="relative w-64 flex-shrink-0 overflow-hidden rounded-lg md:aspect-[16/9] md:h-auto md:max-h-40 md:w-full"
										>
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

										<!-- Body -->
										<div
											class="flex h-40 min-w-0 flex-1 flex-col justify-between pt-1 md:h-auto md:pt-2"
										>
											<!-- Title -->
											<span
												class="group-hover:text-background-primary transition-colors duration-150"
											>
												{course.title}
											</span>

											<!-- Footer -->
											<div class="flex items-start justify-between pt-5">
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
