<script lang="ts">
	import { Separator } from 'bits-ui';

	let { message } = $props();
</script>

<div class="flex w-full justify-center pt-20">
	<div
		class="border-background-alt-5 flex max-w-lg min-w-sm flex-col place-items-center gap-4 rounded-lg border p-5"
	>
		<svg
			xmlns="http://www.w3.org/2000/svg"
			fill="none"
			viewBox="0 0 24 24"
			stroke-width="1.5"
			stroke="currentColor"
			class="text-foreground-error size-10"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"
			/>
		</svg>

		<p class="text-foreground-alt-1 text-center text-lg">{message}</p>

		<Separator.Root class="bg-background-alt-5 h-px w-[15rem] shrink-0" />

		<div class="text-foreground-alt-2 flex justify-center gap-2.5">Please try again</div>
	</div>
</div>
