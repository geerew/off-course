<script lang="ts">
	import { DropdownMenu } from 'bits-ui';

	let {
		ref = $bindable(null),
		value = $bindable(),
		...restProps
	}: DropdownMenu.CheckboxGroupProps = $props();
</script>

<DropdownMenu.CheckboxGroup
	bind:ref
	bind:value
	data-slot="dropdown-menu-checkbox-group"
	{...restProps}
/>
