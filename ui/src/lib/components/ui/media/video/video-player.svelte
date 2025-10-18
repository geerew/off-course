<script lang="ts">
	import { mediaPreferences } from '$lib/preferences.svelte';
	import type {
		HLSManifestLoadedEvent,
		HLSProvider,
		MediaDurationChangeEvent,
		MediaProviderChangeEvent,
		MediaRateChangeEvent,
		MediaSourceChangeEvent,
		MediaTimeUpdateEvent,
		MediaVolumeChangeEvent,
		VideoMimeType
	} from 'vidstack';
	import 'vidstack/bundle';
	import type { MediaPlayerElement } from 'vidstack/elements';
	import MobileControlsLayout from './mobile-controls-layout.svelte';
	import NormalControlsLayout from './normal-controls-layout.svelte';
	import Buffering from './ui/components/buffering.svelte';
	import Gestures from './ui/components/gestures.svelte';
	import { videoStateManager } from './video-state-manager';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		src: string;
		srcType?: VideoMimeType;
		startTime: number;
		onTimeChange: (time: number) => void;
		onCompleted: (time: number) => void;
		playerId?: string;
		useHls?: boolean; // New prop to enable HLS streaming
	};

	let {
		src: videoSrc = $bindable(),
		srcType = 'video/mp4',
		startTime,
		onTimeChange,
		onCompleted,
		playerId,
		useHls = false
	}: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let player: MediaPlayerElement;
	let duration = -1;
	let lastLoggedSecond = -1;
	let completeDispatched = false;
	let uniqueId: string;

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Generate a unique ID for this player instance
	$effect(() => {
		uniqueId = playerId || `video-${Math.random().toString(36).substr(2, 9)}`;
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Handle play event - register with state manager and pause others
	function handlePlay() {
		if (!player || !uniqueId) return;

		// Register this player with the state manager
		videoStateManager.register(uniqueId, () => {
			if (player) {
				player.pause();
			}
		});

		// Set this as the current player and pause others
		videoStateManager.setCurrentPlayer(uniqueId);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Handle pause event - clear current player if this was the active one
	function handlePause() {
		if (!uniqueId) return;

		// Only clear if this was the current player
		if (videoStateManager.getCurrentPlayerId() === uniqueId) {
			videoStateManager.clearCurrentPlayer();
		}
	}

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

		const sec = Math.floor(e.detail.currentTime);
		if (sec === 0) return;

		// If video is < 5s long, consider it "complete" at (duration - 1s),
		// otherwise at (duration - 5s)
		const nearEndThreshold = duration < 5 ? 1 : 5;

		// Fire completion as soon as we cross the near-end threshold.
		if (sec >= Math.max(0, Math.floor(duration) - nearEndThreshold)) {
			if (!completeDispatched) {
				completeDispatched = true;
				onCompleted(Math.max(1, Math.ceil(duration)));
			}
			return;
		}

		completeDispatched = false;

		// Throttle progress: only every 3 seconds
		if (sec % 3 !== 0 || sec === lastLoggedSecond) return;

		lastLoggedSecond = sec;
		onTimeChange(sec);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the currentTime when the video can play
	function canPlay() {
		if (!player) return;

		player.currentTime = Math.floor(startTime) == Math.floor(duration) ? 0 : startTime;

		// This is a workaround for PR https://github.com/vidstack/player/issues/1416
		setTimeout(() => {
			if (player) {
				player.autoPlay = mediaPreferences.current.autoplay;
				player.playbackRate = mediaPreferences.current.playbackRate;
				player.volume = mediaPreferences.current.volume;
				player.muted = mediaPreferences.current.muted;
			}
		}, 0);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the duration. This will be called when the src changes
	function durationChange(e: MediaDurationChangeEvent) {
		duration = e.detail;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update the preferences with the playback rate
	function rateChange(e: MediaRateChangeEvent) {
		mediaPreferences.current.playbackRate = e.detail;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update the preferences with the volume
	function volumeChange(e: MediaVolumeChangeEvent) {
		mediaPreferences.current.volume = e.detail.volume;
		mediaPreferences.current.muted = e.detail.muted;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function handleProviderChange(e: MediaProviderChangeEvent) {
		const provider = e.detail;

		if (provider?.type === 'hls') {
			console.log('providerchange');
			const hlsProvider = provider as HLSProvider;
			hlsProvider.library = () => import('hls.js');

			hlsProvider.config = {
				// startPosition: startTime,

				autoStartLoad: false, // Let HLS.js handle loading automatically
				startLevel: Number.POSITIVE_INFINITY, // Auto quality selection
				abrEwmaDefaultEstimate: 35_000_000, // Bandwidth estimation
				lowLatencyMode: false,

				// Buffer configuration to limit segments
				maxBufferLength: 30, // Buffer up to 30 seconds ahead
				maxMaxBufferLength: 60, // Do not exceed 60 seconds of buffer
				maxBufferSize: 30 * 1024 * 1024, // Limit buffer size to 30 MB
				liveSyncDurationCount: 3, // For live streams, sync to 3 segments
				liveMaxLatencyDurationCount: 5, // Max latency of 5 segments

				// Fragment loading policy
				fragLoadPolicy: {
					default: {
						maxTimeToFirstByteMs: Number.POSITIVE_INFINITY,
						maxLoadTimeMs: 60_000,
						timeoutRetry: { maxNumRetry: 5, retryDelayMs: 100, maxRetryDelayMs: 0 },
						errorRetry: { maxNumRetry: 5, retryDelayMs: 0, maxRetryDelayMs: 100 }
					}
				}
			};
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	function handleHlsManifestLoaded(e: HLSManifestLoadedEvent) {
		console.log('HLS: Manifest loaded');
		console.log(e.detail);

		if (!player || player.provider === null || player.provider.type !== 'hls') return;

		console.log('HLS: Starting load');
		console.log(startTime);

		player.provider.instance?.startLoad(startTime);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		if (!player) return;

		player.addEventListener('source-change', sourceChange);
		player.addEventListener('can-play', canPlay);
		player.addEventListener('time-update', timeChange);
		player.addEventListener('duration-change', durationChange);
		player.addEventListener('rate-change', rateChange);
		player.addEventListener('volume-change', volumeChange);
		player.addEventListener('play', handlePlay);
		player.addEventListener('pause', handlePause);

		// hls
		player.addEventListener('provider-change', handleProviderChange);
		player.addEventListener('hls-manifest-loaded', handleHlsManifestLoaded);

		return () => {
			player.removeEventListener('source-change', sourceChange);
			player.removeEventListener('can-play', canPlay);
			player.removeEventListener('time-update', timeChange);
			player.removeEventListener('duration-change', durationChange);
			player.removeEventListener('rate-change', rateChange);
			player.removeEventListener('volume-change', volumeChange);
			player.removeEventListener('play', handlePlay);
			player.removeEventListener('pause', handlePause);
			player.removeEventListener('provider-change', handleProviderChange);
			player.removeEventListener('hls-manifest-loaded', handleHlsManifestLoaded);

			// Unregister from state manager when component is destroyed
			if (uniqueId) {
				videoStateManager.unregister(uniqueId);
			}
		};
	});
</script>

<!-- TODO Handle src.type instead of hardcoding -->
<div class="transform-gpu backface-hidden">
	<media-player
		bind:this={player}
		playsInline
		autoplay={mediaPreferences.current.autoplay}
		src={{
			src: videoSrc,
			type: useHls ? ('application/x.mpegurl' as any) : srcType
		}}
		class="group/player relative aspect-video overflow-hidden rounded-md"
	>
		<media-provider></media-provider>

		<Gestures />
		<Buffering />

		<!-- Shown when pointer=fine -->
		<NormalControlsLayout />

		<!-- Shown when pointer=coarse -->
		<MobileControlsLayout />
	</media-player>
</div>
