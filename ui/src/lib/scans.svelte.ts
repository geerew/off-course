import { toast } from 'svelte-sonner';
import type { APIError } from './api-error.svelte';
import { GetScans } from './api/scan-api';
import type { ScanStatus } from './models/scan-model';

class ScanMonitor {
	#scans = $state<Record<string, ScanStatus>>({});
	#scansRo = $derived(this.#scans);
	#interval = $state<number | null>(null);

	constructor() {}

	async fetch(): Promise<void> {
		try {
			const resp = await GetScans();

			if (resp.length === 0) {
				this.#scans = {};
				this.stop();
				return;
			} else {
				const tempScans: Record<string, ScanStatus> = {};
				for (const scan of resp) {
					tempScans[scan.courseId] = scan.status;
				}

				this.#scans = tempScans;

				if (this.#interval === null) {
					this.#interval = setInterval(() => {
						this.fetch();
					}, 3000);
				}
			}
		} catch (error) {
			toast.error((error as APIError).message);
		}
	}

	stop(): void {
		if (this.#interval) {
			clearInterval(this.#interval);
			this.#interval = null;
		}
	}

	get scans() {
		return this.#scansRo;
	}
}

export const scanMonitor = new ScanMonitor();
