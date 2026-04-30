export type CameraOpenSettings = MediaStreamConstraints["video"] & {}

export type AudioOpenSettings = MediaStreamConstraints["audio"] & {}

export interface UseStreamProps {
  initateWithAnyCameraExisting?: boolean
}

export interface StreamHookReturn {
  availableCameras: MediaDeviceInfo[]
  availableAudioDevices: MediaDeviceInfo[]
  currrentState: {
    selectedCameraId: string | null
    selectedAudioDeviceId: string | null
    currentMediaStream: MediaStream | null
    isMuted: boolean
    isVideoEnabled: boolean
    currentScreenShareStream: MediaStream | null
    isCurrentlyScreenSharing: boolean
  }
  handlers: {
    handleCameraDeviceChange: (id: string, options?: CameraOpenSettings) => void
    handleAudioDeviceChange: (id: string, options?: AudioOpenSettings) => void
    toggleVideoOutput: () => void
    toggleAudioMute: () => void
    startScreenShare: () => Promise<MediaStream | null>
    stopScreenShare: (stream: MediaStream) => void
    setOptions: React.Dispatch<
      React.SetStateAction<{
        camera?: CameraOpenSettings | undefined
        audio?: AudioOpenSettings | undefined
      }>
    >
  }
}
