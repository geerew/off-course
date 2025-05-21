import { toast } from 'svelte-sonner';
import type { APIError } from './api-error.svelte';
import { GetCourse } from './api/course-api';
import { GetScans } from './api/scan-api';
import type { CourseModel } from './models/course-model';
import type { ScanStatus } from './models/scan-model';

class ScanMonitor {
	#isFetching = false;
	#scans = $state<Record<string, ScanStatus>>({});
	#scansRo = $derived(this.#scans);
	#timeoutId = $state<number | null>(null);
	#trackedCourses = new Map<string, CourseModel>();
	#lastSeenInScans = new Set<string>();
	#trackedCount = $state(0);

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
	trackCourses(courses: CourseModel | CourseModel[]): void {
		const coursesToTrack = Array.isArray(courses) ? courses : [courses];
		let added = false;

		for (const c of coursesToTrack) {
			if (!this.#trackedCourses.has(c.id)) {
				this.#trackedCourses.set(c.id, c);
				added = true;
			}
		}

		this.#trackedCount = this.#trackedCourses.size;

		if (added && !this.#isFetching && this.#timeoutId === null) {
			this.#fetch();
		}
	}

	// Stops a course from being tracked by removing it from the tracked courses. If
	// there are no more tracked courses, the scan monitor will stop
	untrackCourse(courseId: string): void {
		if (this.#trackedCourses.has(courseId)) {
			this.#trackedCourses.delete(courseId);
			this.#trackedCount = this.#trackedCourses.size;

			if (this.#trackedCount === 0 && this.#timeoutId !== null) {
				this.#stop();
			}
		}
	}

	// Clears all tracked courses and stops the scan monitor
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
		this.#lastSeenInScans.clear();
		this.#trackedCount = 0;
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

		if (this.#trackedCount === 0) {
			this.#stop();
			return;
		}

		this.#isFetching = true;
		try {
			const resp = await GetScans();

			const tempScans: Record<string, ScanStatus> = {};
			const currentScanIds = new Set<string>();

			for (const scan of resp) {
				tempScans[scan.courseId] = scan.status;
				currentScanIds.add(scan.courseId);
				this.#lastSeenInScans.add(scan.courseId);
			}

			for (const courseId of this.#lastSeenInScans) {
				if (!currentScanIds.has(courseId)) {
					await this.#updateCourse(courseId);
					this.#lastSeenInScans.delete(courseId);
				}
			}

			this.#scans = tempScans;
		} catch (error) {
			toast.error((error as APIError).message);
		} finally {
			this.#isFetching = false;

			if (this.#trackedCount > 0) {
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

	get trackedCount() {
		return this.#trackedCount;
	}
}

export const scanMonitor = new ScanMonitor();
