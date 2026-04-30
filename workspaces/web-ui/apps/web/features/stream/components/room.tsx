"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import { use, useEffect, useRef } from "react"
import { MediaStreamController } from "../services/local"
import useMeet from "../state/meet"
import { WebRTCService } from "../services/rtc"
import { SocketIoService } from "@/hooks/socket.io"

interface RoomProps {
  roomId: string
}

const socketService = new SocketIoService()

export default function Room({ roomId }: RoomProps) {
  const localControllerRef = useRef<MediaStreamController | null>(null)
  const webRTCServiceRef = useRef<WebRTCService | null>(null)
  const { localStreams } = useMeet()

  useEffect(() => {
    let mounted = true

    const reinitOfferOnReconnect = async (newId: string) => {
      console.log("Reconnected with new socket ID:", newId)
      if (webRTCServiceRef.current) {
        webRTCServiceRef.current.sendOffer()
      }
    }

    const init = async () => {
      await socketService.initSocket()
      if (!mounted) return
      socketService.connectToRoom(roomId)
      socketService.onReconnect = reinitOfferOnReconnect
    }

    init()

    return () => {
      mounted = false
      socketService.leaveRoom()
      socketService.disconnect() // cleanup on unmount
    }
  }, [])

  useEffect(() => {
    const localMediaController = new MediaStreamController()
    localControllerRef.current = localMediaController
    useMeet.getState().setController(localMediaController)

    if (webRTCServiceRef.current) {
      // Re-send the offer to the new socket ID
      webRTCServiceRef.current.sendOffer()
    }

    return () => {
      localMediaController.destroy()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (localStreams.length === 0 || !socketService.isConnected) {
      return
    }

    if (webRTCServiceRef.current) {
      console.warn(
        "WebRTC service already initialized. Skipping re-initialization."
      )
      return
    }

    const webRTCService = new WebRTCService(
      roomId,
      socketService.socket!,
      localStreams.map((stream) => stream.stream),
      {
        onRemoteStream: (stream) => {
          console.log("Received remote stream:", stream)
        },
      }
    )

    webRTCServiceRef.current = webRTCService

    return () => {
      webRTCService.destroy()
    }
    // eslint-disable-next-line react-hooks/refs
  }, [roomId, localStreams])

  return (
    <div>
      <VideosGrid />
      <ControlBar />
    </div>
  )
}
