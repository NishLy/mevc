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
  localController: {
    videoEnabled: boolean
    audioEnabled: boolean
    availableVideoDevices: MediaDeviceInfo[]
    availableAudioDevices: MediaDeviceInfo[]
    currentVideoDeviceId: string | null
    currentAudioDeviceId: string | null
    isCurrentlySharingScreen?: boolean
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
  localController: {
    videoEnabled: true,
    audioEnabled: true,
    availableVideoDevices: [],
    availableAudioDevices: [],
    currentVideoDeviceId: null,
    currentAudioDeviceId: null,
    isCurrentlySharingScreen: false,
  },
  setPinnedStreamIds: (ids: string[]) => set({ pinnedStreamIds: ids }),
  setController: (controller: MediaStreamController) => {
    controller.onVideoToggleCallback = (enabled: boolean) =>
      set((state) => ({
        localController: {
          ...state.localController,
          videoEnabled: enabled,
        },
      }))
    controller.onAudioToggleCallback = (enabled: boolean) =>
      set((state) => ({
        localController: {
          ...state.localController,
          audioEnabled: enabled,
        },
      }))
    controller.onVideoDeviceChangeCallback = (deviceId: string) =>
      set((state) => ({
        localController: {
          ...state.localController,
          currentVideoDeviceId: deviceId,
        },
      }))
    controller.onAudioDeviceChangeCallback = (deviceId: string) =>
      set((state) => ({
        localController: {
          ...state.localController,
          currentAudioDeviceId: deviceId,
        },
      }))
    set({ controller })
    controller.onScreenShareToggleCallback = (isSharing: boolean) => {
      set((state) => ({
        localController: {
          ...state.localController,
          isCurrentlySharingScreen: isSharing,
        },
      }))
    }
    controller.onLocalStreamUpdateCallback = (streams: MediaStreamItem[]) => {
      const newStreams = [...streams]

      console.log("Updating local streams in state", newStreams)

      set((state) => ({
        localController: {
          ...state.localController,
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
        localController: {
          ...state.localController,
          availableVideoDevices,
          availableAudioDevices,
        },
      }))
    }
  },
}))

export default useMeet
