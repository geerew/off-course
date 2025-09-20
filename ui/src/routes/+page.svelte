<script lang="ts">
	import { GetCourses } from '$lib/api/course-api';
	import { LogoIcon, RightChevronIcon, WarningIcon } from '$lib/components/icons';
	import Spinner from '$lib/components/spinner.svelte';
	import { Badge, Button } from '$lib/components/ui';
	import type { CourseReqParams, CoursesModel } from '$lib/models/course-model';
	import { cn, remCalc } from '$lib/utils';
	import { Avatar } from 'bits-ui';
	import theme from 'tailwindcss/defaultTheme';

	type courseType = 'ongoing' | 'newest';

	let ongoingCourses: CoursesModel = $state([]);
	let newestCourses: CoursesModel = $state([]);

	let courseLinks: Record<courseType, string> = {
		ongoing: '/courses/?filter=started',
		newest: '/courses/?filter=newest'
	};

	let paginationPerPage = $state<number>();

	let loadPromise = $state(fetcher());

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
				? 6
				: windowWidth >= +theme.screens.md.replace('rem', '')
					? 6
					: 4;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch courses
	async function fetcher(): Promise<void> {
		if (!paginationPerPage) {
			setPaginationPerPage(remCalc(window.innerWidth));
		}

		try {
			const courseReqParams: CourseReqParams = {
				q: `sort:"courses_progress.updated_at desc" progress:"started"`,
				withUserProgress: true,
				page: 1,
				perPage: paginationPerPage
			};

			const ongoingData = await GetCourses(courseReqParams);
			ongoingCourses = ongoingData.items;

			const newestCourseReqParams: CourseReqParams = {
				q: `sort:"created_at desc" available:true`,
				withUserProgress: true,
				page: 1,
				perPage: paginationPerPage
			};

			const newestData = await GetCourses(newestCourseReqParams);
			newestCourses = newestData.items;
		} catch (error) {
			throw error;
		}
	}
</script>

{#snippet courses(type: courseType, courses: CoursesModel)}
	<div class="flex w-full flex-col gap-4">
		<div>
			<Button
				variant="ghost"
				class="hover:text-background-primary-alt-1 gap-3.5"
				href={courseLinks[type]}
				data-sveltekit-reload
			>
				<span class="text-lg font-semibold">
					{type === 'ongoing' ? 'Ongoing' : 'Newest'} Courses
				</span>

				<RightChevronIcon class="size-4.5 stroke-2" />
			</Button>
		</div>

		{#if courses.length === 0}
			<div class="flex w-full flex-col items-center gap-2 pt-5">
				<div class="flex flex-col items-center gap-2">
					<div>No courses</div>
				</div>
			</div>
		{:else}
			<div class="flex flex-col gap-5">
				<div class="grid grid-cols-1 items-stretch gap-5 md:grid-cols-2 lg:grid-cols-3">
					{#each courses as course}
						<Button
							href={`/course/${course.id}`}
							variant="ghost"
							class="group border-background-alt-3 flex h-full flex-col items-stretch gap-3 overflow-hidden rounded-lg border p-0 pb-2 text-start whitespace-normal"
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
													course.progress.percent === 100 && 'bg-background-success text-foreground'
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
											<Badge class="bg-background-warning text-foreground-alt-1">Initial Scan</Badge
											>
										{:else if course.maintenance}
											<Badge class="bg-background-warning text-foreground-alt-1">Maintenance</Badge>
										{:else if !course.available}
											<Badge class="bg-background-error text-foreground-alt-1">Unavailable</Badge>
										{/if}
									</div>
								</div>
							</div>
						</Button>
					{/each}
				</div>
			</div>
		{/if}
	</div>
{/snippet}

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-7xl flex-col gap-6 px-5 py-10">
		<div class="flex w-full place-content-center">
			<div class="flex w-full flex-col gap-8">
				{#await loadPromise}
					<div class="flex justify-center pt-10">
						<Spinner class="bg-foreground-alt-3 size-4" />
					</div>
				{:then _}
					{@render courses('ongoing', ongoingCourses)}
					{@render courses('newest', newestCourses)}
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
