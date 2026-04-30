"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import usePeer from "@/features/stream/hooks/peer"
import useCurrentRoom from "@/features/stream/state"
import { useEffect, useState } from "react"

export default function Room() {
  const room = useCurrentRoom()

  const peer = usePeer({
    roomId: room.roomId,
    peerId: room.peerId,
    localStreams: room.getLocalStreams(),
    onRemoteStream: (stream) => {
      console.log("Received remote stream:", stream)
    },
  })

  return (
    <div>
      <VideosGrid streams={room.videosStreams} />
      <ControlBar />
    </div>
  )
}
