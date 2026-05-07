export type CameraOpenSettings = MediaStreamConstraints["video"] & {}

export type AudioOpenSettings = MediaStreamConstraints["audio"] & {}

export enum LOCAL_STREAM_TYPE {
  CAMERA = "camera",
  SCREEN_SHARE = "screen_share",
}

export type LocalStreamsTuple = [MediaStreamItem | null, MediaStreamItem | null]

export interface MediaStreamItem<T = Record<string, unknown>> {
  id: string
  stream: MediaStream
  type: LOCAL_STREAM_TYPE
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

export enum MeetConnectionState {
  New = "new",
  Checking = "checking",
  SessionCreated = "session_created",
  Connected = "connected",
  Completed = "completed",
  Disconnected = "disconnected",
  Unknown = "unknown",
  Reconnecting = "reconnecting",
}
