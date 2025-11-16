<script lang="ts">
	import { MediaVolumeHighIcon, MediaVolumeMuteIcon } from '$lib/components/icons';
</script>

<media-mute-button
	class="hover:text-secondary group relative inline-flex cursor-pointer items-center justify-center rounded-md ring-sky-400 outline-none ring-inset data-focus:ring-4"
>
	<MediaVolumeMuteIcon
		class="group-hover:fill-background-primary group-hover:stroke-background-primary hidden size-5 fill-white stroke-white stroke-2 group-data-[state=muted]:block sm:size-6 md:size-7"
	/>
	<MediaVolumeHighIcon
		class="group-hover:[&>path]:first:fill-background-primary group-hover:stroke-background-primary hidden size-5 stroke-white stroke-2 group-data-[state=high]:block sm:size-6 md:size-7 [&>path]:first:fill-white"
	/>
</media-mute-button>
