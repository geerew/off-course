<script lang="ts">
	import type { SVGAttributes } from 'svelte/elements';
	let restProps: SVGAttributes<SVGElement> = $props();
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	fill="none"
	viewBox="0 0 24 24"
	stroke="currentColor"
	{...restProps}
>
	<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5" />
</svg>
