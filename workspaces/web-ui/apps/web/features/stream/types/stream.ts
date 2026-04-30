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

export interface RtcWsMessage<T> {
  clientId: string
  type: "offer" | "answer" | "candidate"
  payload: T
}
