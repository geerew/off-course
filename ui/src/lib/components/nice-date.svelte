<script lang="ts">
	import { cn } from '$lib/utils';
	import ago from 's-ago';

	type Props = {
		date: string;
		prefix?: string;
		class?: string;
	};

	let { date, prefix = '', class: componentClass }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// ----------------------
	// Variables
	// ----------------------
	const d = new Date(date);
	const utcDate = new Date(d.getTime() + d.getTimezoneOffset() * 60000);

	const formattedDate = utcDate.toLocaleDateString(undefined, {
		weekday: 'long',
		year: 'numeric',
		month: 'short',
		day: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		second: 'numeric'
	});
</script>

<div class={cn('text-muted-foreground', componentClass)}>
	<span title={prefix !== '' ? `${prefix} ${formattedDate}` : formattedDate}>
		{ago(utcDate)}
	</span>
</div>
