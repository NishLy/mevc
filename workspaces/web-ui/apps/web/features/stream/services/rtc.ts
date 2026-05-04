import WSservice from "@/lib/ws"
import { MediaStreamItem } from "../types/service"
import { createBlackVideoTrack } from "./local"

interface WebRTCServiceProps {
  onAddedRemoteStream: (stream: MediaStreamItem) => void
  onRemovedRemoteStream?: (streamId: string) => void
}

interface TrackMeta {
  trackId: string
  kind: string
  clientId: string
  streamGroupId: string
  transceiverMid: string
}

interface PendingEntry {
  meta?: TrackMeta
  track?: MediaStreamTrack
}

export class WebRTCService {
  private peerConnection: RTCPeerConnection | null = null
  private wsService: WSservice | null = null
  private localStreams: Map<string, MediaStreamItem> = new Map()
  private roomId: string = ""

  // Rendezvous map: trackId → { meta?, track? }
  // Whichever side (socket or WebRTC) arrives second triggers resolution
  private pending: Map<string, PendingEntry> = new Map()
  private pendingInterval: NodeJS.Timeout | null = null
  private ownTrackIds: Set<string> = new Set()

  // streamGroupId → { clientId, tracks: Map<kind, track> }
  private resolvedStreams: Map<
    string,
    {
      clientId: string
      tracks: Map<string, MediaStreamTrack>
    }
  > = new Map()

  private options: WebRTCServiceProps = {
    onAddedRemoteStream: () => {},
    onRemovedRemoteStream: () => {},
  }

  constructor(
    private clientId: string,
    roomId: string,
    wsService: WSservice,
    localStreams: MediaStreamItem[],
    options?: WebRTCServiceProps
  ) {
    this.roomId = roomId
    this.wsService = wsService

    localStreams.forEach((stream) => {
      this.localStreams.set(stream.id, stream)
    })

    if (options) this.options = options

    this.init().catch((err) => {
      console.error("Failed to initialize WebRTC service:", err)
    })
  }

  private bindMethods() {
    this.init = this.init.bind(this)
    this.createPeerConnection = this.createPeerConnection.bind(this)
    this.tryResolve = this.tryResolve.bind(this)
    this.attachOrUpdate = this.attachOrUpdate.bind(this)
    this.setLocalStreams = this.setLocalStreams.bind(this)
    this.sendOffer = this.sendOffer.bind(this)
    this.emit = this.emit.bind(this)
    this.destroy = this.destroy.bind(this)
  }

  private async init() {
    this.bindMethods()

    this.createPeerConnection()

    // ── Socket side of rendezvous ──────────────────────────────────
    // Fired by server when a publisher's track is being forwarded to us.
    // May arrive before or after ontrack.
    this.wsService?.on("new_track", (clientId: string, meta: TrackMeta) => {
      if (this.clientId === clientId) {
        this.ownTrackIds.add(meta.trackId)
        return
      }

      const mid = meta.transceiverMid

      if (!mid) {
        console.warn(
          "Received new_track event without MID, cannot correlate:",
          meta.trackId
        )
        return
      }

      const entry = this.pending.get(mid) ?? {}
      entry.meta = meta
      this.pending.set(mid, entry)

      this.tryResolve(mid)
    })

    this.wsService?.on("peer_left", (clientId: string) => {
      console.log("Disconnecting streams for client:", clientId)
      const clientsStreamsGroupId = Array.from(this.resolvedStreams.entries())
        .filter(([_, value]) => value.clientId === clientId)
        .map(([key]) => key)

      for (const streamGroupId of clientsStreamsGroupId) {
        this.resolvedStreams.delete(streamGroupId)
        this.options.onRemovedRemoteStream?.(streamGroupId)
      }
    })

    // Server-driven renegotiation: after forwardTrack → ReplaceTrack the server
    // sends a new offer whose MSID contains the publisher's real track ID.
    // Answering causes the browser to fire ontrack with the matching ID.
    this.wsService?.on(
      "new_offer",
      async (_clientId: string, offer: RTCSessionDescriptionInit) => {
        if (!this.peerConnection) return
        await this.peerConnection.setRemoteDescription(offer)
        const answer = await this.peerConnection.createAnswer()
        await this.peerConnection.setLocalDescription(answer)
        this.emit("send_answer", answer)
      }
    )

    this.wsService?.on(
      "receive_answer",
      async (_clientId: string, answer: RTCSessionDescriptionInit) => {
        await this.peerConnection?.setRemoteDescription(answer)
      }
    )

    this.wsService?.on(
      "ice_candidate",
      async (_clientId: string, candidate: RTCIceCandidateInit) => {
        await this.peerConnection?.addIceCandidate(candidate)
      }
    )

    // Emit track metadata before sending offer so server
    // has it ready when new_track fires on subscribers
    this.localStreams.forEach((streamItem) => {
      streamItem.stream.getTracks().forEach((track) => {
        this.ownTrackIds.add(track.id)
        this.emit("track_changed", {
          trackId: track.id,
          kind: track.kind,
          streamGroupId: streamItem.id,
        })
        console.warn(
          "Emitting track_changed for track ID:",
          track.id,
          "kind:",
          track.kind
        )
      })
    })

    await this.sendOffer()
  }

