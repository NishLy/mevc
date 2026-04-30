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
  ws_id: string
  client: ClientPeerProps
  type: "offer" | "answer" | "candidate"
  payload: T
}

export interface ClientPeerProps {
  id: string
  name: string
}
