type VideoPlayerRef = {
	id: string;
	pause: () => void;
};

class VideoStateManager {
	private currentPlayer: VideoPlayerRef | null = null;
	private players = new Map<string, VideoPlayerRef>();

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Register a video player with the manager
	register(id: string, pauseCallback: () => void): void {
		this.players.set(id, {
			id,
			pause: pauseCallback
		});
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Unregister a video player from the manager
	unregister(id: string): void {
		this.players.delete(id);
		if (this.currentPlayer?.id === id) {
			this.currentPlayer = null;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Set the currently playing video and pause all others
	setCurrentPlayer(id: string): void {
		// Pause the previously playing video if it exists and is different
		if (this.currentPlayer && this.currentPlayer.id !== id) {
			this.currentPlayer.pause();
		}

		// Set the new current player
		const player = this.players.get(id);
		if (player) {
			this.currentPlayer = player;
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Pause all videos except the specified one
	pauseOthers(exceptId: string): void {
		for (const [id, player] of this.players) {
			if (id !== exceptId) {
				player.pause();
			}
		}
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Get the currently playing video ID
	getCurrentPlayerId(): string | null {
		return this.currentPlayer?.id || null;
	}

	// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

	// Clear the current player (useful when a video ends)
	clearCurrentPlayer(): void {
		this.currentPlayer = null;
	}
}

export const videoStateManager = new VideoStateManager();
