export type CameraOpenSettings = MediaStreamConstraints["video"] & {}

export type AudioOpenSettings = MediaStreamConstraints["audio"] & {}

export interface MediaStreamItem<T = any> {
  id: string
  stream: MediaStream
  streamId: string
  type: "camera" | "screen_share"
  isLocal: boolean
  metadata?: T
}

export interface MediaStreamOptions {
  cameraOptions?: CameraOpenSettings
  audioOptions?: AudioOpenSettings
  videoEnabled?: boolean
  audioEnabled?: boolean
  cameraDeviceId?: string
  audioDeviceId?: string
  useAnyAvailableCamera?: boolean
  useAnyAvailableAudio?: boolean
  autoStartStream?: boolean
}
