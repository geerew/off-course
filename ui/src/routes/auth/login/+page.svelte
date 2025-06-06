<script lang="ts">
	import { Logo, Spinner } from '$lib/components';
	import FormLogin from '$lib/components/form/form-login.svelte';
	import { WarningIcon } from '$lib/components/icons';
	import { Button } from '$lib/components/ui';

	let signupEnabled = $state(false);
	let loadPromise = $state(fetchSignupStatus());

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	async function fetchSignupStatus(): Promise<void> {
		try {
			const response = await fetch('/api/auth/signup-status');
			if (!response.ok) throw new Error('Failed to fetch sign-up status');
			const data = (await response.json()) as { enabled: boolean };
			signupEnabled = data.enabled;
		} catch (error) {
			console.error('Error fetching signup status:', error);
			throw error;
		}
	}
</script>

<div class="flex h-dvh min-h-110 flex-col items-center justify-center gap-8">
	{#await loadPromise}
		<div class="flex justify-center pt-10">
			<Spinner class="bg-foreground-alt-3 size-4" />
		</div>
	{:then _}
		<Logo />

		<div class="flex min-w-sm flex-col gap-5">
			<div class="mb-2.5 flex flex-col gap-2">
				<div class="text-foreground-alt-1 text-center text-lg">Login to your account</div>
			</div>

			<FormLogin endpoint="/api/auth/login" />

			{#if signupEnabled}
				<div class="text-foreground-alt-3 text-center">
					Don't have an account?
					<Button
						href="/auth/register/"
						variant="ghost"
						class="hover:text-background-primary h-auto px-1"
						data-sveltekit-reload
					>
						Sign up
					</Button>
				</div>
			{/if}
		</div>
	{:catch error}
		<div class="flex w-full flex-col items-center gap-2 pt-10">
			<WarningIcon class="text-foreground-error size-10" />
			<span class="text-lg">{error.message}</span>
		</div>
	{/await}
</div>
