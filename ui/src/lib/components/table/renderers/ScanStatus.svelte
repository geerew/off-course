<script lang="ts">
	import { ErrorMessage, GetCourse } from '$lib/api';
	import { addToast } from '$lib/stores/addToast';
	import type { ClassName } from '$lib/types/general';
	import type { Course } from '$lib/types/models';
	import { cn } from '$lib/utils';
	import { createEventDispatcher, onDestroy, onMount } from 'svelte';

	let className: ClassName = undefined;

	// ----------------------
	// Exports
	// ----------------------
	export let courseId: string;
	export let scanStatus: string;
	export let waitingText = 'queued';
	export let processingText = 'processing';
	export { className as class };

	// ----------------------
	// Variables
	// ----------------------

	// When the scanStatus is not empty, this will set. It is used to determine if we should stop
	// polling during onDestroy()
	let scanPoll = -1;

	// Dispatch events to as the status changes
	const dispatch = createEventDispatcher<Record<'change', Course>>();

	// ----------------------
	// Functions
	// ----------------------

	// When the scan status is set to either waiting or processing, start polling for updates. As
	// the status changes, we dispatch an event to the parent component to update the courses list.
	// When the scan finishes, we clear the interval and set the status to an empty string.
	const startPolling = () => {
		scanPoll = setInterval(async () => {
			await GetCourse(courseId)
				.then((resp) => {
					if (resp) {
						if (resp.scanStatus !== scanStatus) {
							// Scan status changed
							scanStatus = resp.scanStatus;
							dispatch('change', resp);

							if (scanStatus === '') {
								// Scan finished
								clearInterval(scanPoll);
								scanPoll = -1;
							}
						}
					}
				})
				.catch((err) => {
					const errMsg = ErrorMessage(err);
					console.error(errMsg);
					$addToast({
						data: {
							message: errMsg,
							status: 'error'
						}
					});

					scanStatus = '';
					clearInterval(scanPoll);
					scanPoll = -1;
				});
		}, 1500);
	};

	// ----------------------
	// Reactive
	// ----------------------
	$: scanStatus && scanPoll === -1 && startPolling();

	// ----------------------
	// Lifecycle
	// ----------------------
	onMount(() => {
		if (scanStatus && scanPoll === -1) {
			startPolling();
		}
	});

	onDestroy(() => {
		if (scanPoll !== -1) {
			clearInterval(scanPoll);
			scanPoll = -1;
		}
	});
</script>

<div class={cn('text-muted-foreground flex items-center justify-center', className)}>
	{#if !scanStatus}
		-
	{:else if scanStatus === 'waiting'}
		{waitingText}
	{:else}
		<span class="text-secondary">{processingText}</span>
	{/if}
</div>
