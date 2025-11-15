<script lang="ts">
	import { GetCourses } from '$lib/api/course-api';
	import { LogoIcon, WarningIcon } from '$lib/components/icons';
	import Filter from '$lib/components/pages/courses/filter.svelte';
	import Spinner from '$lib/components/spinner.svelte';
	import { Badge, Button } from '$lib/components/ui';
	import type { CourseReqParams, CoursesModel } from '$lib/models/course-model';
	import { cn, remCalc } from '$lib/utils';
	import { Avatar } from 'bits-ui';
	import { tick } from 'svelte';
	import theme from 'tailwindcss/defaultTheme';

	let courses: CoursesModel = $state([]);

	let filterValue = $state('');
	let filterApplied = $state(false);

	let paginationPage = $state(1);
	let paginationPerPage = $state<number>();
	let paginationTotal = $state<number>();

	let loadingMore = $state(false);
	let hasMoreCourses = $state(true);
	let loadingError = $state<string | null>(null);

	let loadPromise = $state(fetcher(false));

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set up intersection observer for infinite scrolling
	function setupIntersectionObserver(node: HTMLElement) {
		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0].isIntersecting && hasMoreCourses && !loadingMore) {
					loadMoreCourses();
				}
			},
			{ threshold: 0.1 }
		);

		observer.observe(node);

		return {
			destroy() {
				observer.disconnect();
			}
		};
	}

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

	// Load more courses for infinite scrolling
	async function loadMoreCourses(): Promise<void> {
		if (loadingMore || !hasMoreCourses) return;

		loadingMore = true;
		paginationPage += 1;

		try {
			loadingError = null;
			await fetcher(true);
		} catch (error) {
			// Reset pagination page on error to prevent getting stuck
			paginationPage -= 1;
			loadingError = error instanceof Error ? error.message : 'Failed to load more courses';
			console.error('Failed to load more courses:', error);
		} finally {
			loadingMore = false;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch courses
	async function fetcher(append: boolean): Promise<void> {
		// Let the filters sort themselves out
		await tick();

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

			// Update hasMoreCourses based on whether we've loaded all courses
			hasMoreCourses = courses.length < paginationTotal;
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
						hasMoreCourses = true;
						loadingError = null;
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
										class="border-background-alt-3 group flex h-full flex-col items-stretch gap-3 overflow-hidden whitespace-normal rounded-lg border p-0 pb-2 text-start"
									>
										<!-- Card -->
										<div class="relative aspect-[16/9] max-h-40 w-full overflow-hidden">
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

										<!-- Contents -->
										<div class="flex min-w-0 flex-1 flex-col justify-between gap-4 px-2 pt-1.5">
											<!-- Title -->
											<span
												class="group-hover:text-background-primary line-clamp-2 min-w-0 break-words transition-colors duration-150 md:line-clamp-none"
											>
												{course.title}
											</span>

											<!-- Footer -->
											<div class="flex items-start justify-between">
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

							{#if hasMoreCourses}
								<!-- Infinite scroll trigger -->
								<div use:setupIntersectionObserver class="flex w-full justify-center pt-5">
									{#if loadingMore}
										<div class="flex items-center gap-2">
											<Spinner class="bg-background-alt-4 size-4" />
											<span class="text-foreground-alt-3">Loading more courses...</span>
										</div>
									{:else if loadingError}
										<div class="flex flex-col items-center gap-2">
											<div class="text-foreground-error text-sm">Failed to load more courses</div>
											<Button
												variant="outline"
												class="text-sm"
												onclick={() => {
													loadingError = null;
													loadMoreCourses();
												}}
											>
												Retry
											</Button>
										</div>
									{/if}
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
