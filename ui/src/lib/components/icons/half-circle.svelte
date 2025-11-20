<script lang="ts">
	import type { SVGAttributes } from 'svelte/elements';
	let restProps: SVGAttributes<SVGElement> = $props();
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	width="24"
	height="24"
	viewBox="0 0 24 24"
	fill="currentColor"
	{...restProps}
>
	<path d="M12 12 L12 2 A10 10 0 1 0 20.66 17 Z" />
</svg>
