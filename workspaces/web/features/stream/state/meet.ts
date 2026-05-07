import { create } from "zustand/react"
import {
  IUser,
  LocalStreamsTuple,
  MediaCombinedStream,
  MeetConnectionState,
} from "../types/service"
import { MediaStreamController } from "../services/local"
import { WebRTCService } from "../services/rtc"
import WSservice from "@/lib/ws"

const dummyClientId = "client_" + Math.random().toString(36).substr(2, 9)
const dummyUsername = "User_" + Math.random().toString(36).substr(2, 5)

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
  setParticipantsInLobby: (participants: IUser[]) => void
  setRoomID: (roomId: string) => void
  setCurrentStatus: (status: MeetConnectionState) => void
  setWSservice: (service: WSservice | null) => void
  setRTCService: (service: WebRTCService | null) => void
  setController: (controller: MediaStreamController) => void
  setPinnedStreamIds: (ids: string[]) => void
}

const useMeet = create<MeetState>((set) => ({
  clientId: dummyClientId,
  userName: dummyUsername,
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
