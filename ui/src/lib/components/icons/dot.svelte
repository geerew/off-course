<script lang="ts">
	import type { HTMLAttributes } from 'svelte/elements';

	type Props = HTMLAttributes<HTMLSpanElement> & {
		class?: string;
	};

	let { class: className = '', ...restProps }: Props = $props();
</script>

<span class={className} {...restProps}>·</span>
