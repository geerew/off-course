<script lang="ts">
	import { cn } from '$lib/utils';

	let { class: containerClass = '', children } = $props();
</script>

<tr class={cn('border-background-alt-3 border-b', containerClass)}>
	{@render children?.()}
</tr>
