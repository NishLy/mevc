"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import { useEffect } from "react"
import { MediaStreamController } from "../services/local"
import useMeet from "../state/meet"
import { WebRTCService } from "../services/rtc"
import WSservice from "@/lib/ws"
import { MediaCombinedStream, MeetConnectionState } from "../types/service"
import ChatTabs from "./chat"
import JoiningVariant from "./loadings/join"
import ReconnectingVariant from "./loadings/rejoin"
import MeetClosedVariant from "./loadings/closed"

interface RoomProps {
  roomId: string
}

const RenderLoading = (status: MeetConnectionState) => {
  if (
    [
      MeetConnectionState.New,
      MeetConnectionState.Checking,
      MeetConnectionState.SessionCreated,
    ].includes(status)
  ) {
    return <JoiningVariant />
  }

  if (
    [
      MeetConnectionState.Disconnected,
      MeetConnectionState.Reconnecting,
    ].includes(status)
  ) {
    return <ReconnectingVariant />
  }

  if (
    [MeetConnectionState.Completed, MeetConnectionState.Unknown].includes(
      status
    )
  ) {
    return <MeetClosedVariant />
  }

  return null
}

export default function Room({ roomId }: RoomProps) {
  const {
    userName,
    clientId,
    localStreams,
    status,
    RTCService,
    ws,
    setCurrentStatus,
    setController,
    setWSservice,
    setRoomID,
    setRTCService,
  } = useMeet()

  console.log("Room component rendered with status:", status, localStreams)

  useEffect(() => {
    if (!roomId || !clientId || !userName) return

    const ws = new WSservice({
      url: process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8000/ws",
      options: {
        reconnect: true, // to make sure it tries to reconnect on disconnection
        autoConnect: true,
        listeners: {
          connect: () => {
            ws.emit("join_room", clientId, roomId)
          },
          joined_room: (joinedRoomId: string) => {
            if (joinedRoomId === roomId) {
              setRoomID(roomId)

              setCurrentStatus(MeetConnectionState.SessionCreated)
            }
          },
          disconnect: () => {
            setCurrentStatus(MeetConnectionState.Disconnected)
            setRTCService(null)
          },
        },
      },
    })

    const localMediaController = new MediaStreamController(clientId, userName)
    setController(localMediaController)
    setWSservice(ws)

    return () => {
      localMediaController.destroy()
      ws.close()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (
      (localStreams[0] === null && localStreams[1] === null) ||
      status !== MeetConnectionState.SessionCreated ||
      !ws ||
      !clientId ||
      !userName
    ) {
      return
    }

    if (RTCService) {
      return
    }

    const webRTCService = new WebRTCService(
      clientId,
      userName,
      roomId,
      ws,
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
                  ...(acc[stream.id] as MediaCombinedStream),
                  stream: stream.stream,
                  isVideoEnabled: stream.isVideoEnabled,
                  isAudioEnabled: stream.isAudioEnabled,
                  metadata: stream.metadata,
                }
              } else {
                acc[stream.id] = stream
              }

              return acc
            },
            {} as Record<string, MediaCombinedStream>
          )

          useMeet.setState({ remoteStreams: Object.values(newRemoteStreams) })
        },
        onRemovedRemoteStream: (streamGroupId) => {
          const newRemoteStreams = useMeet
            .getState()
            .remoteStreams.filter((s) => s.id !== streamGroupId)
          useMeet.setState({ remoteStreams: newRemoteStreams })
        },
        onPeerStatusChanged: (peerStatus) => {
          if (peerStatus === "disconnected" || peerStatus === "failed") {
            webRTCService.destroy()
            setRTCService(null)
            if (useMeet.getState().status === MeetConnectionState.Connected) {
              setCurrentStatus(MeetConnectionState.Disconnected)
            }
          }
          if (peerStatus === "connected") {
            setCurrentStatus(MeetConnectionState.Connected)
          }
        },
      }
    )

    setRTCService(webRTCService)
  }, [
    roomId,
    localStreams,
    ws,
    RTCService,
    setRTCService,
    setCurrentStatus,
    status,
    clientId,
    userName,
  ])

  useEffect(() => {
    if (!RTCService || status !== MeetConnectionState.Connected) return

    RTCService.setLocalStreams(localStreams)
  }, [localStreams, RTCService, status])

  return (
    <>
      <div className="mx-auto flex h-screen w-full items-center justify-center">
        {RenderLoading(status)}

        {status === MeetConnectionState.Connected && (
          <>
            <VideosGrid />
            <ChatTabs />
            <ControlBar />
          </>
        )}
      </div>
    </>
  )
}
