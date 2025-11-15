<script lang="ts">
	import {
		ArrowLeftIcon,
		LeftChevronIcon,
		MediaSettingsIcon,
		RightChevronIcon
	} from '$lib/components/icons';
	import MediaSettingPlayback from '$lib/components/icons/media-setting-playback.svelte';
	import { Button, Drawer, Popover } from '$lib/components/ui';
	import Switch from '$lib/components/ui/switch.svelte';
	import { mediaPreferences } from '$lib/preferences.svelte';
	import { Separator, Slider } from 'bits-ui';
	import { tick } from 'svelte';
	import { cubicOut } from 'svelte/easing';
	import { fly } from 'svelte/transition';
	import { MediaRemoteControl } from 'vidstack';

	let open = $state(false);

	const remote = new MediaRemoteControl();

	// Local UI state
	let panel = $state<'home' | 'playback'>('home');

	// Which way the *incoming* panel slides
	//   1: from right (home -> playback)
	//  -1: from left  (playback -> home)
	//   0: no slide (first open)
	let slideDir = $state<-1 | 0 | 1>(0);

	// Size measurement (for width/height tween on content container)
	let panelW = $state(0);
	let panelH = $state(0);

	let settingsEl: HTMLDivElement | null = $state(null);
	let innerEl: HTMLDivElement | null = $state(null);

	let doResizeAnimation = $state(false);

	let pointerType = $state<'coarse' | 'fine' | 'none'>('none');

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Detects the pointer type
	function detectPointer() {
		if (window.matchMedia('(any-pointer: coarse)').matches) {
			pointerType = 'coarse';
		} else if (window.matchMedia('(any-pointer: fine)').matches) {
			pointerType = 'fine';
		} else {
			pointerType = 'none';
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Called when the dialog/drawer open state completes
	async function onOpenChange(open: boolean) {
		if (open) {
			await tick();

			if (!innerEl) return;
			panelW = innerEl.clientWidth;
			panelH = innerEl.clientHeight;

			requestAnimationFrame(() => {
				doResizeAnimation = true;
			});
		} else {
			panel = 'home';
			slideDir = 0;
			doResizeAnimation = false;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		remote.changePlaybackRate(mediaPreferences.current.playbackRate);
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// When the popover is open, pause controls
	$effect(() => {
		if (open) {
			remote.pauseControls();
		} else {
			remote.resumeControls();
		}
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Get the player for the remote
	$effect(() => {
		const player = remote.getPlayer(settingsEl);
		if (!player) return;

		detectPointer();

		const coarse = window.matchMedia('(any-pointer: coarse)');
		const fine = window.matchMedia('(any-pointer: fine)');

		coarse.addEventListener('change', detectPointer);
		fine.addEventListener('change', detectPointer);

		// Unsubscribe
		return () => {
			remote.resumeControls();
			coarse.removeEventListener('change', detectPointer);
			fine.removeEventListener('change', detectPointer);
		};
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Keep the container sized if the user resizes text/viewport while open
	$effect(() => {
		if (!innerEl) return;

		const r = new ResizeObserver(() => {
			panelW = innerEl!.clientWidth;
			panelH = innerEl!.clientHeight;
		});

		r.observe(innerEl);
		return () => r.disconnect();
	});
</script>

{#snippet contents()}
	<div
		class="inline-grid w-fit min-w-0 max-w-[90vw]"
		bind:this={innerEl}
		bind:clientWidth={panelW}
		bind:clientHeight={panelH}
	>
		{#key panel}
			<div
				in:fly={{
					x: slideDir === 0 ? 0 : slideDir > 0 ? 100 : -100,
					y: 0,
					opacity: 0,
					duration: 250,
					easing: cubicOut
				}}
				style="will-change: transform, opacity"
			>
				{#if panel === 'home'}
					<div class="flex flex-col px-1 py-2">
						<Button
							variant="ghost"
							class="hover:bg-background-alt-2 w-50 justify-between"
							onclick={() => {
								slideDir = 1;
								panel = 'playback';
							}}
						>
							<div class="flex items-center gap-2.5">
								<MediaSettingPlayback class="size-4 stroke-[1.5]" />
								<span>Playback</span>
							</div>

							<RightChevronIcon class="text-foreground/50 size-3" />
						</Button>
					</div>
				{:else if panel === 'playback'}
					<Button
						variant="ghost"
						class="hover:bg-background-alt-2 w-60 justify-start gap-4"
						onclick={() => {
							slideDir = -1;
							panel = 'home';
						}}
					>
						<ArrowLeftIcon class="text-foreground size-4 stroke-2" />
						<div class="flex items-center gap-2.5">
							<MediaSettingPlayback class="size-4 stroke-[1.5]" />
							<span>Playback</span>
						</div>
					</Button>

					<div class="flex flex-col gap-3 px-1 py-2">
						<!-- Autoplay -->
						<div class="bg-background-alt-2 flex items-center justify-between px-1 py-1">
							<Switch
								labelText="Autoplay"
								checked={mediaPreferences.current.autoplay}
								onCheckedChange={(value: boolean) => {
									mediaPreferences.current.autoplay = value;
								}}
							/>
						</div>

						<!-- Autoload Next -->
						<div class="bg-background-alt-2 flex items-center justify-between px-1 py-1">
							<Switch
								labelText="Autoload Next"
								checked={mediaPreferences.current.autoloadNext}
								onCheckedChange={(value: boolean) => {
									mediaPreferences.current.autoloadNext = value;
								}}
							/>
						</div>

						<!-- Rate -->
						<div class="flex flex-col gap-2">
							<div class="text-muted-foreground/80 flex flex-row items-center justify-between">
								<span class="pl-1">Speed</span>
								<span class="text-foreground-alt-2 pr-2.5 text-xs"
									>{mediaPreferences.current.playbackRate === 1
										? 'Normal'
										: mediaPreferences.current.playbackRate + 'x'}</span
								>
							</div>

							<div class="bg-background-alt-2 flex items-center justify-between px-1 py-3.5">
								<LeftChevronIcon class="size-4 text-white/70" />
								<Slider.Root
									type="single"
									class="group relative flex w-full touch-none select-none items-center py-3 hover:cursor-pointer"
									bind:value={mediaPreferences.current.playbackRate}
									min={0.25}
									max={2}
									step={[0.25, 0.5, 0.75, 1, 1.25, 1.5, 1.75, 2]}
								>
									{#snippet children({ thumbItems, tickItems })}
										<span
											class="relative h-[5px] w-full grow cursor-pointer overflow-hidden rounded-full bg-white/30"
										>
											<Slider.Range
												class="bg-background-primary-alt-1 absolute h-full opacity-100 transition-opacity duration-200 group-hover:opacity-0"
											/>
										</span>

										{#each tickItems as { index, value } (index)}
											<Slider.Tick
												{index}
												class="bg-foreground-alt-2 z-1 h-[5px] w-0.5 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
											/>
										{/each}

										{#each thumbItems as { index } (index)}
											<Slider.Thumb
												{index}
												class="bg-foreground-alt-1 focus-visible:outline-hidden z-10 block size-[15px] cursor-pointer rounded-full shadow-sm transition-colors focus-visible:ring-2 focus-visible:ring-white/40 focus-visible:ring-offset-2"
											/>
										{/each}
									{/snippet}
								</Slider.Root>
								<RightChevronIcon class="size-4 text-white/70" />
							</div>
						</div>
					</div>
				{/if}
			</div>
		{/key}
	</div>
{/snippet}

<div bind:this={settingsEl} class="flex">
	{#if pointerType === 'fine'}
		<Popover.Root bind:open onOpenChangeComplete={onOpenChange}>
			<!-- 
				A little hack to stop events from propagating to the video player when 
				clicking outside the popover but inside the video player
			-->
			{#if open}
				<Button
					variant="ghost"
					class="z-49 fixed inset-0 h-full w-full cursor-default bg-transparent hover:bg-transparent"
					onpointerup={() => (open = false)}
				></Button>
			{/if}

			<Popover.Trigger class="h-auto w-auto border-none px-0 [&[data-state=open]>svg]:rotate-90">
				<MediaSettingsIcon
					class="group-hover:fill-background-primary group-hover:stroke-background-primary size-7 stroke-white stroke-2"
				/>
			</Popover.Trigger>

			<Popover.Content
				class="drop-shadow-black/70 z-50 w-auto origin-bottom-right overflow-hidden border-none drop-shadow-md"
				side="top"
				align="end"
				sideOffset={8}
				preventScroll={false}
				portalProps={{ disabled: false }}
				interactOutsideBehavior="ignore"
			>
				<div
					style={`transition:${
						doResizeAnimation
							? `width 250ms cubic-bezier(.2,.8,.2,1), height 250ms cubic-bezier(.2,.8,.2,1)`
							: 'none'
					}; width:${panelW}px; height:${panelH}px; will-change:width,height;`}
				>
					{@render contents()}
				</div>
			</Popover.Content>
		</Popover.Root>
	{:else if pointerType === 'coarse'}
		<Drawer.Root bind:open onOpenChangeComplete={onOpenChange}>
			<Drawer.Trigger
				class="h-auto w-auto border-none bg-transparent px-0 [&[data-state=open]>svg]:rotate-90"
			>
				<MediaSettingsIcon
					class="group-hover:fill-background-primary group-hover:stroke-background-primary size-5 stroke-white sm:size-6 md:size-7"
				/>
			</Drawer.Trigger>

			<Drawer.Content class="bg-background-alt-2" handleClass="bg-background-alt-4">
				<div class="flex place-self-center py-3">
					{@render contents()}
				</div>
			</Drawer.Content>
		</Drawer.Root>
	{/if}
</div>
