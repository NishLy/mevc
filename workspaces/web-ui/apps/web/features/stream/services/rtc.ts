import WSservice from "@/lib/ws"
import { MediaStreamItem } from "../types/service"

interface WebRTCServiceProps {
  onAddedRemoteStream: (stream: MediaStreamItem) => void
}

interface TrackMeta {
  trackId: string
  kind: string
  clientId: string
  streamGroupId: string
}

interface PendingEntry {
  meta?: TrackMeta
  track?: MediaStreamTrack
}

export class WebRTCService {
  private peerConnection: RTCPeerConnection | null = null
  private wsService: WSservice | null = null
  private localStreams: MediaStreamItem[] = []
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
    this.localStreams = localStreams
    if (options) this.options = options

    this.init().catch((err) => {
      console.error("Failed to initialize WebRTC service:", err)
    })
  }

  private async init() {
    this.createPeerConnection()

    // ── Socket side of rendezvous ──────────────────────────────────
    // Fired by server when a publisher's track is being forwarded to us.
    // May arrive before or after ontrack.
    this.wsService?.on("new_track", (meta: TrackMeta) => {
      if (meta.clientId === this.clientId) {
        this.ownTrackIds.add(meta.trackId)
        return
      }

      const entry = this.pending.get(meta.trackId) ?? {}
      entry.meta = meta
      this.pending.set(meta.trackId, entry)

      // No addTransceiver here — server pre-created sendrecv slots are already
      // in SDP. A server-driven renegotiation (new_offer) will update the MSID
      // so ontrack fires with the correct publisher track ID.
      this.tryResolve(meta.trackId)
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

      if (this.ownTrackIds.has(track.id)) {
        // This is our own track being looped back by the server; ignore it
        return
      }

      console.warn("Received track event for track ID:", track.id)

      const entry = this.pending.get(track.id) ?? {}
      entry.track = track
      this.pending.set(track.id, entry)

      this.tryResolve(track.id)
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
  private tryResolve(trackId: string) {
    const entry = this.pending.get(trackId)

    const pendingArr = Array.from(this.pending.entries())
    console.log(
      "Pending that have meta but no track:",
      pendingArr.filter(([_, e]) => e.meta && !e.track)
    )
    console.log(
      "Pending that have track but no meta:",
      pendingArr.filter(([_, e]) => e.track && !e.meta)
    )
    console.log(
      "Pending that have both meta and track:",
      pendingArr.filter(([_, e]) => e.track && e.meta)
    )
    if (!entry?.meta || !entry?.track) return // wait for the other side

    this.pending.delete(trackId)

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

  private attachOrUpdate(streamGroupId: string) {
    const resolved = this.resolvedStreams.get(streamGroupId)!
    const videoTrack = resolved.tracks.get("video")
    const audioTrack = resolved.tracks.get("audio")

    // Attach as soon as we have video; update again if audio arrives later
    if (!videoTrack) return

    const ms = new MediaStream()
    ms.addTrack(videoTrack)
    if (audioTrack) ms.addTrack(audioTrack)

    this.options.onAddedRemoteStream({
      id: streamGroupId,
      stream: ms,
      type: "camera",
      isLocal: false,
    })
  }

  async setLocalStreams(newStreams: MediaStreamItem[]) {
    if (!this.peerConnection) throw new Error("Peer connection not initialized")

    for (const newStream of newStreams) {
      const isAlreadyAdded = this.localStreams.some(
        (s) => s.id === newStream.id
      )
      if (isAlreadyAdded) continue

      newStream.stream.getTracks().forEach((track) => {
        this.emit("track_changed", {
          trackId: track.id,
          kind: track.kind,
          streamGroupId: newStream.id,
        })
        this.peerConnection?.addTrack(track, newStream.stream)
      })
    }

    this.localStreams = newStreams
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
