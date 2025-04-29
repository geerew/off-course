<script lang="ts">
	import '@fontsource/source-sans-pro/400.css';
	import '@fontsource/source-sans-pro/700.css';
	import '../app.css';

	import { dev } from '$app/environment';
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import { TailwindIndicator } from '$lib/components';
	import { Header, Oops, Spinner } from '$lib/components/';
	import { Tooltip } from 'bits-ui';
	import { Toaster } from 'svelte-sonner';

	let { children } = $props();

	const isAuthPath = page.url.pathname.startsWith('/auth');

	$effect(() => {
		if (isAuthPath) return;

		auth.me();

		const interval = setInterval(() => {
			auth.me();
		}, 3000);

		return () => clearInterval(interval);
	});
</script>

<Toaster theme="dark" richColors />

<main>
	<Tooltip.Provider>
		{#if !isAuthPath}
			{#if auth.error !== null}
				<div class="container-px flex w-full">
					<Oops message={'Failed to fetch user: ' + auth.error} />
				</div>
			{:else if auth.user === null}
				<div class="flex w-full justify-center pt-14 sm:pt-20">
					<Spinner class="bg-foreground-alt-3 size-6" />
				</div>
			{:else}
				<Header />
				<div class="min-h-dvh pt-[calc(var(--height-header)+1px)]">
					{@render children()}
				</div>
			{/if}
		{:else}
			{@render children()}
		{/if}
	</Tooltip.Provider>

	{#if dev}
		<TailwindIndicator />
	{/if}
</main>
