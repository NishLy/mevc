'use client"'

import { create } from "zustand/react"
import {
  ChatMessage,
  IUser,
  LocalStreamsTuple,
  MediaCombinedStream,
  MeetConnectionState,
  ParticipantData,
  RoomMetadata,
  RoomState,
} from "../types/service"
import { MediaStreamController } from "../services/local"
import { WebRTCService } from "../services/rtc"
import WSservice from "@/lib/ws"

const dummyClientId = Math.random().toString(36).substr(2, 5)

interface VideoPagination {
  currentPage: number
  totalPages: number
  maxiumPerPage: number
  visibleStreams: (MediaCombinedStream | null)[]
}

interface MeetState {
  clientId: string | null
  userName: string | null
  roomId: string | null
  pc: RTCPeerConnection | null
  localStreams: LocalStreamsTuple
  remoteStreams: MediaCombinedStream[]
  pinnedStreamIds: string[]
  controller?: MediaStreamController
  controllerState: {
    videoEnabled: boolean
    audioEnabled: boolean
    availableVideoDevices: MediaDeviceInfo[]
    availableAudioDevices: MediaDeviceInfo[]
    currentVideoDeviceId: string | null
    currentAudioDeviceId: string | null
    isCurrentlySharingScreen?: boolean
    isCurrentlyRecording?: boolean
  }
  uiControls: {
    isChatOpen: boolean
    isParticipantsOpen: boolean
    isSettingsOpen: boolean
  }
  status: MeetConnectionState
  RTCService: WebRTCService | null
  ws: WSservice | null
  lobbyParticipants: IUser[]
  roomState: RoomState
  participants: ParticipantData[]
  chatMessages: ChatMessage[]
  isChatAllFetched: boolean
  pagination: VideoPagination
  roomMetadata?: RoomMetadata
  setRoomState: (state: RoomState) => void
  addRemoteStream: (stream: MediaCombinedStream) => void
  removeRemoteStream: (streamId: string) => void
  setParticipantsInLobby: (participants: IUser[]) => void
  setRoomID: (roomId: string) => void
  setCurrentStatus: (status: MeetConnectionState) => void
  setWSservice: (service: WSservice | null) => void
  setRTCService: (service: WebRTCService | null) => void
  setController: (controller: MediaStreamController) => void
  setPinnedStreamIds: (ids: string[]) => void
}

