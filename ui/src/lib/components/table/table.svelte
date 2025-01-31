<script lang="ts">
	import { cn } from '$lib/utils';

	let { class: containerClass = '', children } = $props();
</script>

<table class={cn('w-full border-collapse', containerClass)}>
	{@render children?.()}
</table>
