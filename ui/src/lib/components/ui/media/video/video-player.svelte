<!-- TODO Add settings menu -->
<!-- TODO Persist things from the settings menu, volume, muted in local storage -->
<script lang="ts">
	import type {
		MediaDurationChangeEvent,
		MediaSourceChangeEvent,
		MediaTimeUpdateEvent
	} from 'vidstack';
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
		startTime: number;
		onTimeChange: (time: number) => void;
		onCompleted: (time: number) => void;
	};

	let { src: videoSrc = $bindable(), startTime, onTimeChange, onCompleted }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let player: MediaPlayerElement;
	let duration = -1;
	let currentTime = -1;
	let lastLoggedSecond = -1;
	let completeDispatched = false;

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// SourceChange is called when the source of the player changes, allowing us to reset values and
	// set the current time to the start time
	function sourceChange(e: MediaSourceChangeEvent) {
		if (!e.detail) return;

		lastLoggedSecond = -1;
		completeDispatched = false;

		if (!player) return;

		if (Math.floor(startTime) == Math.floor(duration)) {
			player.currentTime = 0;
		} else {
			player.currentTime = startTime ?? 0;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// TimeChange is called when the current time of the player changes
	function timeChange(e: MediaTimeUpdateEvent) {
		if (duration === -1) return;

		const currentSecond = Math.floor(e.detail.currentTime);
		if (currentSecond === 0 || currentSecond === lastLoggedSecond) return;

		lastLoggedSecond = currentSecond;

		if (currentSecond >= duration - 5) {
			if (completeDispatched) return;
			completeDispatched = true;
			onCompleted(duration);
		} else {
			completeDispatched = false;
			onTimeChange(currentSecond);
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the currentTime when the video can play
	function canPlay() {
		if (!player) return;

		player.currentTime = Math.floor(startTime) == Math.floor(duration) ? 0 : startTime;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the duration. This will be called when the src changes
	function durationChange(e: MediaDurationChangeEvent) {
		duration = e.detail;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (!player) return;

		player.addEventListener('source-change', sourceChange);
		player.addEventListener('can-play', canPlay);
		player.addEventListener('time-update', timeChange);
		player.addEventListener('duration-change', durationChange);

		return () => {
			player.removeEventListener('source-change', sourceChange);
			player.removeEventListener('can-play', canPlay);
			player.removeEventListener('time-update', timeChange);
			player.removeEventListener('duration-change', durationChange);
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
			<!--<Settings isMobile={false} /> -->
		</media-controls-group>

		<div
			class="pointer-events-none absolute bottom-0 left-0 z-[-1] h-[99px] w-full [background-image:_url(data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAADGCAYAAAAT+OqFAAAAdklEQVQoz42QQQ7AIAgEF/T/D+kbq/RWAlnQyyazA4aoAB4FsBSA/bFjuF1EOL7VbrIrBuusmrt4ZZORfb6ehbWdnRHEIiITaEUKa5EJqUakRSaEYBJSCY2dEstQY7AuxahwXFrvZmWl2rh4JZ07z9dLtesfNj5q0FU3A5ObbwAAAABJRU5ErkJggg==)] bg-bottom bg-repeat-x"
		></div>
	</media-controls>
</media-player>
