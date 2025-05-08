import { toast } from 'svelte-sonner';
import type { APIError } from './api-error.svelte';
import { GetCourse } from './api/course-api';
import { GetScans } from './api/scan-api';
import type { CourseModel } from './models/course-model';
import type { ScanStatus } from './models/scan-model';

// In scans.svelte.ts
class ScanMonitor {
	#scans = $state<Record<string, ScanStatus>>({});
	#scansRo = $derived(this.#scans);
	#interval = $state<number | null>(null);
	#trackedCourses = new Map<string, CourseModel>();
	#lastSeenInScans = new Set<string>(); // Track all course IDs we've seen in scans

	constructor() {}

	trackCourses(courses: CourseModel | CourseModel[]): void {
		const courseArray = Array.isArray(courses) ? courses : [courses];
		for (const course of courseArray) {
			this.#trackedCourses.set(course.id, course);
		}
	}

	async fetch(): Promise<void> {
		try {
			const resp = await GetScans();

			// Create a new scans object
			const tempScans: Record<string, ScanStatus> = {};
			const currentScanIds = new Set<string>();

			for (const scan of resp) {
				tempScans[scan.courseId] = scan.status;
				currentScanIds.add(scan.courseId);
				this.#lastSeenInScans.add(scan.courseId); // Add to our tracking set
			}

			// Find courses that were in scans before but aren't anymore
			for (const courseId of this.#lastSeenInScans) {
				if (!currentScanIds.has(courseId)) {
					// This course was in scans but isn't anymore - it completed!
					await this.#updateCourse(courseId);
					this.#lastSeenInScans.delete(courseId); // Remove from tracking
				}
			}

			this.#scans = tempScans;

			// If no more scans, clean up
			if (Object.keys(tempScans).length === 0) {
				this.#trackedCourses.clear();
				this.#lastSeenInScans.clear();
				this.stop();
				return;
			}

			if (this.#interval === null) {
				this.#interval = setInterval(() => {
					this.fetch();
				}, 3000);
			}
		} catch (error) {
			toast.error((error as APIError).message);
		}
	}

	async #updateCourse(courseId: string): Promise<void> {
		if (!this.#trackedCourses.has(courseId)) return;

		try {
			// Fetch updated course data
			const updatedCourse = await GetCourse(courseId);

			// Get the reference to the original course object
			const originalCourse = this.#trackedCourses.get(courseId);
			if (originalCourse) {
				// Update the properties of the original course
				Object.assign(originalCourse, updatedCourse);
			}

			// Remove from tracked courses
			this.#trackedCourses.delete(courseId);
		} catch (error) {
			console.error(`Failed to update course ${courseId}:`, error);
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
