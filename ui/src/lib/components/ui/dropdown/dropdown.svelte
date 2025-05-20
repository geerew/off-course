<!-- TODO Rework to be more like drawer -->
<script lang="ts">
	import { DropdownMenu } from 'bits-ui';

	let { open = $bindable(false), ...restProps }: DropdownMenu.RootProps = $props();
</script>

<DropdownMenu.Root bind:open {...restProps} />
