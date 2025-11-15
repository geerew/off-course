<media-time-slider
	class="group relative mx-[7.5px] inline-flex h-6 w-full cursor-pointer touch-none select-none items-center outline-none"
>
	<!-- Track -->
	<div
		class="group-data-focus:ring-1 relative z-0 h-1 w-full rounded-sm bg-white/30 ring-sky-400 transition-[height] group-hover:h-2.5"
	>
		<!-- Progress -->
		<div
			class="w-(--slider-progress) absolute z-10 h-full rounded-sm bg-white/50 will-change-[width]"
		></div>

		<!-- Fill -->
		<div
			class="bg-background-primary-alt-1 z-11 w-(--slider-fill) absolute h-full rounded-sm will-change-[width]"
		></div>
	</div>

	<!-- Thumb -->
	<div
		class="left-(--slider-fill) group-data-active:opacity-100 group-data-dragging:ring-4 absolute top-1/2 z-20 size-4 -translate-x-1/2 -translate-y-1/2 rounded-full border border-[#cacaca] bg-white opacity-0 ring-white/40 transition-opacity will-change-[left]"
	></div>

	<!-- Time preview -->
	<media-slider-preview
		class="data-visible:opacity-100 pointer-events-none flex flex-col items-center opacity-0 transition-opacity duration-200"
		noClamp={false}
	>
		<media-slider-value class="rounded-sm bg-white px-2 py-px text-[13px] font-medium text-black"
		></media-slider-value>
	</media-slider-preview>
</media-time-slider>
