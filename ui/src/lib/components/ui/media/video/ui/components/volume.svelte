<script lang="ts">
	import {
		MediaVolumeHighIcon,
		MediaVolumeLowIcon,
		MediaVolumeMuteIcon
	} from '$lib/components/icons';
	import { cn } from '$lib/utils';

	let show = false;
</script>

<div
	class="relative inline-flex"
	role="button"
	tabindex="0"
	aria-haspopup="true"
	aria-expanded={show}
	on:mouseenter={() => {
		show = true;
	}}
	on:mouseleave={() => {
		show = false;
	}}
>
	<!-- Volume/mute -->
	<media-mute-button
		class="group hover:text-secondary relative inline-flex cursor-pointer items-center justify-center rounded-md ring-sky-400 outline-none ring-inset data-[focus]:ring-4"
	>
		<MediaVolumeMuteIcon
			class="group-hover:fill-background-primary group-hover:stroke-background-primary hidden size-7 fill-white stroke-white stroke-2 group-data-[state=muted]:block"
		/>
		<MediaVolumeLowIcon
			class="group-hover:stroke-background-primary group-hover:fill-background-primary hidden size-7 fill-white stroke-white stroke-2 group-data-[state=low]:block"
		/>
		<MediaVolumeHighIcon
			class="group-hover:[&>path]:first:fill-background-primary group-hover:stroke-background-primary hidden size-7 stroke-white stroke-2 group-data-[state=high]:block [&>path]:first:fill-white"
		/>
	</media-mute-button>

	<media-volume-slider
		class={cn(
			'group relative inline-flex w-0 cursor-pointer touch-none items-center transition-all duration-200 outline-none select-none',
			show && 'ml-3.5 w-20'
		)}
		orientation="horizontal"
	>
		<!-- Track -->
		<div
			class="relative z-0 h-[5px] w-full rounded-sm bg-white/30 ring-sky-400 group-data-[focus]:ring-[3px]"
		>
			<!-- Fill -->
			<div
				class="bg-background-primary absolute h-full w-[var(--slider-fill)] rounded-sm will-change-[width]"
			></div>
		</div>

		<!-- Thumb -->
		<div
			class="absolute top-1/2 left-[var(--slider-fill)] z-20 h-[15px] w-[15px] -translate-x-1/2 -translate-y-1/2 rounded-full border border-[#cacaca] bg-white opacity-0 ring-white/40 transition-opacity will-change-[left] group-data-[active]:opacity-100 group-data-[dragging]:ring-4"
		></div>

		<media-slider-preview
			class="pointer-events-none flex flex-col items-center opacity-0 transition-opacity duration-200 data-[visible]:opacity-100"
			noClamp={false}
		>
			<media-slider-value class="rounded-sm bg-white px-2 py-px text-[13px] font-medium text-black"
			></media-slider-value>
		</media-slider-preview>
	</media-volume-slider>
</div>
