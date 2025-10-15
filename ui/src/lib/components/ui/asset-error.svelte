<script lang="ts">
	import { FilesIcon, RefreshIcon, VideoIcon, WarningIcon } from '$lib/components/icons';
	import { Button } from '$lib/components/ui';
	import { cn } from '$lib/utils';

	interface Props {
		assetType: 'markdown' | 'text' | 'video';
		assetTitle?: string;
		error?: string;
		onRetry?: () => void;
		class?: string;
	}

	let { assetType, assetTitle, error, onRetry, class: className }: Props = $props();
</script>

<div class={cn('flex flex-col items-center gap-4 p-8 text-center', className)}>
	<div class="flex flex-col items-center gap-3">
		<div class="bg-background-alt-2 flex size-12 items-center justify-center rounded-full">
			{#if error?.includes('not found') || error?.includes('does not exist')}
				{#if assetType === 'video'}
					<VideoIcon class="text-foreground-alt-3 size-6" />
				{:else}
					<FilesIcon class="text-foreground-alt-3 size-6" />
				{/if}
			{:else}
				<WarningIcon class="text-foreground-error size-6" />
			{/if}
		</div>

		<div class="flex flex-col gap-1">
			<h3 class="text-foreground text-lg font-medium">
				{#if error?.includes('not found') || error?.includes('does not exist')}
					Content Not Available
				{:else}
					Failed to Load Content
				{/if}
			</h3>
		</div>
	</div>

	<div class="text-foreground-alt-3 flex flex-col gap-3 text-sm">
		{#if error?.includes('not found') || error?.includes('does not exist')}
			<p>This asset appears to be missing or has been moved.</p>
		{:else}
			<p>
				There was an error loading this {assetType === 'video' ? 'video' : assetType} content. This could
				be due to a network issue or server problem.
			</p>
		{/if}
	</div>

	{#if onRetry}
		<Button variant="outline" class="h-auto gap-2 px-3 py-1.5 text-sm" onclick={onRetry}>
			<RefreshIcon class="size-4" />
			Try Again
		</Button>
	{/if}
</div>
