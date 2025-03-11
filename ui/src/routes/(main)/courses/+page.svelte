<script lang="ts">
	import { GetCourses } from '$lib/api/course-api';
	import { WarningIcon } from '$lib/components/icons';
	import Spinner from '$lib/components/spinner.svelte';
	import { Badge, Button } from '$lib/components/ui';
	import type { CoursesModel } from '$lib/models/course-model';
	import { scanMonitor } from '$lib/scans.svelte';
	import { remCalc } from '$lib/utils';
	import theme from 'tailwindcss/defaultTheme';

	let courses: CoursesModel = $state([]);

	let paginationPage = $state(1);
	let paginationPerPage = $state<number>();
	let paginationTotal = $state<number>();

	let loadingMore = $state(false);

	let loadPromise = $state(fetchCourses(false));

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Stop the scan monitor when the component is destroyed
	$effect(() => {
		return () => scanMonitor.stop();
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Determine the number of courses to load base on the screen size
	$effect(() => {
		const windowWidth = remCalc(window.innerWidth);

		paginationPerPage =
			windowWidth >= +theme.screens.xl.replace('rem', '')
				? 24
				: windowWidth >= +theme.screens.md.replace('rem', '')
					? 16
					: 8;
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch courses
	async function fetchCourses(doScan: boolean): Promise<void> {
		try {
			if (doScan) await scanMonitor.fetch();

			const data = await GetCourses({
				orderBy: 'title',
				page: paginationPage,
				perPage: paginationPerPage
			});
			paginationTotal = data.totalItems;

			courses.push(...data.items.filter((course) => course.available));
		} catch (error) {
			throw error;
		}
	}
</script>

<div class="flex w-full place-content-center">
	<div class="flex w-full max-w-7xl flex-col gap-6 px-5 py-10">
		<div class="flex w-full place-content-center">
			{#await loadPromise}
				<div class="flex justify-center pt-10">
					<Spinner class="bg-foreground-alt-2 size-4" />
				</div>
			{:then _}
				<div class="flex w-full flex-col gap-8">
					<div>
						{#if courses.length === 0}
							<div class="flex w-full flex-col items-center gap-2 pt-10">
								<span class="text-lg">No courses found</span>
							</div>
						{:else}
							<div class="flex flex-col gap-5">
								<div class="flex flex-row justify-end">
									<Badge>
										{paginationTotal} courses
									</Badge>
								</div>
								<div class="grid-col-1 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
									{#each courses as course}
										<div
											class="border-background-alt-3 hover:border-background-alt-5 flex flex-row gap-4 overflow-hidden rounded-lg border duration-200"
										>
											<div class="bg-background-alt-2 h-full min-h-35 w-45"></div>

											<div class="flex w-full flex-col gap-2 p-2">
												<div class="flex w-full justify-between">
													<span class="text-base font-semibold">{course.title}</span>
												</div>
											</div>
										</div>
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
												// sleep for 3 seconds
												await new Promise((r) => setTimeout(r, 3000));
												await fetchCourses(false);
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
