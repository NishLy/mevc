"use client"

import ControlBar from "@/features/stream/components/control_bar"
import VideosGrid from "@/features/stream/components/video_grid"
import { use, useEffect } from "react"
import { MediaStreamController } from "../services/local"
import useMeet from "../state/meet"
import { WebRTCService } from "../services/rtc"
import WSservice from "@/lib/ws"
import { IUser, MeetConnectionState } from "../types/service"
import ChatTabs from "./chat"
import JoiningVariant from "./loadings/join"
import ReconnectingVariant from "./loadings/rejoin"
import MeetClosedVariant from "./loadings/closed"
import ReactionComponent from "./reaction"
import IRoom from "../types"

interface RoomProps {
  data: IRoom
}

const RenderLoading = (status: MeetConnectionState) => {
  if (
    [
      MeetConnectionState.New,
      MeetConnectionState.Checking,
      MeetConnectionState.Lobby,
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
    [
      MeetConnectionState.Completed,
      MeetConnectionState.Unknown,
      MeetConnectionState.Closed,
    ].includes(status)
  ) {
    return <MeetClosedVariant status={status} />
  }

  return null
}

export default function Room({ data }: RoomProps) {
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
    setParticipantsInLobby,
    addRemoteStream,
    removeRemoteStream,
    setRoomState,
  } = useMeet()

  useEffect(() => {
    if (!data || !clientId || !userName) return

    const ws = new WSservice({
      url: process.env.NEXT_PUBLIC_WS_URL || "ws://localhost:8000/ws",
      options: {
        reconnect: true, // to make sure it tries to reconnect on disconnection
        autoConnect: true,
        listeners: {
          connect: () => {
            ws.emit("join_room", clientId, data.code, userName)
          },
          joined_lobby: (joinedRoomId: string, participants: IUser[]) => {
            if (joinedRoomId === data.code) {
              setCurrentStatus(MeetConnectionState.Lobby)
              setParticipantsInLobby(participants)
            }
          },
          joined_room: (joinedRoomId: string) => {
            if (joinedRoomId === data.code) {
              setRoomID(data.code)
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

    const localMediaController = new MediaStreamController()
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
      data.code,
      ws,
      localStreams,
      {
        onAddedRemoteStream: (streamItem) => {
          addRemoteStream(streamItem)
        },
        onRemovedRemoteStream: (streamGroupId) => {
          removeRemoteStream(streamGroupId)
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
        onRoomStateChanged: (roomState) => {
          setRoomState(roomState)
        },
        onParticipantsDataChanged: (participants) => {
          useMeet.setState({ participants: participants })
        },
        onParticipantDataChanged(participant) {
          // Update the specific participant in the state
          useMeet.setState((state) => ({
            participants: state.participants.map((p) =>
              p.clientId === participant.clientId ? participant : p
            ),
          }))
        },
        onChatMessageReceived: (message) => {
          useMeet.setState((state) => ({
            chatMessages: [...state.chatMessages, message],
          }))
        },
        onChatHistoryReceived(messages) {
          if (!messages) return useMeet.setState({ isChatAllFetched: true })

          useMeet.setState((state) => ({
            // Prepend the new messages to the existing chatMessages
            chatMessages: [...messages, ...state.chatMessages],
          }))
        },
        onRoomMetadataChanged(metadata) {
          useMeet.setState({ roomMetadata: metadata })
        },
      }
    )

    setRTCService(webRTCService)
  }, [
    localStreams,
    ws,
    RTCService,
    setRTCService,
    setCurrentStatus,
    status,
    clientId,
    userName,
    addRemoteStream,
    removeRemoteStream,
    setRoomState,
  ])

  useEffect(() => {
    if (!RTCService || status !== MeetConnectionState.Connected) return

    RTCService.setLocalStreams(localStreams)
  }, [localStreams, RTCService, status])

  return (
    <>
      <div className="mx-auto flex h-screen w-full items-center justify-center overflow-hidden">
        {RenderLoading(status)}

        {status === MeetConnectionState.Connected && (
          <>
            <VideosGrid />
            <ChatTabs />
            <ControlBar />
            <ReactionComponent />
          </>
        )}
      </div>
    </>
  )
}
