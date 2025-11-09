import { toast } from 'svelte-sonner';
import { SvelteMap } from 'svelte/reactivity';
import type { APIError } from './api-error.svelte';
import { GetCourse } from './api/course-api';
import { subscribeToScans, type ScanUpdateEvent } from './api/scan-api';
import type { CourseModel } from './models/course-model';
import type { ScanModel, ScansModel, ScanStatus } from './models/scan-model';

class ScanMonitor {
	#scans = $state<Record<string, ScanStatus>>({});
	#scansRo = $derived(this.#scans);

	#sseClose: (() => void) | null = null;
	#isRunning = $derived(this.#sseClose !== null);

	// Course tracking
	#trackedCourses = new SvelteMap<string, CourseModel>();
	#lastSeenCourseIds = new Set<string>();

	// Scan tracking
	#trackedScans = new SvelteMap<string, ScanModel>();
	#lastSeenScanIds = new Set<string>();
	#trackedScanArrays = new Set<ScansModel>();

	#trackingCount = $derived(this.#trackedCourses.size + this.#trackedScans.size);

	constructor() {}

	// Starts tracking a course or an array of courses. If a course is already being
	// tracked, it will not be added again. If the scan monitor is not already
	// running, it will start fetching the scan status for the courses.
	//
	// When a course is no long being tracked, the course will be updated with the
	// latest data from the backend. The if course(s) passed in are managed by
	// $state, the calling component will be re-rendered with the latest course
	// information
	//
	// This function should not be called in an $effect if the course(s) are
	// managed by $state. It may cause an infinite loop
	trackCourses(courses: CourseModel | CourseModel[]) {
		const arr = Array.isArray(courses) ? courses : [courses];
		let added = false;
		for (const c of arr) {
			if (!this.#trackedCourses.has(c.id)) {
				this.#trackedCourses.set(c.id, c);
				added = true;
			}
		}
		// if we just added the very first thing to watch, start SSE connection
		if (added && !this.#isRunning) {
			this.#startSSE();
		}
	}

	// Stops a course from being tracked by removing it from the tracked courses. If
	// there are no more tracked courses, the scan monitor will stop
	untrackCourse(courseId: string): void {
		if (this.#trackedCourses.delete(courseId) && !this.#isRunning) {
			this.#stop();
		}
	}

	// Starts tracking a scan or an array of scans. If a scan is already being
	// tracked, it will not be added again. If the scan monitor is not already
	// running, it will start fetching the scan status
	//
	// This function should not be called in an $effect if the scan(s) are
	// managed by $state. It may cause an infinite loop
	trackScans(items: ScanModel | ScansModel) {
		const arr = Array.isArray(items) ? items : [items];
		let added = false;
		for (const it of arr) {
			if (!this.#trackedScans.has(it.courseId)) {
				this.#trackedScans.set(it.courseId, it);
				added = true;
			}
		}
		if (added && !this.#isRunning) {
			this.#startSSE();
		}
	}

	// Stops a scan from being tracked by removing it from the tracked scans. If
	// there are no more tracked scans, the scan monitor will stop
	untrackScan(courseId: string) {
		if (this.#trackedScans.delete(courseId) && !this.#isRunning) {
			this.#stop();
		}
	}

	// Starts tracking a scans array by adding it to the tracked scans arrays
	// and tracking all scans in the array
	trackScansArray(scansArray: ScansModel) {
		this.#trackedScanArrays.add(scansArray);
		this.trackScans(scansArray);
	}

	// Stops tracking a scans array by removing it from the tracked scans arrays
	// and untracking all scans in the array
	untrackScansArray(scansArray: ScansModel) {
		this.#trackedScanArrays.delete(scansArray);
		scansArray.forEach((scan) => {
			this.untrackScan(scan.courseId);
		});
	}

	// Clears all tracked courses/scans and stops the scan monitor
	//
	// Use in a component to cleanup on destroy
	//
	// ```svelte
	// $effect(() => {
	// 	   return () => scanMonitor.clearAll();
	// });
	// ```
	clearAll(): void {
		this.#stop();
		this.#trackedCourses.clear();
		this.#lastSeenCourseIds.clear();
		this.#trackedScans.clear();
		this.#lastSeenScanIds.clear();
		this.#trackedScanArrays.clear();
		this.#scans = {};
	}

	// Starts the SSE connection to receive real-time scan updates
	//
	// This function is private and should not be called directly
	#startSSE(): void {
		if (this.#trackingCount === 0) {
			return;
		}

		if (this.#sseClose) {
			// Already connected
			return;
		}

		this.#sseClose = subscribeToScans({
			onUpdate: (event: ScanUpdateEvent) => {
				if (event.type === 'scan_update') {
					const scan = event.data as ScanModel;
					this.#scans[scan.courseId] = scan.status;

					// Update tracked ScanModel in-place
					const sModel = this.#trackedScans.get(scan.courseId);
					if (sModel) {
						Object.assign(sModel, scan);
					}

					// Update tracked scan arrays
					this.#trackedScanArrays.forEach((scansArray) => {
						const index = scansArray.findIndex((s) => s.courseId === scan.courseId);
						if (index !== -1) {
							Object.assign(scansArray[index], scan);
						}
					});

					// Remember for courses & scans
					this.#lastSeenScanIds.add(scan.courseId);
					this.#lastSeenCourseIds.add(scan.courseId);
				} else if (event.type === 'scan_deleted') {
					const deletedId = (event.data as { id: string }).id;
					// Find courseId from tracked scans
					for (const [courseId, scan] of this.#trackedScans.entries()) {
						if (scan.id === deletedId) {
							delete this.#scans[courseId];
							this.untrackScan(courseId);
							this.#lastSeenScanIds.delete(courseId);

							// Update tracked scan arrays
							this.#trackedScanArrays.forEach((scansArray) => {
								const index = scansArray.findIndex((s) => s.courseId === courseId);
								if (index !== -1) {
									scansArray.splice(index, 1);
								}
							});

							// Check if course should be updated
							if (this.#lastSeenCourseIds.has(courseId)) {
								this.#updateCourse(courseId);
								this.#lastSeenCourseIds.delete(courseId);
							}
							break;
						}
					}
				} else if (event.type === 'error') {
					toast.error((event.data as { message: string }).message || 'Scan update error');
				}
			},
			onError: (error: Error) => {
				toast.error(error.message);
			},
			onClose: () => {
				this.#sseClose = null;
				// Auto-reconnect if there are still tracked items
				if (this.#trackingCount > 0) {
					setTimeout(() => this.#startSSE(), 1000);
				}
			}
		});
	}

	// Updates the course with the latest data from the backend
	//
	// This function is private and should not be called directly
	async #updateCourse(courseId: string): Promise<void> {
		if (!this.#trackedCourses.has(courseId)) return;

		try {
			const updatedCourse = await GetCourse(courseId);

			const originalCourse = this.#trackedCourses.get(courseId);
			if (originalCourse) Object.assign(originalCourse, updatedCourse);

			this.untrackCourse(courseId);
		} catch (error) {
			toast.error((error as APIError).message);
		}
	}

	// Stops the scan monitor and closes SSE connection
	//
	// This function is private and should not be called directly
	#stop(): void {
		if (this.#sseClose) {
			this.#sseClose();
			this.#sseClose = null;
		}
	}

	get scans() {
		return this.#scansRo;
	}
}

export const scanMonitor = new ScanMonitor();
