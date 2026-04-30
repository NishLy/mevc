"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import usePeer from "@/features/stream/hooks/peer"
import useCurrentRoom from "@/features/stream/state"
import { StreamVideoEntityType } from "../types/stream"

interface RoomProps {
  roomId: string
}

export default function Room({ roomId }: RoomProps) {
  const localStreams = useCurrentRoom((state) => state.localStreams)
  const setVideoStreams = useCurrentRoom((state) => state.setVideoStreams)

  const peer = usePeer({
    roomId: roomId,
    client: {
      id: crypto.randomUUID(),
      name: Math.random() > 0.5 ? "Alice" : "Bob", // Random name for testing
    },
    localStreams: localStreams
      .map((s) => s.stream!)
      .filter((s): s is MediaStream => !!s),
    onRemoteStream: (client, stream) => {
      const newStreams = [
        {
          id: client.id,
          stream,
          type: StreamVideoEntityType.PEER,
        },
      ]
      const existingIds = new Set()
      const uniqueStreams = newStreams.filter((s) => {
        if (existingIds.has(s.id)) {
          return false
        }
        existingIds.add(s.id)
        return true
      })

      setVideoStreams(uniqueStreams)
    },
  })

  return (
    <div>
      <VideosGrid />
      <ControlBar />
    </div>
  )
}
