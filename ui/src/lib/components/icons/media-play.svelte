<script lang="ts">
	import type { SVGAttributes } from 'svelte/elements';
	let restProps: SVGAttributes<SVGElement> = $props();
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	width="24"
	height="24"
	viewBox="0 0 24 24"
	stroke="currentColor"
	{...restProps}
>
	<polygon points="6 3 20 12 6 21 6 3" />
</svg>
