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
console.log("Generated dummy client ID:", dummyClientId)

export default function Room({ roomId }: RoomProps) {
  const localControllerRef = useRef<MediaStreamController | null>(null)
  const webRTCServiceRef = useRef<WebRTCService | null>(null)
  const wsocketService = useRef<WSservice | null>(null)
  const [wsConnected, setWsConnected] = useState(false)
  const [roomJoined, setRoomJoined] = useState(false)
  const { localStreams } = useMeet()

  useEffect(() => {
    wsocketService.current = new WSservice({
      url: "ws://localhost:8000/ws?tenant_id=123",
      options: {
        autoConnect: true,
        listeners: {
          connect: () => {
            wsocketService.current?.emit("join_room", dummyClientId, roomId)
            useMeet.setState({ roomId })
            setWsConnected(true)
          },
          joined_room: (joinedRoomId: string) => {
            if (joinedRoomId === roomId) {
              setRoomJoined(true)
            }
          },
        },
      },
    })

    const localMediaController = new MediaStreamController(dummyClientId)
    localControllerRef.current = localMediaController
    useMeet.getState().setController(localMediaController)

    return () => {
      localMediaController.destroy()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (localStreams.length === 0 || !wsConnected || !roomJoined) {
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
        onAddedRemoteStream: (streamItem) => {
          const newRemoteStreams = [
            ...useMeet.getState().remoteStreams,
            streamItem,
          ].reduce(
            (acc, stream) => {
              if (!stream.id) return acc

              if (acc[stream.id]) {
                acc[stream.id] = {
                  ...(acc[stream.id] as MediaStreamItem),
                  stream: stream.stream, // Update the stream reference
                }
              } else {
                acc[stream.id] = stream
              }

              return acc
            },
            {} as Record<string, MediaStreamItem>
          )

          useMeet.setState({ remoteStreams: Object.values(newRemoteStreams) })
        },
        onRemovedRemoteStream: (streamGroupId) => {
          const newRemoteStreams = useMeet
            .getState()
            .remoteStreams.filter((s) => s.id !== streamGroupId)
          useMeet.setState({ remoteStreams: newRemoteStreams })
        },
      }
    )

    webRTCServiceRef.current = webRTCService

    return () => {
      webRTCService.destroy()
    }
  }, [roomId, localStreams, roomJoined, wsConnected])

  return (
    <div>
      <VideosGrid />
      <ControlBar />
    </div>
  )
}
