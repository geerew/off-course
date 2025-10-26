<script lang="ts">
	import Fullscreen from './ui/components/fullscreen.svelte';
	import Mute from './ui/components/mute.svelte';
	import Play from './ui/components/play.svelte';
	import Settings from './ui/components/settings.svelte';
	import TimeSlider from './ui/components/time-slider.svelte';
	import Timestamp from './ui/components/timestamp.svelte';
</script>

<media-controls
	class="z-2 data-visible:opacity-100 data-visible:ease-in pointer-events-none absolute inset-0 box-border flex h-full w-full flex-col opacity-0 transition-opacity duration-200 ease-out group-data-[pointer=fine]/player:hidden"
>
	<media-controls-group
		class="pointer-events-auto relative flex w-full basis-3/12 flex-col items-end gap-2 px-4 pt-3"
	>
		<div class="flex items-center gap-4">
			<Mute />
			<Settings />
		</div>
	</media-controls-group>

	<media-controls-group
		class="pointer-events-auto relative flex w-full basis-4/12 place-content-center items-center gap-5 px-4"
	>
		<Play isMobile={true} />
	</media-controls-group>

	<media-controls-group
		class="pointer-events-auto relative flex w-full basis-5/12 flex-col items-center justify-end px-4 pb-3"
	>
		<div class="mb-2 flex w-full flex-row justify-between">
			<Timestamp />
			<Fullscreen isMobile={true} />
		</div>

		<TimeSlider />
	</media-controls-group>

	<div
		class="pointer-events-none absolute left-0 top-0 z-[-1] h-[99px] w-full bg-top bg-repeat-x [background-image:_url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAADGCAYAAAAT+OqFAAAAdklEQVQoz42QQQ7AIAgEF/T/D+kbq/RWAlnQyyazA4aoAB4FsBSA/bFjuF1EOL7VbrIrBuusmrt4ZZORfb6ehbWdnRHEIiITaEUKa5EJqUakRSaEYBJSCY2dEstQY7AuxahwXFrvZmWl2rh4JZ07z9dLtesfNj5q0FU3A5ObbwAAAABJRU5ErkJggg==)]"
	></div>

	<div
		class="pointer-events-none absolute bottom-0 left-0 z-[-1] h-[99px] w-full bg-bottom bg-repeat-x [background-image:_url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAADGCAYAAAAT+OqFAAAAdklEQVQoz42QQQ7AIAgEF/T/D+kbq/RWAlnQyyazA4aoAB4FsBSA/bFjuF1EOL7VbrIrBuusmrt4ZZORfb6ehbWdnRHEIiITaEUKa5EJqUakRSaEYBJSCY2dEstQY7AuxahwXFrvZmWl2rh4JZ07z9dLtesfNj5q0FU3A5ObbwAAAABJRU5ErkJggg==)]"
	></div>
</media-controls>
