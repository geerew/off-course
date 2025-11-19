<script lang="ts">
	import { GithubIcon } from '$lib/components/icons';
	import { GetVersion } from '$lib/api/version-api';
	import { Separator } from 'bits-ui';
	import { onMount } from 'svelte';

	let version = $state<string>('');
	let latestRelease = $state<string | undefined>(undefined);
	let isLoading = $state(true);
	let error = $state<string | null>(null);

	onMount(async () => {
		try {
			const data = await GetVersion();
			version = data.version;
			latestRelease = data.latestRelease;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load version';
		} finally {
			isLoading = false;
		}
	});
</script>

<div class="border-background-alt-3 flex w-full flex-col border-t py-8">
	<footer class="flex flex-col gap-2 px-2.5">
		<div class="flex items-center justify-between gap-2">
			<a
				href="https://github.com/geerew/offcourse/"
				target="_blank"
				rel="noopener noreferrer"
				class="text-foreground-alt-2 hover:text-foreground flex items-center gap-2 transition-colors"
			>
				<GithubIcon class="size-5 stroke-[1.5]" />
			</a>
			<div class="flex flex-col items-end gap-1">
				{#if !isLoading && !error}
					{@const isDev = version === 'dev' || version.startsWith('dev-')}
					{@const showCurrent = !isDev && latestRelease !== undefined}
					<div class="text-foreground-alt-2 text-xs">
						{#if showCurrent}
							Current: <span>{version}</span>
						{:else}
							<span>{version}</span>
						{/if}
					</div>
					{#if !isDev && latestRelease}
						<a
							href="https://github.com/geerew/offcourse/releases/tag/{latestRelease}"
							target="_blank"
							rel="noopener noreferrer"
							class="text-foreground-alt-3 hover:text-foreground-alt-2 text-xs transition-colors"
						>
							Latest: {latestRelease}
						</a>
					{/if}
				{:else if isLoading}
					<div class="text-foreground-alt-2 text-xs">
						<span>Loading...</span>
					</div>
				{:else if error}
					<div class="text-foreground-alt-2 text-xs">
						<span class="text-foreground-error">{error}</span>
					</div>
				{/if}
			</div>
		</div>
	</footer>
</div>
