<script lang="ts">
	import { DropdownMenu } from 'bits-ui';

	let {
		ref = $bindable(null),
		value = $bindable(),
		...restProps
	}: DropdownMenu.RadioGroupProps = $props();
</script>

<DropdownMenu.RadioGroup bind:ref bind:value data-slot="dropdown-menu-radio-group" {...restProps} />
