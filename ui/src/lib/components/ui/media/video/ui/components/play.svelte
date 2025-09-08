<script lang="ts">
	import { MediaPauseIcon, MediaPlayIcon } from '$lib/components/icons';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		isMobile?: boolean;
	};

	let { isMobile = false }: Props = $props();
</script>

<!-- TODO Fix issue where the video restarts following it ending and the user seeking and pressing play  -->
<div>
	{#if isMobile}
		<media-play-button
			class="group bg-background-primary/60 relative inline-flex cursor-pointer items-center justify-center rounded-full p-3 ring-sky-400 outline-none ring-inset hover:text-white data-[focus]:ring-4"
		>
			<MediaPlayIcon
				class="group-hover:fill-background-primary group-hover:stroke-background-primary hidden size-7 fill-white stroke-white group-data-[paused]:block"
			/>
			<MediaPauseIcon
				class="group-hover:fill-background-primary group-hover:stroke-background-primary size-7 fill-white stroke-white stroke-[2] group-data-[paused]:hidden"
			/>
		</media-play-button>
	{:else}
		<media-tooltip showDelay={300} class="contents">
			<media-tooltip-trigger>
				<media-play-button
					class="group relative inline-flex cursor-pointer items-center justify-center rounded-md ring-sky-400 outline-none ring-inset hover:text-white data-[focus]:ring-4"
				>
					<MediaPlayIcon
						class="group-hover:fill-background-primary group-hover:stroke-background-primary hidden size-7 fill-white stroke-white group-data-[paused]:block"
					/>
					<MediaPauseIcon
						class="group-hover:fill-background-primary group-hover:stroke-background-primary size-7 fill-white stroke-white stroke-[2] group-data-[paused]:hidden"
					/>
				</media-play-button>
			</media-tooltip-trigger>

			<media-tooltip-content
				class="animate-out fade-out slide-out-to-bottom-2 data-[visible]:animate-in data-[visible]:fade-in data-[visible]:slide-in-from-bottom-4 z-10 rounded-sm border border-gray-400/20 bg-white px-1 py-0.5 text-sm font-medium text-black"
				placement="top start"
				offset={10}
			>
				<span class="media-ended:!hidden media-paused:block hidden">Play</span>
				<span class="media-playing:block hidden">Pause</span>
			</media-tooltip-content>
		</media-tooltip>
	{/if}
</div>
