<script lang="ts">
	import Fullscreen from './ui/components/fullscreen.svelte';
	import Play from './ui/components/play.svelte';
	import Settings from './ui/components/settings.svelte';
	import TimeSlider from './ui/components/time-slider.svelte';
	import Timestamp from './ui/components/timestamp.svelte';
	import Volume from './ui/components/volume.svelte';
</script>

<media-controls
	class="pointer-events-none absolute inset-0 z-2 box-border flex h-full w-full flex-col opacity-0 transition-opacity duration-200 ease-out group-data-[pointer=coarse]/player:hidden data-visible:opacity-100 data-visible:ease-in"
>
	<div class="flex-1"></div>

	<media-controls-group class="pointer-events-auto flex w-full items-center px-3">
		<TimeSlider />
	</media-controls-group>

	<media-controls-group
		class="pointer-events-auto relative flex w-full items-center gap-5 rounded-md px-4 pt-1 pb-3"
	>
		<Play />
		<Volume />
		<Timestamp />
		<div class="flex-1"></div>
		<Settings />
		<Fullscreen />
	</media-controls-group>

	<div
		class="pointer-events-none absolute bottom-0 left-0 z-[-1] h-[99px] w-full rounded-b-md [background-image:_url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAADGCAYAAAAT+OqFAAAAdklEQVQoz42QQQ7AIAgEF/T/D+kbq/RWAlnQyyazA4aoAB4FsBSA/bFjuF1EOL7VbrIrBuusmrt4ZZORfb6ehbWdnRHEIiITaEUKa5EJqUakRSaEYBJSCY2dEstQY7AuxahwXFrvZmWl2rh4JZ07z9dLtesfNj5q0FU3A5ObbwAAAABJRU5ErkJggg==)] bg-bottom bg-repeat-x"
	></div>
</media-controls>
