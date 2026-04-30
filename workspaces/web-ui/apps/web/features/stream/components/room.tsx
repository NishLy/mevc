"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import { useEffect } from "react"
import { MediaStreamController } from "../services/local"
import useMeet from "../state/meet"

interface RoomProps {
  roomId: string
}

export default function Room({ roomId }: RoomProps) {
  useEffect(() => {
    const localMediaController = new MediaStreamController()

    useMeet.getState().setController(localMediaController)

    return () => {
      localMediaController.destroy()
    }
  }, [])

  return (
    <div>
      <VideosGrid />
      <ControlBar />
    </div>
  )
}