const useMeet = create<MeetState>((set, state) => ({
  clientId: null,
  userName: null,
  roomId: null,
  pc: null,
  localStreams: [null, null],
  remoteStreams: [],
  pinnedStreamIds: [],
  controller: undefined,
  controllerState: {
    videoEnabled: true,
    audioEnabled: true,
    availableVideoDevices: [],
    availableAudioDevices: [],
    currentVideoDeviceId: null,
    currentAudioDeviceId: null,
    isCurrentlySharingScreen: false,
    isCurrentlyRecording: false,
  },
  uiControls: {
    isChatOpen: false,
    isParticipantsOpen: false,
    isSettingsOpen: false,
  },
  status: MeetConnectionState.New,
  RTCService: null,
  ws: null,
  lobbyParticipants: [],
  roomState: {
    maxium_per_page: 10,
    current_total_participants: 1,
    current_total_grouped_streams: 1,
  },
  participants: [],
  chatMessages: [],
  isChatAllFetched: false,
  pagination: {
    currentPage: 1,
    totalPages: 1,
    maxiumPerPage: 10,
    visibleStreams: [],
  },
  roomMetadata: undefined,
  setRoomState: (roomState) => {
    const maxiumPerPage = roomState.maxium_per_page

    const totalPages = Math.ceil(
      (roomState.current_total_grouped_streams -
        state().localStreams.filter((s) => !!s).length) /
        maxiumPerPage
    )

    set((state) => ({
      roomState,
      pagination: {
        ...state.pagination,
        totalPages,
        maxiumPerPage,
      },
    }))
  },
  removeRemoteStream(streamId) {
    const newRemoteStreams = state().remoteStreams.filter(
      (s) => s.id !== streamId
    )

    const maxiumPerPage = state().pagination.maxiumPerPage

    const totalPages = Math.ceil(
      (state().roomState.current_total_grouped_streams -
        state().localStreams.filter((s) => !!s).length) /
        maxiumPerPage
    )
    const villedStreams: (MediaCombinedStream | null)[] = Array.from(
      { length: maxiumPerPage },
      (_, i) => newRemoteStreams[i] || null
    )

    set((state) => ({
      remoteStreams: newRemoteStreams,
      pagination: {
        ...state.pagination,
        totalPages,
        visibleStreams: villedStreams,
      },
    }))
  },
  addRemoteStream: (stream: MediaCombinedStream) => {
    const newRemoteStreams = [
      ...state().remoteStreams.filter((s) => s.id !== stream.id),
      stream,
    ]

    const maxiumPerPage = state().pagination.maxiumPerPage

    const totalPages = Math.ceil(
      (state().roomState.current_total_grouped_streams -
        state().localStreams.filter((s) => !!s).length) /
        maxiumPerPage
    )

    const villedStreams: (MediaCombinedStream | null)[] = Array.from(
      { length: maxiumPerPage },
      (_, i) => newRemoteStreams[i] || null
    )

    set((state) => ({
      remoteStreams: newRemoteStreams,
      pagination: {
        ...state.pagination,
        totalPages,
        visibleStreams: villedStreams,
      },
    }))
  },
  setParticipantsInLobby: (participants: IUser[]) =>
    set({ lobbyParticipants: participants }),
  setRoomID: (roomId) => set({ roomId }),
  setCurrentStatus: (status) => set({ status }),
  setWSservice: (service: WSservice | null) => set({ ws: service }),
  setRTCService: (service: WebRTCService | null) =>
    set({ RTCService: service }),
  setPinnedStreamIds: (ids: string[]) => set({ pinnedStreamIds: ids }),
  setController: (controller: MediaStreamController) => {
    controller.onVideoToggleCallback = (enabled: boolean) =>
      set((state) => ({
        controllerState: {
          ...state.controllerState,
          videoEnabled: enabled,
        },
      }))
    controller.onAudioToggleCallback = (enabled: boolean) =>
      set((state) => ({
        controllerState: {
          ...state.controllerState,
          audioEnabled: enabled,
        },
      }))
    controller.onVideoDeviceChangeCallback = (deviceId: string) =>
      set((state) => ({
        controllerState: {
          ...state.controllerState,
          currentVideoDeviceId: deviceId,
        },
      }))
    controller.onAudioDeviceChangeCallback = (deviceId: string) =>
      set((state) => ({
        controllerState: {
          ...state.controllerState,
          currentAudioDeviceId: deviceId,
        },
      }))
    set({ controller })
    controller.onScreenShareToggleCallback = (isSharing: boolean) => {
      set((state) => ({
        controllerState: {
          ...state.controllerState,
          isCurrentlySharingScreen: isSharing,
        },
      }))
    }
    controller.onLocalStreamUpdateCallback = (streams: LocalStreamsTuple) => {
      const newStreams = [...streams] as LocalStreamsTuple
      set((state) => ({
        controllerState: {
          ...state.controllerState,
          localStreams: newStreams,
        },
        localStreams: newStreams,
      }))
    }
    controller.onDevicesUpdatedCallback = (
      availableVideoDevices: MediaDeviceInfo[],
      availableAudioDevices: MediaDeviceInfo[]
    ) => {
      set((state) => ({
        controllerState: {
          ...state.controllerState,
          availableVideoDevices,
          availableAudioDevices,
        },
      }))
    }

    controller.onRecordingToggleCallback = (isRecording: boolean) => {
      set((state) => ({
        controllerState: {
          ...state.controllerState,
          isCurrentlyRecording: isRecording,
        },
      }))
    }
  },
}))

export default useMeet
