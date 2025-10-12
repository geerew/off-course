<script lang="ts">
	import { Spinner } from '$lib/components';
	import { Button, Input, PasswordInput } from '$lib/components/ui';
	import type { AuthRegisterModel } from '$lib/models/auth-model';
	import { cn } from '$lib/utils';
	import { toast } from 'svelte-sonner';

	let { endpoint }: { endpoint: string } = $props();

	let username = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let posting = $state(false);
	let passwordMismatchError = $state(false);
	let previousPassword = $state('');
	let previousConfirmPassword = $state('');

	// Clear error when user types in either password field
	$effect(() => {
		if (
			passwordMismatchError &&
			(password !== previousPassword || confirmPassword !== previousConfirmPassword)
		) {
			passwordMismatchError = false;
		}
		previousPassword = password;
		previousConfirmPassword = confirmPassword;
	});

	async function submitForm(event: Event) {
		event.preventDefault();

		// Check if passwords match
		if (password !== confirmPassword) {
			passwordMismatchError = true;
			toast.error('Passwords do not match');
			return;
		}

		posting = true;

		const response = await fetch(endpoint, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				username,
				password
			} satisfies AuthRegisterModel)
		});

		if (response.ok) {
			window.location.href = '/';
		} else {
			const data = await response.json();
			toast.error(data.message);
			posting = false;
		}
	}
</script>

<form onsubmit={submitForm} class="flex flex-col gap-5">
	<Input bind:value={username} name="username" type="text" placeholder="Username" />
	<PasswordInput bind:value={password} placeholder="Password" />
	<PasswordInput
		bind:value={confirmPassword}
		placeholder="Confirm Password"
		class={cn(
			'transition-colors duration-300',
			passwordMismatchError && 'border-foreground-error focus:border-foreground-error'
		)}
	/>
	<Button variant="default" disabled={!username || !password || !confirmPassword || posting}>
		{#if posting}
			<Spinner class="bg-background-alt-4  size-4" />
		{:else}
			Create account
		{/if}
	</Button>
</form>
