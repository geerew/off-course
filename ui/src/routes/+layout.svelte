<script lang="ts">
	import '@fontsource/source-sans-pro/400.css';
	import '@fontsource/source-sans-pro/700.css';
	import '../app.css';

	import { dev } from '$app/environment';
	import { TailwindIndicator } from '$lib/components';
	import { Toaster } from 'svelte-sonner';

	let { children } = $props();
</script>

<Toaster theme="dark" richColors />

<main>
	{@render children()}

	{#if dev}
		<TailwindIndicator />
	{/if}
</main>
