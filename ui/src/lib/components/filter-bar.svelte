<script lang="ts">
	import { cn } from '$lib/utils';
	import { XIcon } from './icons';
	import { Button } from './ui';

	type Props = {
		value: string;
		onApply?: () => void;
	};

	let { value = $bindable(''), onApply }: Props = $props();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	let containerEl = $state<HTMLDivElement>();
	let textareaEl = $state<HTMLTextAreaElement>();
	let overlayTextEl = $state<HTMLDivElement>();
	let dummyDivEl = $state<HTMLDivElement>();

	let textareaFocused = $state(false);
	let filterApplied = $state(false);
	let lineHeight = $state(0);
	let paddingTop = $state(0);
	let paddingBottom = $state(0);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Tokenize the input text
	function tokenize(text: string): string[] {
		return text.match(/(\S+|\s+)/g) || [];
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Build the overlay text
	function renderOverlay() {
		if (!overlayTextEl || value === undefined) return;

		overlayTextEl.innerHTML = tokenize(value)
			.map((token) => {
				let spanClass = '';

				if (token.trim() === 'AND' || token.trim() === 'OR') {
					spanClass = 'text-blue-500';
				}

				const whitespace =
					textareaFocused && textareaEl && textareaEl.rows > 1
						? 'whitespace-pre-wrap'
						: 'whitespace-pre';
				return `<span class="${spanClass} ${whitespace}">${token}</span>`;
			})
			.join('');
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// As the textarea gain/looses focus or as text is entered/deleted adjust the row count
	function adjustHeight() {
		if (!textareaEl || !containerEl || !dummyDivEl) return;

		let count = 1;

		if (textareaFocused && value) {
			dummyDivEl.textContent = textareaEl.value;
			document.body.appendChild(dummyDivEl);
			const dummyHeight = dummyDivEl.offsetHeight - paddingTop - paddingBottom;
			document.body.removeChild(dummyDivEl);

			// Calculate number of lines upto 4
			count = Math.min(Math.max(1, Math.ceil(dummyHeight / lineHeight)), 4);
		}

		textareaEl.rows = count;

		// Update container styles based on the line count and focus state.
		if (textareaFocused && count > 1) {
			containerEl.style.position = 'absolute';
			containerEl.style.zIndex = '10';
			document.body.style.overflow = 'hidden';
		} else {
			containerEl.style.cssText = '';
			document.body.style.overflow = '';
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Handle focus events
	function onFocus() {
		textareaFocused = true;
		requestAnimationFrame(() => {
			adjustHeight();
			renderOverlay();
		});
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Handle blur events
	function onBlur() {
		textareaFocused = false;
		requestAnimationFrame(() => {
			adjustHeight();
			renderOverlay();
		});
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Remove newlines from a paste string
	function onPaste(event: ClipboardEvent) {
		if (!textareaEl) return;
		event.preventDefault();

		const clipboardData = event.clipboardData;
		if (!clipboardData) return;

		const pastedText = clipboardData.getData('text').replace(/[\r\n]+/g, '');

		// Insert at cursor position
		const startPos = textareaEl.selectionStart;
		const endPos = textareaEl.selectionEnd;
		const textBefore = value.substring(0, startPos);
		const textAfter = value.substring(endPos);

		value = textBefore + pastedText + textAfter;

		setTimeout(() => {
			if (!textareaEl) return;
			textareaEl.selectionStart = textareaEl.selectionEnd = startPos + pastedText.length;
			renderOverlay();
		}, 0);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Handle input events
	function onInput() {
		requestAnimationFrame(() => {
			adjustHeight();
			renderOverlay();
		});
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Handle keydown events
	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();
			onApply?.();

			filterApplied = value !== '' ? true : false;
			return;
		}

		requestAnimationFrame(adjustHeight);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Sync scroll between the textarea and overlay
	function onScroll() {
		if (textareaEl && overlayTextEl) {
			overlayTextEl.scrollTop = textareaEl.scrollTop;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Update height when value changes and clean newlines
	$effect(() => {
		if (value === undefined) return;

		const cleanValue = value.replace(/[\r\n]+/g, '');
		if (cleanValue !== value) value = cleanValue;

		requestAnimationFrame(() => {
			adjustHeight();
			renderOverlay();
		});
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Initialize the dummy div element and several key styles like line height, padding, etc.
	$effect(() => {
		if (!textareaEl || dummyDivEl) return;

		const style = window.getComputedStyle(textareaEl);

		lineHeight = parseFloat(style.lineHeight);
		paddingTop = parseFloat(style.paddingTop);
		paddingBottom = parseFloat(style.paddingBottom);

		dummyDivEl = document.createElement('div');
		dummyDivEl.style.position = 'absolute';
		dummyDivEl.style.visibility = 'hidden';
		dummyDivEl.style.whiteSpace = 'pre-wrap';
		dummyDivEl.style.width = style.width;
		dummyDivEl.style.fontSize = style.fontSize;
		dummyDivEl.style.lineHeight = style.lineHeight;
		dummyDivEl.style.fontFamily = style.fontFamily;
		dummyDivEl.style.padding = style.padding;
	});

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Initial render
	$effect(() => {
		if (overlayTextEl) renderOverlay();
	});
</script>

<div class="relative flex flex-1 flex-row">
	<div
		bind:this={containerEl}
		class={cn(
			'border-background-alt-5 bg-background  relative flex min-h-10 w-full flex-row items-center rounded-lg border px-2',
			filterApplied
				? 'focus-within:border-background-primary-alt-2 border-background-primary-alt-2 '
				: 'focus-within:border-foreground-alt-2'
		)}
	>
		<div class="flex-1 pr-2">
			<div class="relative">
				<div
					bind:this={overlayTextEl}
					class={cn(
						'pointer-events-none absolute inset-0 overflow-hidden overflow-y-auto py-[7px]',
						textareaFocused ? 'whitespace-normal' : 'whitespace-nowrap'
					)}
					style={textareaFocused && textareaEl && textareaEl.rows > 1
						? 'white-space: pre-wrap;'
						: 'white-space: nowrap;'}
				></div>

				<textarea
					bind:this={textareaEl}
					bind:value
					class={cn(
						'caret-foreground scrollbar-hide placeholder-foreground-alt-2 flex h-auto w-full resize-none overflow-hidden overflow-y-auto py-[7px] text-transparent ring-0 transition-colors duration-200 focus:outline-none',
						textareaFocused ? 'whitespace-normal' : 'whitespace-nowrap'
					)}
					onfocus={onFocus}
					onblur={onBlur}
					oninput={onInput}
					onkeydown={onKeydown}
					onpaste={onPaste}
					onscroll={onScroll}
					rows="1"
					placeholder={filterApplied && !value ? '' : 'Filter'}
				></textarea>
			</div>
		</div>

		<Button
			class={cn(
				'bg-background-alt-4 text-foreground-alt-2 enabled:hover:text-foreground-alt-1 enabled:hover:bg-background-alt-6 size-6 p-0',
				!value && !filterApplied && 'cursor-default opacity-0'
			)}
			onclick={() => {
				value = '';
				filterApplied = false;
				onApply?.();
			}}
		>
			<XIcon class="size-4 stroke-[3]" />
		</Button>
	</div>
</div>
