<script lang="ts">
	import { Icons } from '$components/icons';
	import { Button } from '$components/ui/button';
	import type { Builder } from 'bits-ui';

	// ----------------------
	// Exports
	// ----------------------
	export let builder: Builder;
</script>

<Button
	builders={[builder]}
	variant="ghost"
	class="group relative inline-flex h-auto cursor-pointer items-center px-0 py-0 text-white hover:bg-transparent hover:text-secondary"
>
	<Icons.Settings
		weight="fill"
		class="size-6 fill-white duration-200 group-hover:fill-secondary group-data-[state=open]:rotate-90"
	/>
</Button>
