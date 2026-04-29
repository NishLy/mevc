"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import useCurrentRoom from "@/features/stream/state"

export default function Room() {
  const room = useCurrentRoom()

  return (
    <div>
      <VideosGrid streams={room.videosStreams} />
      <ControlBar />
    </div>
  )
}
