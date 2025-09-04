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
	import MobileControlsLayout from './mobile-controls-layout.svelte';
	import NormalControlsLayout from './normal-controls-layout.svelte';
	import Buffering from './ui/components/buffering.svelte';
	import Gestures from './ui/components/gestures.svelte';

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
			onCompleted(Math.ceil(duration));
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
	class="group/player relative overflow-hidden rounded-md"
>
	<media-provider></media-provider>

	<Gestures />
	<Buffering />

	<!-- Shown when pointer=fine -->
	<NormalControlsLayout />

	<!-- Shown when pointer=coarse -->
	<MobileControlsLayout />
</media-player>
