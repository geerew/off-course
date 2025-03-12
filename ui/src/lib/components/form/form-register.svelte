<script lang="ts">
	import { Spinner } from '$lib/components';
	import { Button, Input, PasswordInput } from '$lib/components/ui';
	import { toast } from 'svelte-sonner';

	let { endpoint }: { endpoint: string } = $props();

	let username = $state('');
	let password = $state('');
	let posting = $state(false);

	async function submitForm(event: Event) {
		event.preventDefault();
		posting = true;

		const response = await fetch(endpoint, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				username,
				password
			})
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
	<PasswordInput bind:value={password} placeholder="password" />
	<Button disabled={!username || !password || posting}>
		{#if posting}
			<Spinner class="bg-background-alt-4  size-4" />
		{:else}
			Create account
		{/if}
	</Button>
</form>