  private createPeerConnection() {
    this.peerConnection = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    })

    // Add local tracks
    this.localStreams.forEach((streamItem) => {
      streamItem.stream.getTracks().forEach((track) => {
        this.peerConnection?.addTrack(track, streamItem.stream)
      })
    })

    // ── WebRTC side of rendezvous ──────────────────────────────────
    // May arrive before or after new_track socket event.
    this.peerConnection.ontrack = (event: RTCTrackEvent) => {
      const track = event.track
      const mid = event.transceiver.mid

      console.log("Received track:", track.id, "MID:", mid)
      if (this.ownTrackIds.has(track.id)) {
        return
      }

      if (!mid) {
        console.warn("Received track without MID, cannot correlate:", track.id)
        return
      }

      if (!this.pending.has(mid)) {
        this.emit("request_track_meta", {
          trackId: track.id,
          transceiverMid: mid,
        })
      }

      const entry = this.pending.get(mid) ?? {}
      entry.track = track
      this.pending.set(mid, entry)

      this.tryResolve(mid)
    }

    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        this.emit("ice_candidate", event.candidate)
      }
    }

    this.peerConnection.onconnectionstatechange = () => {
      if (this.peerConnection?.connectionState === "connected") {
        this.emit("ice_connected", this.roomId)
      }
    }
  }

  // Called from both sides — only acts when both meta + track are present
  private tryResolve(mid: string) {
    const entry = this.pending.get(mid)

    if (!entry?.meta || !entry?.track) return // wait for the other side

    this.pending.delete(mid)

    const { clientId, streamGroupId, kind } = entry.meta
    const track = entry.track

    if (!this.resolvedStreams.has(streamGroupId)) {
      this.resolvedStreams.set(streamGroupId, {
        clientId,
        tracks: new Map(),
      })
    }

    const resolved = this.resolvedStreams.get(streamGroupId)!
    resolved.tracks.set(kind, track)

    this.attachOrUpdate(streamGroupId)
  }

  private intervalId: NodeJS.Timeout | null = null
  private attachOrUpdate(streamGroupId: string) {
    const resolved = this.resolvedStreams.get(streamGroupId)!
    const videoTrack = resolved.tracks.get("video")
    const audioTrack = resolved.tracks.get("audio")

    const ms = new MediaStream()
    if (videoTrack) ms.addTrack(videoTrack)
    if (audioTrack) ms.addTrack(audioTrack)

    if (!videoTrack) {
      const blackTrack = createBlackVideoTrack()
      if (blackTrack) {
        ms.addTrack(blackTrack)
      }
    }

    this.options.onAddedRemoteStream({
      id: streamGroupId,
      stream: ms,
      type: "camera",
      isLocal: false,
    })

    const pc = this.peerConnection
    if (!pc) return

    if (this.intervalId) {
      clearInterval(this.intervalId)
    }

    pc.getReceivers().forEach((r) => {
      console.log(r.track?.kind, r.getParameters().codecs)
    })

    this.intervalId = setInterval(async () => {
      const stats = await pc.getStats()

      stats.forEach((report) => {
        if (report.type === "inbound-rtp" && report.kind === "video") {
          const codec = stats.get(report.codecId)

          console.log("Codec:", codec?.mimeType)
          console.log("PayloadType:", codec?.payloadType)

          console.log("bytesReceived:", report.bytesReceived)
          console.log("framesDecoded:", report.framesDecoded)
        }
      })
    }, 1000)
  }

  async setLocalStreams(newStreams: MediaStreamItem[]) {
    if (!this.peerConnection) throw new Error("Peer connection not initialized")

    const newStreamsIds = new Set(newStreams.map((s) => s.id))
    let hasChanged = false
    const newStreamsMap = new Map<string, MediaStreamItem>()

    for (const stream of this.localStreams.values()) {
      if (newStreamsIds.has(stream.id)) newStreamsMap.set(stream.id, stream)

      hasChanged = true
      this.emit("track_removed", stream.id)

      if (this.options.onRemovedRemoteStream) {
        this.options.onRemovedRemoteStream(stream.id)
      }
    }

    for (const stream of newStreamsMap.values()) {
      if (!this.localStreams.has(stream.id)) {
        hasChanged = true

        stream.stream.getTracks().forEach((track) => {
          this.peerConnection?.addTrack(track, stream.stream)
          this.emit("track_changed", {
            trackId: track.id,
            kind: track.kind,
            streamGroupId: stream.id,
          })
        })
      }
    }

    if (hasChanged) {
      this.sendOffer().catch((err) => {
        console.error("Failed to renegotiate after local stream change:", err)
      })
    }

    this.localStreams = newStreamsMap
  }

  private async sendOffer() {
    if (!this.peerConnection) throw new Error("Peer connection not initialized")
    const offer = await this.peerConnection.createOffer()
    await this.peerConnection.setLocalDescription(offer)
    this.emit("send_offer", offer)
  }

  private emit(eventName: string, data: any) {
    if (!this.wsService || !this.roomId) {
      throw new Error("Socket or room ID not initialized")
    }
    this.wsService.emit(eventName, this.clientId, data)
  }

  destroy() {
    this.peerConnection?.close()
    this.peerConnection = null
    this.pending.clear()
    this.resolvedStreams.clear()
    this.wsService?.off("receive_answer")
    this.wsService?.off("ice_candidate")
    this.wsService?.off("new_track")
  }
}
