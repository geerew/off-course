<script lang="ts">
	let { class: containerClass = '', children } = $props();
</script>

<tbody class={containerClass}>
	{@render children?.()}
</tbody>
