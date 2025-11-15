<!-- TODO Support dropdown for filter options and values (if supported) -->
<!-- TODO Support info button -->
<!-- TODO test how the dropdown looks when multiline -->
<script lang="ts">
	import { cn } from '$lib/utils';
	import { WarningIcon, XIcon } from './icons';
	import { Button } from './ui';

	type Props = {
		value: string;
		disabled?: boolean;
		onApply?: () => void;
		filterOptions?: Record<string, string[]>;
	};

	let { value = $bindable(''), disabled = false, onApply, filterOptions = {} }: Props = $props();

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

	let filterErrors = $state<string[]>([]);
	let showErrorDisplay = $state(false);

	let tokens = $state<string[]>([]);

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Tokenize the input text with support for quoted strings
	function tokenize(text: string): string[] {
		const _tokens = [];
		let i = 0;
		let currentToken = '';
		let inQuotes = false;
		let inWhitespace = false;

		while (i < text.length) {
			const char = text[i];

			if (char === '"') {
				const isPrecededByColon =
					currentToken.length > 0 && currentToken[currentToken.length - 1] === ':';

				if (!inQuotes) {
					if (inWhitespace) {
						_tokens.push(currentToken);
						currentToken = '';
						inWhitespace = false;
					}

					if (isPrecededByColon) {
						currentToken += char;
					} else {
						if (currentToken) {
							_tokens.push(currentToken);
							currentToken = '';
						}

						currentToken = char;
					}

					inQuotes = true;
				} else {
					currentToken += char;

					if (!currentToken.includes(':')) {
						_tokens.push(currentToken);
						currentToken = '';
					}

					inQuotes = false;
				}
			} else if (char === ' ' && !inQuotes) {
				if (inWhitespace) {
					currentToken += char;
				} else {
					if (currentToken) {
						_tokens.push(currentToken);
					}

					currentToken = ' ';
					inWhitespace = true;
				}
			} else {
				if (inWhitespace) {
					_tokens.push(currentToken);
					currentToken = '';
					inWhitespace = false;
				}

				currentToken += char;
			}

			i++;
		}

		if (currentToken) {
			_tokens.push(currentToken);
		}

		tokens = _tokens;
		return _tokens;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Build the overlay text
	function renderOverlay() {
		if (!overlayTextEl || value === undefined) return;

		overlayTextEl.innerHTML = tokenize(value)
			.map((token) => {
				let whitespaceStyle =
					textareaFocused && textareaEl && textareaEl.rows > 1
						? 'whitespace-pre-wrap'
						: 'whitespace-pre';
				let finalToken = token;

				// AND/OR
				if (token.trim() === 'AND' || token.trim() === 'OR') {
					return `<span class="text-blue-600 ${whitespaceStyle}">${finalToken}</span>`;
				}

				// Whitespace
				if (token.trim() === '') {
					return `<span class="${whitespaceStyle}">${finalToken}</span>`;
				}

				// Filter keys
				const keyMatch = token.match(/^(\w+):/);
				if (keyMatch && filterOptions && filterOptions[keyMatch[1]]) {
					const key = keyMatch[0];
					const rest = token.slice(key.length);
					return `<span class="text-amber-600 ${whitespaceStyle}">${key}</span><span class="${whitespaceStyle}">${rest}</span>`;
				}

				// Everything else
				return `<span class="${whitespaceStyle}">${finalToken}</span>`;
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

	// Validate filter and collect all errors
	function validateFilter(): string[] {
		const errors: string[] = [];

		// Check for balanced quotes
		if (!checkQuoteBalance(value)) {
			errors.push('Unbalanced quotation marks');
		}

		// Check values against filter options
		tokens.forEach((token) => {
			const keyMatch = token.match(/^(\w+):/);
			if (keyMatch && filterOptions && filterOptions[keyMatch[1]]) {
				const key = keyMatch[1];
				const value = token.slice(keyMatch[0].length);

				if (filterOptions[key].length > 0) {
					if (value.trim() === '') {
						errors.push(
							`Missing value for <span class="text-amber-600 font-semibold">${key}</span>`
						);
						return;
					}

					const options = filterOptions[key].map((option) => option.toLowerCase());
					let restLower = value.toLowerCase();

					// Ignore quotes around the value
					if (restLower.startsWith('"') && restLower.endsWith('"')) {
						restLower = restLower.slice(1, -1);
					}

					if (!options.includes(restLower)) {
						errors.push(
							`Invalid value <span class="text-foreground-error">${value}</span> for <span class="text-amber-600 font-semibold">${key}</span>`
						);
					}
				}
			}
		});

		return errors;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Check if quotes are balanced
	function checkQuoteBalance(text: string): boolean {
		let count = 0;
		for (let i = 0; i < text.length; i++) {
			if (text[i] === '"') count++;
		}
		return count % 2 === 0;
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

	// Handle enter key
	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			e.preventDefault();

			// Validate filter and collect all errors
			filterErrors = validateFilter();
			showErrorDisplay = filterErrors.length > 0;

			if (!showErrorDisplay) {
				onApply?.();
				filterApplied = value !== '' ? true : false;
			}
			return;
		}

		// Clear error display when typing
		if (showErrorDisplay) {
			showErrorDisplay = false;
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

<div class="relative flex flex-1 flex-col">
	<div
		bind:this={containerEl}
		class={cn(
			'border-background-alt-5 bg-background relative flex min-h-10 w-full flex-row items-start rounded-lg border px-2',
			filterApplied
				? 'focus-within:border-background-primary-alt-2 border-background-primary-alt-2'
				: 'focus-within:border-foreground-alt-3'
		)}
	>
		<div class={cn('flex-1 pr-2', disabled && 'pr-0')}>
			<div class="relative">
				<div
					bind:this={overlayTextEl}
					class={cn(
						'pointer-events-none absolute inset-0 overflow-hidden overflow-y-auto py-[7px]',
						textareaFocused ? 'whitespace-normal' : 'whitespace-nowrap',
						textareaFocused && textareaEl && textareaEl.rows > 1 && 'whitespace-pre-wrap'
					)}
				></div>

				<textarea
					bind:this={textareaEl}
					bind:value
					{disabled}
					class={cn(
						'caret-foreground scrollbar-hide placeholder-foreground-alt-3 flex h-auto w-full resize-none overflow-hidden overflow-y-auto py-[7px] text-transparent ring-0 transition-colors duration-200 focus:outline-none',
						textareaFocused ? 'whitespace-normal' : 'whitespace-nowrap',
						disabled && 'cursor-not-allowed opacity-50'
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

		{#if !disabled}
			<Button
				variant="ghost"
				class={cn(
					'text-foreground-alt-3 hover:text-foreground-alt-1 mt-[7px] size-6 p-0',
					!value && !filterApplied && 'cursor-default opacity-0'
				)}
				onclick={() => {
					value = '';
					filterApplied = false;
					showErrorDisplay = false;
					onApply?.();
				}}
			>
				<XIcon class="stroke-3 size-4" />
			</Button>
		{/if}
	</div>

	{#if showErrorDisplay}
		<div
			class="text-foreground-alt-1 bg-background-error/20 mt-2 flex flex-row gap-2.5 rounded-md border border-red-800/50 p-2 text-sm"
		>
			<WarningIcon class="text-foreground-error size-6 stroke-[1.5]" />
			<div class="pt-0.5">
				Filter contains {filterErrors.length}
				{filterErrors.length === 1 ? 'issue' : 'issues'}:
				<ul class="ml-4 mt-1">
					{#each filterErrors as error}
						<li class="list-disc">
							{@html error}
						</li>
					{/each}
				</ul>
			</div>
		</div>
	{/if}
</div>
