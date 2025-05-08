<media-time-slider
	class="group relative mx-[7.5px] inline-flex h-6 w-full cursor-pointer touch-none items-center outline-none select-none"
>
	<!-- Track -->
	<div
		class="relative z-0 h-1 w-full rounded-sm bg-white/30 ring-sky-400 transition-[height] group-hover:h-2.5 group-data-[focus]:ring-1"
	>
		<!-- Progress -->
		<div
			class="absolute z-10 h-full w-[var(--slider-progress)] rounded-sm bg-white/50 will-change-[width]"
		></div>

		<!-- Fill -->
		<div
			class="bg-background-primary-alt-1 absolute z-[11] h-full w-[var(--slider-fill)] rounded-sm will-change-[width]"
		></div>
	</div>

	<!-- Thumb -->
	<div
		class="absolute top-1/2 left-[var(--slider-fill)] z-20 size-4 -translate-x-1/2 -translate-y-1/2 rounded-full border border-[#cacaca] bg-white opacity-0 ring-white/40 transition-opacity will-change-[left] group-data-[active]:opacity-100 group-data-[dragging]:ring-4"
	></div>

	<!-- Time preview -->
	<media-slider-preview
		class="pointer-events-none flex flex-col items-center opacity-0 transition-opacity duration-200 data-[visible]:opacity-100"
		noClamp={false}
	>
		<media-slider-value class="rounded-sm bg-white px-2 py-px text-[13px] font-medium text-black"
		></media-slider-value>
	</media-slider-preview>
</media-time-slider>
