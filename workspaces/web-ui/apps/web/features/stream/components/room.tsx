"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import usePeer from "@/features/stream/hooks/peer"
import useCurrentRoom from "@/features/stream/state"

interface RoomProps {
  roomId: string
}

export default function Room({ roomId }: RoomProps) {
  const room = useCurrentRoom()
  //   room.setRoomId(roomId)

  //   if (!useSocketIo.getState().initialized) {
  //     useSocketIo
  //       .getState()
  //       .initSocket()
  //       .catch((error) => {
  //         console.error("Failed to initialize Socket.IO:", error)
  //       })
  //   }

  const peer = usePeer({
    roomId: roomId,
    peerId: crypto.randomUUID(),
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
