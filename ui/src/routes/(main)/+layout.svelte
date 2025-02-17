<script lang="ts">
	import { auth } from '$lib/auth.svelte';
	import { Error, Header, Spinner } from '$lib/components/';

	let { children } = $props();

	$effect(() => {
		auth.me();

		const interval = setInterval(() => {
			auth.me();
		}, 3000);

		return () => clearInterval(interval);
	});
</script>

{#if auth.error !== null}
	<div class="container-px flex w-full">
		<Error message={'Failed to fetch user: ' + auth.error} />
	</div>
{:else if auth.user === null}
	<div class="flex w-full justify-center pt-14 sm:pt-20">
		<Spinner class="bg-foreground-alt-2 size-6" />
	</div>
{:else}
	<Header />
	<div class="min-h-dvh pt-[calc(var(--height-header)+1px)]">
		{@render children()}
	</div>
{/if}
