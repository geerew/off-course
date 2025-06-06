<script lang="ts">
	import { cn } from '$lib/utils';
	import type { Snippet } from 'svelte';
	import type { HTMLAttributes } from 'svelte/elements';
	import type { TransitionConfig } from 'svelte/transition';

	function noop(): TransitionConfig {
		return { delay: 0, duration: 0 };
	}

	type Props = HTMLAttributes<HTMLElement> & {
		class?: string;
		inTransition?: (node: Element, params?: any) => TransitionConfig;
		inTransitionParams?: any;
		outTransition?: (node: Element, params?: any) => TransitionConfig;
		outTransitionParams?: any;
		children?: Snippet;
	};

	let {
		class: containerClass = '',
		inTransition = noop,
		inTransitionParams,
		outTransition = noop,
		outTransitionParams,
		children,
		...restProps
	}: Props = $props();
</script>

<div
	role="cell"
	class={cn(
		'border-background-alt-3 flex min-h-14 items-center justify-center border-b py-2',
		containerClass
	)}
	in:inTransition={inTransitionParams}
	out:outTransition={outTransitionParams}
	{...restProps}
>
	{@render children?.()}
</div>
