<script lang="ts">
	import { page } from '$app/state';
	import { GetAllCourseAssets, GetCourseFromParams } from '$lib/api/course-api';

	import { WarningIcon } from '$lib/components/icons';
	import Spinner from '$lib/components/spinner.svelte';
	import type { AssetModel, ChapteredAssets } from '$lib/models/asset-model';
	import type { CourseModel } from '$lib/models/course-model';

	let course = $state<CourseModel>();
	let chapteredAssets = $state<ChapteredAssets>({});

	let loadPromise = $state(fetchCourse());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Fetch course
	async function fetchCourse(): Promise<void> {
		try {
			course = await GetCourseFromParams(page.url.searchParams);

			const assets = await GetAllCourseAssets(course.id, {
				q: `sort:"assets.chapter asc" sort:"assets.prefix asc"`
			});

			chapteredAssets = BuildChapterStructure(assets);
		} catch (error) {
			throw error;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Build the course chapter structure
	export function BuildChapterStructure(courseAssets: AssetModel[]): ChapteredAssets {
		const chapters: ChapteredAssets = {};

		for (const courseAsset of courseAssets) {
			const chapter = courseAsset.chapter || '(no chapter)';
			!chapters[chapter]
				? (chapters[chapter] = [courseAsset])
				: chapters[chapter]?.push(courseAsset);
		}

		return chapters;
	}
</script>

{#await loadPromise}
	<div class="flex justify-center pt-10">
		<Spinner class="bg-foreground-alt-3 size-4" />
	</div>
{:then _}
	{#if course}
		<div
			class="grid grid-cols-[var(--course-menu-width)_1fr] grid-rows-1 gap-6 pt-[calc(var(--height-header)+1))]"
		>
			<div class="relative row-span-full">
				<div class="absolute inset-0">
					<nav
						class="border-foreground-alt-4 sticky top-[calc(var(--height-header)+1px)] left-0 flex h-[calc(100dvh-(var(--height-header)+1px))] w-[--course-menu-width] flex-col gap-2 overflow-x-hidden overflow-y-auto border-r pb-8"
					>
						<div
							class="bg-background border-background-alt-4 sticky top-0 z-[1] flex flex-row gap-3 border-b py-5 pr-3"
						>
							<span class="container-pl font-semibold">{course.title}</span>
						</div>

						{#each Object.keys(chapteredAssets) as chapter}
							<div class="container-pl leading-5">
								<div class="border-background-alt-6 flex flex-col gap-1.5 border-b py-1.5 pr-2">
									<span class=" text-background-primary text-sm font-semibold tracking-wide">
										{chapter}
									</span>
									<div class="flex flex-row items-center gap-1">
										<span class="text-foreground-alt-3 text-xs">
											0 / {chapteredAssets[chapter].length}
										</span>
									</div>
								</div>

								<ul class="ml-auto flex flex-col gap-2 pt-4 pb-3">
									{#each chapteredAssets[chapter] as asset, index}
										<li class="text-foreground-alt-2">
											{index + 1}. {asset.title}
										</li>
									{/each}
								</ul>
							</div>
						{/each}
					</nav>
				</div>
			</div>
			<main class="container-pr flex w-full py-8">
				<div class="flex w-full place-content-center">
					<div class="flex w-full max-w-7xl min-w-4xl flex-col gap-6 pt-1">
						<div class="flex w-full flex-col gap-8">// video</div>
					</div>
				</div>
			</main>
		</div>
	{/if}
{:catch error}
	<div class="flex w-full flex-col items-center gap-2 pt-10">
		<WarningIcon class="text-foreground-error size-10" />
		<span class="text-lg">Failed to fetch course</span>
		<span class="text-foreground-alt-3 text-sm">{error.message}</span>
	</div>
{/await}
