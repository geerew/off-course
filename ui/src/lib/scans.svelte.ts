import { toast } from 'svelte-sonner';
import { SvelteMap } from 'svelte/reactivity';
import type { APIError } from './api-error.svelte';
import { GetCourse } from './api/course-api';
import { GetAllScans } from './api/scan-api';
import type { CourseModel } from './models/course-model';
import type { ScanModel, ScansModel, ScanStatus } from './models/scan-model';

class ScanMonitor {
	#isFetching = false;
	#scans = $state<Record<string, ScanStatus>>({});
	#scansRo = $derived(this.#scans);

	#timeoutId = $state<number | null>(null);
	#isRunning = $derived(this.#timeoutId !== null);

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
		// if we just added the very first thing to watch, fire the poller
		if (added && !this.#isFetching && !this.#isRunning) {
			this.#fetch();
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
		if (added && !this.#isFetching && !this.#isRunning) {
			this.#fetch();
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

	// Fetches the scan status for all courses and determines which ones have changed
	// If a course is no longer in the scan list, it will be updated with the latest data
	// from the backend. It is called 3 seconds after the last fetch until there are no
	// more courses to track
	//
	// This function is private and should not be called directly
	async #fetch(): Promise<void> {
		if (this.#isFetching) return;

		if (this.#trackingCount === 0) {
			this.#stop();
			return;
		}

		this.#isFetching = true;
		try {
			const allScans = await GetAllScans();
			const newStatus: Record<string, ScanStatus> = {};
			const seenIds = new Set<string>();

			for (const scan of allScans) {
				newStatus[scan.courseId] = scan.status;
				seenIds.add(scan.courseId);

				// update any tracked ScanModel in-place
				const sModel = this.#trackedScans.get(scan.courseId);
				if (sModel) Object.assign(sModel, scan);

				// remember for both courses & scans
				this.#lastSeenScanIds.add(scan.courseId);
				this.#lastSeenCourseIds.add(scan.courseId);
			}

			// prune scans that disappeared
			for (const id of Array.from(this.#lastSeenScanIds)) {
				if (!seenIds.has(id)) {
					this.untrackScan(id);
					this.#lastSeenScanIds.delete(id);
				}
			}

			// prune courses that disappeared (and update them)
			for (const id of Array.from(this.#lastSeenCourseIds)) {
				if (!seenIds.has(id)) {
					await this.#updateCourse(id);
					this.#lastSeenCourseIds.delete(id);
				}
			}

			// Update all tracked scan arrays to match what's on the backend
			this.#trackedScanArrays.forEach((scansArray) => {
				// Filter out scans that no longer exist on backend
				const stillExistingScans = scansArray.filter((scan) => seenIds.has(scan.courseId));

				// Replace the array contents while maintaining the reference
				scansArray.length = 0;
				scansArray.push(...stillExistingScans);
			});

			this.#scans = newStatus;
		} catch (e) {
			toast.error((e as APIError).message);
		} finally {
			this.#isFetching = false;

			// if any watchers remain, schedule the next round
			if (this.#trackingCount > 0) {
				this.#timeoutId = window.setTimeout(() => this.#fetch(), 3000);
			} else {
				this.#stop();
			}
		}
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

	// Stops the scan monitor and clears the timeout
	//
	// This function is private and should not be called directly
	#stop(): void {
		if (this.#timeoutId) {
			clearTimeout(this.#timeoutId);
			this.#timeoutId = null;
		}
		this.#isFetching = false;
	}

	get scans() {
		return this.#scansRo;
	}
}

export const scanMonitor = new ScanMonitor();
