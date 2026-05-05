import { create } from "zustand/react"
import { MediaStreamItem, MeetConnectionState } from "../types/service"
import { MediaStreamController } from "../services/local"
import { WebRTCService } from "../services/rtc"
import WSservice from "@/lib/ws"

interface MeetState {
  roomId: string | null
  pc: RTCPeerConnection | null
  localStreams: MediaStreamItem[]
  remoteStreams: MediaStreamItem[]
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
  status: MeetConnectionState
  RTCService: WebRTCService | null
  ws: WSservice | null
  setRoomID: (roomId: string) => void
  setCurrentStatus: (status: MeetConnectionState) => void
  setWSservice: (service: WSservice | null) => void
  setRTCService: (service: WebRTCService | null) => void
  setController: (controller: MediaStreamController) => void
  setPinnedStreamIds: (ids: string[]) => void
}

const useMeet = create<MeetState>((set) => ({
  roomId: null,
  pc: null,
  localStreams: [],
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
  status: MeetConnectionState.New,
  RTCService: null,
  ws: null,
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
    controller.onLocalStreamUpdateCallback = (streams: MediaStreamItem[]) => {
      const newStreams = [...streams]
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
