<script lang="ts">
	import 'vidstack/bundle';
	import type { MediaPlayerElement } from 'vidstack/elements';
	import Buffering from './ui/buffering.svelte';
	import Fullscreen from './ui/fullscreen.svelte';
	import Gestures from './ui/gestures.svelte';
	import Play from './ui/play.svelte';
	import TimeSlider from './ui/time-slider.svelte';
	import Timestamp from './ui/timestamp.svelte';
	import Volume from './ui/volume.svelte';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		src: string;
	};

	let { src: videoSrc = $bindable() }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let player: MediaPlayerElement;

	$effect(() => {
		if (!player) return;
		// player.addEventListener('source-change', (e) => {});

		return () => {
			player.removeEventListener('source-change', (e) => {});
		};
	});
</script>

<!-- TODO Handle src.type instead of hardcoding -->
<media-player
	bind:this={player}
	playsInline
	src={{
		src: videoSrc,
		type: 'video/mp4'
	}}
	class="relative"
>
	<media-provider></media-provider>

	<Gestures />

	<Buffering />

	<media-controls
		class="pointer-events-none absolute inset-0 z-10 box-border flex h-full w-full flex-col opacity-0 transition-opacity duration-200 ease-out data-visible:opacity-100 data-visible:ease-in"
	>
		<div class="flex-1"></div>

		<media-controls-group class="pointer-events-auto flex w-full items-center px-3">
			<TimeSlider />
		</media-controls-group>

		<media-controls-group
			class="pointer-events-auto relative flex w-full items-center gap-5 px-4 pt-1 pb-3"
		>
			<Play />
			<Volume />
			<Timestamp />
			<div class="flex-1"></div>
			<Fullscreen />
			<!-- 


			<Settings isMobile={false} />
 -->
		</media-controls-group>

		<!-- Gradient bottom -->
		<div
			class="pointer-events-none absolute bottom-0 left-0 z-[-1] h-[99px] w-full [background-image:_url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAADGCAYAAAAT+OqFAAAAdklEQVQoz42QQQ7AIAgEF/T/D+kbq/RWAlnQyyazA4aoAB4FsBSA/bFjuF1EOL7VbrIrBuusmrt4ZZORfb6ehbWdnRHEIiITaEUKa5EJqUakRSaEYBJSCY2dEstQY7AuxahwXFrvZmWl2rh4JZ07z9dLtesfNj5q0FU3A5ObbwAAAABJRU5ErkJggg==)] bg-bottom bg-repeat-x"
		></div>
	</media-controls>
</media-player>
