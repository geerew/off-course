<script lang="ts">
	import { Spinner } from '$lib/components';
	import { RightArrowIcon } from '$lib/components/icons';
	import { Input, InputPassword, SubmitButton } from '$lib/components/ui';
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
	<InputPassword bind:value={password} placeholder="password" />
	<SubmitButton disabled={!username || !password || posting}>
		{#if !posting}
			Create account
		{:else}
			<Spinner class="bg-foreground-alt-3 size-4" />
		{/if}

		<RightArrowIcon
			class="relative left-0 size-5 transition-all duration-200 ease-in-out group-enabled:group-hover:left-1.5"
		/>
	</SubmitButton>
</form>
