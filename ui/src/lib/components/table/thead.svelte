<script lang="ts">
	import { cn } from '$lib/utils';

	let { class: containerClass = '', children } = $props();
</script>

<thead
	class={cn(
		'border-background-alt-3 text-foreground-alt-2 border-b text-xs uppercase',
		containerClass
	)}
>
	{@render children?.()}
</thead>
