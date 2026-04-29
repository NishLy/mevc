"use client"

import ControlBar from "@/components/stream/control_bar"
import VideosGrid from "@/components/stream/video_grid"

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
