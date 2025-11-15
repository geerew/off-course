<script lang="ts">
	import { MediaMaximizeIcon, MediaMinimizeIcon } from '$lib/components/icons';

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	type Props = {
		isMobile?: boolean;
	};

	let { isMobile = false }: Props = $props();
</script>

{#if isMobile}
	<media-fullscreen-button
		target="provider"
		class="data-focus:ring-4 group relative inline-flex cursor-pointer items-center justify-center rounded-md outline-none ring-inset ring-sky-400"
	>
		<MediaMaximizeIcon
			class="not-media-fullscreen:block group-hover:stroke-background-primary hidden size-5 stroke-white stroke-2 sm:size-6 md:size-7"
		/>
	</media-fullscreen-button>
{:else}
	<media-tooltip showDelay={300} class="contents">
		<media-tooltip-trigger>
			<media-fullscreen-button
				class="data-focus:ring-4 group relative inline-flex cursor-pointer items-center justify-center rounded-md outline-none ring-inset ring-sky-400"
			>
				<MediaMaximizeIcon
					class="not-media-fullscreen:block  group-hover:stroke-background-primary hidden size-7 stroke-white stroke-2"
				/>
				<MediaMinimizeIcon
					class="media-fullscreen:block group-hover:stroke-background-primary hidden size-7 stroke-white stroke-2"
				/>
			</media-fullscreen-button>
		</media-tooltip-trigger>
		<media-tooltip-content
			class="animate-out fade-out slide-out-to-bottom-2 data-visible:animate-in data-visible:fade-in data-visible:slide-in-from-bottom-4 z-10 rounded-sm border border-gray-400/20 bg-white px-1 py-0.5 text-sm font-medium text-black"
			placement="top end"
			offset={10}
		>
			<span class="not-media-fullscreen:block hidden">Fullscreen</span>
			<span class="media-fullscreen:block hidden">Exit Fullscreen</span>
		</media-tooltip-content>
	</media-tooltip>
{/if}
