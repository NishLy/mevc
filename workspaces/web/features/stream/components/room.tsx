"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import { use, useEffect, useRef, useState } from "react"
import { MediaStreamController } from "../services/local"
import useMeet from "../state/meet"
import { WebRTCService } from "../services/rtc"
import WSservice from "@/lib/ws"
import { MediaStreamItem, MeetConnectionState } from "../types/service"
import ChatTabs from "./chat"
import JoiningVariant from "./loadings/join"
import ReconnectingVariant from "./loadings/rejoin"
import MeetClosedVariant from "./loadings/closed"
import { disconnect } from "node:cluster"

interface RoomProps {
  roomId: string
}

const dummyClientId =
  window.location.search.split("client_id=")[1] ||
  "client_" + Math.random().toString(36).substr(2, 9)

console.log("Generated dummy client ID:", dummyClientId)

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

  console.log("Room component rendered with status:", status)

  useEffect(() => {
    const ws = new WSservice({
      url: process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8000/ws",
      options: {
        reconnect: true, // to make sure it tries to reconnect on disconnection
        autoConnect: true,
        listeners: {
          connect: () => {
            ws.emit("join_room", dummyClientId, roomId)
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

    const localMediaController = new MediaStreamController(dummyClientId)
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
      localStreams.length === 0 ||
      status !== MeetConnectionState.SessionCreated ||
      !ws
    ) {
      return
    }

    if (RTCService) {
      console.warn(
        "WebRTC service already initialized. Skipping re-initialization."
      )
      return
    }

    const webRTCService = new WebRTCService(
      dummyClientId,
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
        onPeerStatusChanged: (peerStatus) => {
          if (peerStatus === "disconnected" || peerStatus === "failed") {
            webRTCService.destroy()
            setRTCService(null)

            console.warn(
              "Peer connection lost. Status:",
              peerStatus,
              useMeet.getState().status
            )

            if (useMeet.getState().status === MeetConnectionState.Connected) {
              setCurrentStatus(MeetConnectionState.Disconnected)
            }
          }
          if (peerStatus === "connected") {
            setCurrentStatus(MeetConnectionState.Connected)
          }

          console.log("Peer connection status changed:", peerStatus)
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
