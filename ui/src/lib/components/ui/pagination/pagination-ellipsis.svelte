<script lang="ts">
	import { Icons } from '$components/icons';
	import { cn } from '$lib/utils';
	import type { HTMLAttributes } from 'svelte/elements';

	type $$Props = HTMLAttributes<HTMLSpanElement>;

	let className: $$Props['class'] = undefined;
	export { className as class };
</script>

<span
	aria-hidden
	class={cn('flex h-9 w-9 items-center justify-center', className)}
	{...$$restProps}
>
	<Icons.MoreHorizontal class="h-4 w-4" />
	<span class="sr-only">More pages</span>
</span>
