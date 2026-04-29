export type CameraOpenSettings = MediaStreamConstraints["video"] & {}

export type AudioOpenSettings = MediaStreamConstraints["audio"] & {}

export enum StreamVideoEntityType {
  SELF = "self",
  PEER = "peer",
  SCREEN_SHARE = "screen_share",
}

export interface StreamVideoState {
  id: string
  stream?: MediaStream
  isLocal?: boolean
  type: StreamVideoEntityType
}
