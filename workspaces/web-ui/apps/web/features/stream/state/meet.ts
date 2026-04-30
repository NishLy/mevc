import { create } from "zustand/react"
import { MediaStreamItem } from "../types/service"
import { MediaStreamController } from "../services/local"

interface MeetState {
  roomId: string | null
  pc: RTCPeerConnection | null
  localStreams: MediaStreamItem[]
  remoteStreams: MediaStreamItem[]
  pinnedStreamIds: string[]
  controller: MediaStreamController | null
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
  setController: (controller: MediaStreamController) => void
  setPinnedStreamIds: (ids: string[]) => void
}

const useMeet = create<MeetState>((set) => ({
  roomId: null,
  pc: null,
  localStreams: [],
  remoteStreams: [],
  pinnedStreamIds: [],
  controller: null,
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
