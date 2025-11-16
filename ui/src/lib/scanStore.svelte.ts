import { subscribeToScans, type ScanUpdateEvent } from './api/scan-api';
import type { ScanModel, ScanStatus } from './models/scan-model';

class ScanStore {
	// Scan status by courseId
	#scans = $state<Record<string, ScanStatus>>({});
	#scansRo = $derived(this.#scans);

	// All scans array
	#allScans = $state<ScanModel[]>([]);
	#allScansRo = $derived(this.#allScans);

	// Scan count
	#scanCount = $state(0);
	#scanCountRo = $derived(this.#scanCount);

	// Map scan ID to courseId (for handling deletions)
	#scanIdToCourseId = new Map<string, string>();

	// Track seen scan IDs for accurate count
	#seenScanIds = new Set<string>();

	#sseClose: (() => void) | null = null;
	#connectionCount = 0;

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Start SSE connection
	#startSSE(): void {
		if (this.#sseClose) {
			return;
		}

		this.#sseClose = subscribeToScans({
			onUpdate: (event: ScanUpdateEvent) => {
				if (event.type === 'all_scans') {
					const scans = event.data as ScanModel[];

					// Update scans by courseId
					for (const scan of scans) {
						this.#scans[scan.courseId] = scan.status;
						this.#scanIdToCourseId.set(scan.id, scan.courseId);
					}

					this.#allScans = scans;
					this.#scanCount = scans.length;
					this.#seenScanIds = new Set(scans.map((scan) => scan.id));
				} else if (event.type === 'scan_update') {
					const scan = event.data as ScanModel;

					// Update scans by courseId
					this.#scans[scan.courseId] = scan.status;
					this.#scanIdToCourseId.set(scan.id, scan.courseId);

					// Update allScans array
					const index = this.#allScans.findIndex((s) => s.id === scan.id);
					if (index !== -1) {
						// Update existing scan
						this.#allScans[index] = scan;
					} else {
						// Add new scan
						this.#allScans = [...this.#allScans, scan];

						// Update count if new scan
						if (!this.#seenScanIds.has(scan.id)) {
							this.#seenScanIds.add(scan.id);
							this.#scanCount++;
						}
					}
				} else if (event.type === 'scan_deleted') {
					const deletedId = (event.data as { id: string }).id;
					const courseId = this.#scanIdToCourseId.get(deletedId);

					if (courseId) {
						// Remove from scans by courseId
						delete this.#scans[courseId];
						this.#scanIdToCourseId.delete(deletedId);
					}

					// Remove from allScans array
					this.#allScans = this.#allScans.filter((s) => s.id !== deletedId);

					// Update count if we were tracking this scan
					if (this.#seenScanIds.has(deletedId)) {
						this.#seenScanIds.delete(deletedId);
						this.#scanCount = Math.max(0, this.#scanCount - 1);
					}
				}
			},
			onError: () => {
				this.#scans = {};
				this.#allScans = [];
				this.#scanCount = 0;
				this.#scanIdToCourseId.clear();
				this.#seenScanIds.clear();
			},
			onClose: () => {
				this.#sseClose = null;
				setTimeout(() => {
					if (this.#connectionCount > 0) {
						this.#startSSE();
					}
				}, 1000);
			}
		});
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Stop SSE connection
	#stopSSE(): void {
		if (this.#sseClose) {
			this.#sseClose();
			this.#sseClose = null;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Register that a component is using the store
	// Returns cleanup function
	register(): () => void {
		this.#connectionCount++;
		if (!this.#sseClose) {
			this.#startSSE();
		}

		return () => {
			this.#connectionCount--;

			// We could stop SSE when no components are using it?
			if (this.#connectionCount <= 0) {
				this.#connectionCount = 0;
			}
		};
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Get scan status for a course
	getScanStatus(courseId: string): ScanStatus | undefined {
		return this.#scansRo[courseId];
	}

	// Get courseId for a scan ID
	getCourseIdForScan(scanId: string): string | undefined {
		return this.#scanIdToCourseId.get(scanId);
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	get scans() {
		return this.#scansRo;
	}

	get allScans() {
		return this.#allScansRo;
	}

	get scanCount() {
		return this.#scanCountRo;
	}
}

export const scanStore = new ScanStore();
