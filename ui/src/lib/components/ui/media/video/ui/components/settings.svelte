<script lang="ts">
	// Control visibility of panels
	let showMain = true;
	let showDetail = false;

	// Item selection
	let selectedItem: string | null = null;
	let nextItem: string | null = null;

	// List of items and their details
	const items: string[] = ['Item 1', 'Item 2', 'Item 3'];
	const itemDetails: Record<string, string[]> = {
		'Item 1': ['Detail A', 'Detail B', 'Detail C'],
		'Item 2': ['Detail D', 'Detail E'],
		'Item 3': ['Detail F', 'Detail G', 'Detail H']
	};

	// Reactive array of details for the current selected item
	let details: string[] = [];
	$: details = selectedItem ? itemDetails[selectedItem] : [];

	// Initiate transition from main to detail
	function selectItem(item: string) {
		nextItem = item;
		showMain = false;
	}

	// After main menu has fully slid out, show detail
	function handleMainOutro() {
		selectedItem = nextItem;
		nextItem = null;
		showDetail = true;
	}

	// Initiate transition from detail back to main
	function goBack() {
		showDetail = false;
	}

	// After detail panel has fully slid out, show main
	function handleDetailOutro() {
		selectedItem = null;
		showMain = true;
	}
</script>

<!-- <main style="margin-top:4rem; padding:2rem; background:white; color:black;">
	{#if showMain}
		<ul
			style="list-style:none; padding:0; max-width:200px;"
			in:fly={{ x: -200, duration: 50 }}
			out:fly={{ x: -200, duration: 50 }}
			onoutroend={handleMainOutro}
		>
			{#each items as item}
				<li
					style="cursor:pointer; margin:0.5rem 0; padding:0.5rem; border:1px solid #ddd;"
					onclick={() => selectItem(item)}
				>
					{item}
				</li>
			{/each}
		</ul>
	{:else if showDetail}
		<div
			style="position:relative; padding:1rem; border:1px solid #ddd; max-width:300px;"
			in:fly={{ x: 200, duration: 50 }}
			out:fly={{ x: 200, duration: 50 }}
			onoutroend={handleDetailOutro}
		>
			<button
				onclick={goBack}
				style="position:absolute; top:1rem; right:1rem; background:black; color:white; border:none; padding:0.5rem 1rem; cursor:pointer;"
			>
				Back
			</button>
			<h2 style="margin-top:2rem;">{selectedItem} Details</h2>
			<ul style="margin-top:1rem; padding-left:1rem;">
				{#each details as detail}
					<li style="margin:0.5rem 0;">{detail}</li>
				{/each}
			</ul>
		</div>
	{/if}
</main> -->
