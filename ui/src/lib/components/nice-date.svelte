<script lang="ts">
	import { cn } from '$lib/utils';
	import ago from 's-ago';

	type Props = {
		date: string;
		prefix?: string;
		class?: string;
	};

	let { date, prefix = '', class: componentClass }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	const d = $derived(new Date(date));
	const formattedDate = $derived(
		d.toLocaleDateString(undefined, {
			weekday: 'long',
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: 'numeric',
			second: 'numeric'
		})
	);
</script>

<div class={cn('text-muted-foreground', componentClass)}>
	<span title={prefix !== '' ? `${prefix} ${formattedDate}` : formattedDate}>
		{ago(d)}
	</span>
</div>
