<script lang="ts">
	import { ASSET_API, GetBackendUrl } from '$lib/api';
	import type { Asset } from '$lib/types/models';
	import { createEventDispatcher } from 'svelte';
	import {
		type MediaDurationChangeEvent,
		type MediaSourceChangeEvent,
		type MediaTimeUpdateEvent
	} from 'vidstack';
	import 'vidstack/bundle';
	import type { MediaPlayerElement } from 'vidstack/elements';
	import 'vidstack/icons';
	import BufferingIndicator from './_internal/components/buffering-indicator.svelte';
	import Gestures from './_internal/gestures.svelte';
	import MobileControlsLayout from './_internal/mobile-controls-layout.svelte';
	import NormalControlsLayout from './_internal/normal-controls-layout.svelte';
	import { isSettingsMenuOpen, preferences } from './_internal/store';

	// ----------------------
	// Exports
	// ----------------------

	export let title: string;
	export let src: string;
	export let startTime = 0;
	export let nextAsset: Asset | null;

	// ----------------------
	// Variables
	// ----------------------

	// The player element
	let player: MediaPlayerElement;

	const dispatch = createEventDispatcher<Record<'progress' | 'complete', number>>();
	const dispatchNext = createEventDispatcher();

	// Current time of the player
	let currentTime = -1;

	// Used to only do stuff when the logged second changes
	let lastLoggedSecond = -1;

	// True when the completed event is dispatched
	let completeDispatched = false;

	// Set by the player
	let duration = -1;

	// When loading the component store the local storage values. We do this because vidstack tries to set
	// them to defaults like volume 1 and muted false. So until the bug is fixed, we force override
	let storagePlaybackRate = $preferences.playbackRate ?? 1;
	let storageVolume = $preferences.volume ?? 1;
	let storageMuted = $preferences.muted ?? false;

	// Holds the current timeout for hiding the controls (when the pointer is coarse)
	let courseControlsTimeout: number = -1;

	// ----------------------
	// Functions
	// ----------------------

	// Called when the source changes. Resets the logged second and completed state
	function srcChange(e: MediaSourceChangeEvent) {
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

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the currentTime when the video can play
	function canPlay() {
		if (!player) return;

		if (Math.floor(startTime) == Math.floor(duration)) {
			player.currentTime = 0;
		} else {
			player.currentTime = startTime ?? 0;
		}

		// This is a temporary fix until vidstack respects setting these in the <media-player>
		setTimeout(() => {
			player.volume = storageVolume;
			player.muted = storageMuted;
			player.playbackRate = storagePlaybackRate;
		}, 0);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the duration. Update when the src changes
	function durationChange(e: MediaDurationChangeEvent) {
		duration = e.detail;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Each second a progress event will be dispatched. Additionally, when the time is > 3 seconds
	// a started event will be dispatched and when the time is > duration - 5 seconds, a completed
	// event will be dispatched
	//
	// When the video is paused, nothing will happen
	function timeChange(e: MediaTimeUpdateEvent) {
		if (duration === -1) return;

		currentTime = e.detail.currentTime;

		// Clear the ended state when the current time changes
		// if ($ctx.ended && currentTime !== duration) ctx.set({ ...$ctx, ended: false });

		// Do nothing when we have already processed this second
		const currentSecond = Math.floor(currentTime);
		if (currentSecond === 0 || currentSecond === lastLoggedSecond) return;
		lastLoggedSecond = currentSecond;

		// For the last 5 seconds of the video, dispatch the completed event. After dispatching the
		// event, completeDispatched will be set to true, so we do not dispatch the event again.
		// Prior dispatch common progress events. This will set completeDispatched to false
		if (currentSecond >= duration - 5) {
			if (completeDispatched) return;
			dispatch('complete', Math.floor(duration));
		} else {
			completeDispatched = false;
			dispatch('progress', currentSecond);
		}
	}

	// ----------------------
	// Reactive
	// ----------------------

	// When the settings menu is open clear the courseControlsTimeout
	$: if ($isSettingsMenuOpen && courseControlsTimeout !== -1) {
		clearTimeout(courseControlsTimeout);
	}
</script>

<media-player
	bind:this={player}
	{title}
	playsInline
	autoPlay={$preferences.autoplay}
	src={{
		src: GetBackendUrl(ASSET_API) + '/' + src + '/serve',
		type: 'video/mp4'
	}}
	playbackRate={storagePlaybackRate}
	volume={storageVolume}
	muted={storageMuted}
	on:source-change={srcChange}
	on:can-play={canPlay}
	on:duration-change={durationChange}
	on:time-update={timeChange}
	on:ended={() => {
		if (player && nextAsset && $preferences.autoloadNext && player.duration !== 0) {
			dispatchNext('next');
		}
	}}
	on:controls-change={(e) => {
		if (!player) return;

		// If the pointer is coarse, when the controls show, set a timer to hid them again
		// after 2.5 seconds. The timeout will be cancelled when the video settings menu
		// is open and restarted when closed
		const pointerValue = player.getAttribute('data-pointer');
		if (e.detail && pointerValue === 'coarse') {
			courseControlsTimeout = setTimeout(() => {
				if (player && pointerValue === 'coarse') {
					player.remoteControl.resumeControls();
				}
			}, 2500);
		}
	}}
	class="group/player"
>
	<media-provider />

	<Gestures />

	<BufferingIndicator />

	<NormalControlsLayout />
	<MobileControlsLayout />
</media-player>
