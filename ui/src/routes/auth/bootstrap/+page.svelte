<script lang="ts">
	import { Logo } from '$lib/components';
</script>

<div class="flex h-dvh min-h-110 flex-col items-center justify-center gap-16">
	<Logo />

	<div class="flex min-w-sm flex-col gap-5 text-center">
		<div class="mb-2.5 flex flex-col gap-4">
			<div class="text-foreground-alt-1 text-xl">Application Setup Required</div>
			<div class="text-foreground-alt-2 max-w-lg text-base">
				Please check the server console for the bootstrap URL to set up your administrator account
			</div>
		</div>
	</div>
</div>
