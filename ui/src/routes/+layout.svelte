<script lang="ts">
	import '../app.css';

	import { dev } from '$app/environment';
	import { page } from '$app/state';
	import { auth } from '$lib/auth.svelte';
	import { TailwindIndicator } from '$lib/components';
	import { Header, Oops, Spinner } from '$lib/components/';
	import { Tooltip } from 'bits-ui';
	import { Toaster } from 'svelte-sonner';

	let { children } = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	const isAuthPath = $derived(page.url.pathname.startsWith('/auth'));

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	$effect(() => {
		page.url.pathname;
		if (isAuthPath) return;
		auth.me();
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
				<div class="min-h-dvh pt-[calc(var(--header-height)+1px)]">
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
