export type CameraOpenSettings = MediaStreamConstraints["video"] & {}

export type AudioOpenSettings = MediaStreamConstraints["audio"] & {}

export enum LOCAL_STREAM_TYPE {
  CAMERA = "camera",
  SCREEN_SHARE = "screen_share",
}

export type LocalStreamsTuple = [
  MediaCombinedStream | null,
  MediaCombinedStream | null,
]

export interface TrackMeta {
  trackId: string
  kind: string
  clientId: string
  streamGroupId: string
  transceiverMid: string
  streamId: string
  label: string
  enabled: boolean
  username: string
}

export interface PendingEntry {
  meta?: TrackMeta
  stream?: MediaStream
}

export interface MediaCombinedStream {
  id: string
  stream: MediaStream
  type: LOCAL_STREAM_TYPE
  isLocal: boolean
  metadata: {
    audio: TrackMeta | null
    video: TrackMeta | null
  }
  isAudioEnabled: boolean
  isVideoEnabled: boolean
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
  Lobby = "lobby",
  SessionCreated = "session_created",
  Connected = "connected",
  Completed = "completed",
  Disconnected = "disconnected",
  Unknown = "unknown",
  Reconnecting = "reconnecting",
}

export interface IUser {
  id: string
  username: string
}

export type RoomState = {
  maxium_per_page: number
  current_total_participants: number
  current_total_grouped_streams: number
}

export interface ParticipantData {
  clientId: string
  username: string
  role: string
  isMuted: boolean
  isVideoOff: boolean
  isRaisedHand: boolean
}

export interface ReactionData {
  type: "unicode" | "assets"
  clientId: string
  username: string
  value: string
}

// export interface ReactionManager {
//   sendReaction: (reaction: ReactionRequestData) => void
// }

export interface ReactionRequestData {
  type: "unicode" | "assets"
  value: string
}
