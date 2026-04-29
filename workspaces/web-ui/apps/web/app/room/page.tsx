"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"

export default function Room() {
  return (
    <div>
      <VideosGrid
        streams={[
          { id: "localVideo", isLocal: true },
          { id: "remote-peer-1" },
          { id: "remote-peer-2" },
        ]}
      />
      <ControlBar />
    </div>
  )
}
