"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import { useEffect, useRef, useState } from "react"
import { MediaStreamController } from "../services/local"
import useMeet from "../state/meet"
import { WebRTCService } from "../services/rtc"
import WSservice from "@/lib/ws"
import { MediaStreamItem } from "../types/service"

interface RoomProps {
  roomId: string
}

const dummyClientId = crypto.randomUUID()

export default function Room({ roomId }: RoomProps) {
  const localControllerRef = useRef<MediaStreamController | null>(null)
  const webRTCServiceRef = useRef<WebRTCService | null>(null)
  const wsocketService = useRef<WSservice | null>(null)
  const [wsConnected, setWsConnected] = useState(false)
  const { localStreams } = useMeet()

  useEffect(() => {
    wsocketService.current = new WSservice({
      url: "ws://localhost:8000/ws?tenant_id=123",
      options: {
        autoConnect: true,
        listeners: {
          connect: () => {
            wsocketService.current?.emit("join_room", roomId)

            useMeet.setState({ roomId })
            setWsConnected(true)
          },
          new_track: (data: {
            clientId: string
            trackId: string
            kind: string
          }) => {
            console.log("New track received:", {
              clientId: data.clientId,
              trackId: data.trackId,
              kind: data.kind,
            })
          },
        },
      },
    })

    const localMediaController = new MediaStreamController(dummyClientId)
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
    if (localStreams.length === 0 || !wsConnected) {
      return
    }

    if (webRTCServiceRef.current) {
      console.warn(
        "WebRTC service already initialized. Skipping re-initialization."
      )
      return
    }

    const webRTCService = new WebRTCService(
      dummyClientId,
      roomId,
      wsocketService.current!,
      localStreams,
      {
        onRemoteStream: (streamItem) => {
          const newRemoteStreams = [
            ...useMeet.getState().remoteStreams,
            streamItem,
          ]
          useMeet.setState({ remoteStreams: newRemoteStreams })
        },
      }
    )

    webRTCServiceRef.current = webRTCService

    return () => {
      webRTCService.destroy()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [roomId, localStreams])

  return (
    <div>
      <VideosGrid />
      <ControlBar />
    </div>
  )
}
