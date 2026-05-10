import WSservice from "@/lib/ws"
import {
  LOCAL_STREAM_TYPE,
  LocalStreamsTuple,
  MediaCombinedStream,
  ParticipantData,
  PendingEntry,
  ReactionData,
  ReactionRequestData,
  RoomState,
  TrackMeta,
} from "../types/service"
import { createBlackVideoTrack } from "./local"
import { metadata, th, track } from "framer-motion/m"

interface WebRTCServiceProps {
  onAddedRemoteStream?: (stream: MediaCombinedStream) => void
  onRemovedRemoteStream?: (streamId: string) => void
  onPeerStatusChanged?: (status: string) => void
  onRoomStateChanged?: (state: RoomState) => void
  onParticipantDataChanged?: (participants: ParticipantData[]) => void
  onReactionReceived?: (reaction: ReactionData) => void
}

export class WebRTCService {
  private peerConnection: RTCPeerConnection | null = null
  private wsService: WSservice | null = null
  private localStreams: LocalStreamsTuple = [null, null]

  // Rendezvous map: trackId → { meta?, track? }
  // Whichever side (socket or WebRTC) arrives second triggers resolution
  private pending: Map<string, PendingEntry> = new Map()
  private generatedStreamIds: Map<
    string,
    { audio: string | null; video: string | null }
  > = new Map() // client-generated streamId → generated streamId for correlating tracks that arrive before the stream is fully ready

  // streamId → { clientId, meta, streams: Map<kind, MediaStream> }
  private resolvedStreams: Map<
    string,
    {
      clientId: string
      metas: {
        audio: TrackMeta | null
        video: TrackMeta | null
      }
      streams: Map<string, MediaStream>
      localStream: MediaCombinedStream | null
    }
  > = new Map()

  private options: WebRTCServiceProps = {
    onAddedRemoteStream: () => {},
    onRemovedRemoteStream: () => {},
    onPeerStatusChanged: (status: string) => {},
    onRoomStateChanged: (state: RoomState) => {},
    onReactionReceived: (reaction: ReactionData) => {},
  }

  constructor(
    private clientId: string,
    private username: string,
    private roomId: string,
    wsService: WSservice,
    localStreams: LocalStreamsTuple,
    options?: WebRTCServiceProps
  ) {
    this.wsService = wsService

    localStreams.forEach((stream, index) => {
      this.localStreams[index] = stream
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
    await this.createPeerConnection()
    this.attachWSListeners()

    // Ensure any initial local streams are added to the peer connection and the other side is aware of them
    await Promise.all(
      this.localStreams.map(async (streamItem) => {
        if (!streamItem) return
        await this.addStreamToPeerConnection(streamItem)
      })
    )

    await this.sendOffer()
  }

  /**
   * Splits a MediaStream into separate video-only and audio-only streams to work around limitations in some browsers that prevent adding tracks from the same stream more than once, allowing for better track management and renegotiation in the peer connection
   * @param sourceStream - The source MediaStream to split
   * @returns Object containing videoOnlyStream and audioOnlyStream
   */
  private splitMediaStream(sourceStream: MediaStream): {
    videoOnlyStream: MediaStream | null
    audioOnlyStream: MediaStream | null
  } {
    let videoOnlyStream: MediaStream | null = null
    let audioOnlyStream: MediaStream | null = null

    sourceStream.getTracks().forEach((track) => {
      if (track.kind === "video") {
        videoOnlyStream = new MediaStream()
        videoOnlyStream.addTrack(track)
      } else if (track.kind === "audio") {
        audioOnlyStream = new MediaStream()
        audioOnlyStream.addTrack(track)
      }
    })

    return { videoOnlyStream, audioOnlyStream }
  }

  private generateStreamGroupId() {
    return crypto.randomUUID()
  }

  private async addStreamToPeerConnection(streamItem: MediaCombinedStream) {
    const { videoOnlyStream, audioOnlyStream } = this.splitMediaStream(
      streamItem.stream
    )

    const streamGroupId =
      streamItem.metadata.audio?.streamGroupId ??
      streamItem.metadata.video?.streamGroupId ??
      this.generateStreamGroupId()

    videoOnlyStream?.getTracks().forEach((track) => {
      this.peerConnection?.addTrack(track, videoOnlyStream)

      const meta: TrackMeta = {
        trackId: track.id,
        kind: track.kind,
        clientId: this.clientId,
        streamGroupId,
        transceiverMid: "",
        streamId: videoOnlyStream.id,
        label: track.label,
        enabled: track.enabled,
        username: this.username,
      }

      this.emit("track_changed", meta)
    })

    audioOnlyStream?.getTracks().forEach((track) => {
      this.peerConnection?.addTrack(track, audioOnlyStream)

      const meta: TrackMeta = {
        trackId: track.id,
        kind: track.kind,
        clientId: this.clientId,
        streamGroupId,
        transceiverMid: "",
        streamId: audioOnlyStream.id,
        label: track.label,
        enabled: track.enabled,
        username: this.username,
      }

      this.emit("track_changed", meta)
    })

    // Store the generated stream group ID to correlate with incoming tracks that arrive before the stream is fully ready
    this.generatedStreamIds.set(streamGroupId, {
      audio: audioOnlyStream?.id || null,
      video: videoOnlyStream?.id || null,
    })
  }

  private async removeStreamFromPeerConnection(stream: MediaStream) {
    if (!this.peerConnection) return

    const senders = this.peerConnection.getSenders()
    stream.getTracks().forEach((track) => {
      const sender = senders.find((s) => s.track === track)
      if (sender) {
        this.peerConnection!.removeTrack(sender)
      }
    })
  }

  private async createPeerConnection() {
    this.peerConnection = new RTCPeerConnection({
      iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
    })

    // May arrive before or after new_track socket event.
    this.peerConnection.ontrack = (event: RTCTrackEvent) => {
      const track = event.track
      const streamId = event.streams[0]?.id ?? ""

      console.log(
        "Received new track event from peer with track ID:",
        track.id,
        "and stream ID:",
        streamId
      )
      if (!streamId) {
        console.warn(
          "Received track without stream ID, cannot correlate:",
          track.id
        )
        return
      }

      const entry = this.pending.get(streamId) ?? {}
      entry.stream = event.streams[0]
      this.pending.set(streamId, entry)

      this.tryResolve(streamId)
    }

    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        this.emit("ice_candidate", event.candidate)
      }
    }

    this.peerConnection.oniceconnectionstatechange = (event) => {
      switch (this.peerConnection?.iceConnectionState) {
        case "connected":
          setTimeout(() => {
            this.emit("peer_status_changed", "connected")
          }, 1000) // add a slight delay to ensure the connection is fully established before notifying

          this.emit("participant_data_request")
          break
        case "completed":
        case "disconnected":
          console.warn(
            "Peer connection state changed to disconnected/completed, treating as disconnected:",
            this.peerConnection.iceConnectionState,
            "Reason",
            event
          )
          break
        case "failed":
        default:
      }

      this.options.onPeerStatusChanged?.(
        this.peerConnection?.iceConnectionState ?? "unknown"
      )
    }
  }

  // Called from both sides — only acts when both meta + track are present
  private tryResolve(streamId: string) {
    const entry = this.pending.get(streamId)

    if (!entry?.meta || !entry?.stream) return // wait for the other side
    this.pending.delete(streamId)

    const { clientId, streamGroupId, kind } = entry.meta
    const stream = entry.stream

    if (!clientId || !streamGroupId || !kind) {
      console.warn(
        "Received incomplete track information, cannot resolve:",
        entry.meta
      )
      return
    }

    if (!this.resolvedStreams.has(streamGroupId)) {
      this.resolvedStreams.set(streamGroupId, {
        clientId,
        streams: new Map(),
        metas: {
          audio: kind === "audio" ? entry.meta : null,
          video: kind === "video" ? entry.meta : null,
        },
        localStream: null,
      })
    }

    const resolved = this.resolvedStreams.get(streamGroupId)!
    resolved.streams.set(kind, stream)

    this.attachOrUpdate(streamGroupId)
  }

  private attachOrUpdate(streamGroupId: string) {
    const resolved = this.resolvedStreams.get(streamGroupId)!
    const videoStream = resolved.streams.get("video")
    const audioStream = resolved.streams.get("audio")

    const videoTrack = videoStream?.getVideoTracks()[0]
    const audioTrack = audioStream?.getAudioTracks()[0]

    const ms = new MediaStream()
    if (videoTrack) ms.addTrack(videoTrack)
    if (audioTrack) ms.addTrack(audioTrack)

    // fallback to a black video track if no video is provided to ensure the stream is rendered on the other side and can be used for renegotiation when a real video track is added later (e.g. user enables camera after joining)
    if (!videoTrack) {
      const blackTrack = createBlackVideoTrack(1280, 720)
      if (blackTrack) {
        ms.addTrack(blackTrack)
      }
    }

    resolved.localStream = {
      id: streamGroupId,
      stream: ms,
      type: LOCAL_STREAM_TYPE.CAMERA,
      isLocal: false,
      metadata: resolved.metas,
      isAudioEnabled: !!audioTrack,
      isVideoEnabled: !!videoTrack,
    }

    this.options.onAddedRemoteStream?.({
      id: streamGroupId,
      stream: ms,
      type: LOCAL_STREAM_TYPE.CAMERA,
      isLocal: false,
      metadata: resolved.metas,
      isAudioEnabled: !!audioTrack,
      isVideoEnabled: !!videoTrack,
    })
  }

  // Updates the local media stream in the list of local streams, either replacing the existing stream or adding a new entry if it doesn't exist, and triggers an update callback to notify of changes
  async setLocalStreams(newStreams: LocalStreamsTuple) {
    if (!this.peerConnection) throw new Error("Peer connection not initialized")

    let isModified = false
    const initialLocalStreams: LocalStreamsTuple = [null, null]

    // Remove old streams that are not in the new set
    for (let i = 0; i < newStreams.length; i++) {
      const stream = newStreams[i]
      if (!stream) continue

      const isExist = this.localStreams.some((s) => s?.id === stream.id)

      if (isExist) {
        initialLocalStreams[i] = stream
        continue
      }

      const streamID = stream.stream.id
      this.emit("track_removed", streamID)
      this.options.onRemovedRemoteStream?.(streamID)
      isModified = true
    }

    // Add new streams (or update existing ones)
    Promise.all(
      initialLocalStreams
        .filter((stream) => !!stream)
        .map(async (stream) => {
          const existingStream = this.localStreams.find(
            (s) => s?.id === stream?.id
          )

          if (existingStream) {
            for (const meta in existingStream.metadata) {
              const value =
                existingStream.metadata[
                  meta as keyof typeof existingStream.metadata
                ]

              if (!value) continue

              const generatedIds = this.generatedStreamIds.get(
                value.streamGroupId
              )

              const alteredMeta: TrackMeta = {
                ...value,
                streamId: generatedIds
                  ? meta === "video"
                    ? generatedIds.video || value.streamId
                    : generatedIds.audio || value.streamId
                  : value.streamId,
                enabled:
                  meta === "video"
                    ? stream!.isVideoEnabled
                    : stream!.isAudioEnabled,
              }

              this.emit("track_changed", alteredMeta)
            }

            return existingStream
          }

          await this.addStreamToPeerConnection(stream)
          isModified = true
          return stream!
        })
    ).catch((err) => {
      console.error("Failed to add local stream to peer connection:", err)
    })

    if (isModified) {
      this.sendOffer().catch((err) => {
        console.error("Failed to renegotiate after local stream change:", err)
      })
    }

    this.localStreams = initialLocalStreams
  }

  // send a renegotiation offer to the other peer whenever local stream changes (e.g. user toggles camera/mic or starts/stops screen share) so the new tracks are added/removed on the other side and the connection is kept up to date with the current state of local media
  private async sendOffer() {
    if (!this.peerConnection) throw new Error("Peer connection not initialized")
    const offer = await this.peerConnection.createOffer()
    await this.peerConnection.setLocalDescription(offer)
    this.emit("send_offer", offer)
  }

  // Simple wrapper around emit to ensure clientId and roomId are always included and to provide better type safety
  private emit(eventName: string, ...data: unknown[]) {
    if (!this.wsService || !this.roomId) {
      throw new Error("Socket or room ID not initialized")
    }
    this.wsService.emit(eventName, this.clientId, ...data)
  }

  private attachWSListeners() {
    this.wsService?.on("new_track", (clientId: string, meta: TrackMeta) => {
      console.log("Received new_track event with meta:", meta)
      if (this.clientId === clientId) {
        return
      }

      const streamId = meta.streamId

      if (!streamId) {
        console.warn(
          "Received new_track event without stream ID, cannot correlate:",
          meta.trackId
        )
        return
      }

      const entry = this.pending.get(streamId) ?? {}
      entry.meta = meta
      this.pending.set(streamId, entry)

      this.tryResolve(streamId)
    })

    this.wsService?.on("peer_left", (clientId: string) => {
      const clientsStreamsGroupId = Array.from(this.resolvedStreams.entries())
        .filter(([, value]) => value.clientId === clientId)
        .map(([key]) => key)

      for (const streamGroupId of clientsStreamsGroupId) {
        const resolved = this.resolvedStreams.get(streamGroupId)

        if (resolved) {
          resolved.streams.forEach((stream) => {
            this.removeStreamFromPeerConnection(stream).catch((err) => {
              console.error(
                "Failed to remove stream from peer connection:",
                err
              )
            })
          })
        }

        this.resolvedStreams.delete(streamGroupId)
        this.options.onRemovedRemoteStream?.(streamGroupId)
      }

      // this.sendOffer().catch((err) => {
      //   console.error("Failed to renegotiate after peer left:", err)
      // })
    })

    this.wsService?.on(
      "new_offer",
      async (_clientId: string, offer: RTCSessionDescriptionInit) => {
        if (!this.peerConnection) return

        //
        offer.sdp = offer.sdp?.replace(
          /rtpmap:0 opus\/48000/g,
          "rtpmap:111 opus/48000"
        )

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

    this.wsService?.on("track_removed", (streamGroupId: string) => {
      const resolved = this.resolvedStreams.get(streamGroupId)

      if (resolved) {
        resolved.streams.forEach((stream) => {
          this.removeStreamFromPeerConnection(stream).catch((err) => {
            console.error("Failed to remove stream from peer connection:", err)
          })
        })

        this.resolvedStreams.delete(streamGroupId)
        this.options.onRemovedRemoteStream?.(streamGroupId)
      }
    })

    // listen to remote track
    this.wsService?.on("track_changed", (clientId: string, meta: TrackMeta) => {
      if (meta.clientId === this.clientId) {
        return
      }

      if (!meta.streamGroupId || !meta.kind) {
        console.warn(
          "Received track_changed event with incomplete meta, cannot correlate:",
          meta
        )
        return
      }

      const resolved = this.resolvedStreams.get(meta.streamGroupId)

      if (!resolved || !resolved.streams) return
      const stream = resolved.streams.get(meta.kind)
      if (!stream) return

      const kind = meta.kind

      let localStream = resolved.localStream
      if (!localStream) return

      localStream = {
        ...localStream,
        isAudioEnabled:
          kind === "audio" ? meta.enabled : localStream.isAudioEnabled,
        isVideoEnabled:
          kind === "video" ? meta.enabled : localStream.isVideoEnabled,
      }
      resolved.localStream = localStream

      this.options.onAddedRemoteStream?.(resolved.localStream)
    })

    this.wsService?.on("room_state_changed", (state: RoomState) => {
      this.options.onRoomStateChanged?.(state)
      // Trigger a participant data refresh whenever room state changes to ensure participant list is up to date (e.g. when someone joins/leaves or toggles their media)
      this.emit("participant_data_request")
    })

    this.wsService?.on(
      "participants_data_response",
      (participants: ParticipantData[]) => {
        this.options.onParticipantDataChanged?.(participants)
      }
    )

    this.wsService?.on(
      "reaction_received",
      (_: string, reaction: ReactionData) => {
        this.options.onReactionReceived?.(reaction)
      }
    )
  }

  requestPageChange(page: number) {
    this.emit("page_change_request", page)
  }

  requestReaction(type: "unicode" | "assets", value: string) {
    this.emit("reaction_sent", { type, value } as ReactionRequestData)
  }

  onReactionReceived(callback: (reaction: ReactionData) => void) {
    this.options.onReactionReceived = callback
  }

  destroy() {
    this.peerConnection?.close()
    this.peerConnection = null
    this.pending.clear()
    this.resolvedStreams.clear()
    this.wsService?.off("receive_answer")
    this.wsService?.off("ice_candidate")
    this.wsService?.off("new_track")
    this.wsService?.off("track_removed")
    this.wsService?.off("room_state_changed")
    this.wsService?.off("track_changed")
    this.wsService?.off("peer_left")
    this.wsService?.off("page_change_request")
    this.wsService?.off("reaction_received")
    this.wsService = null
  }
}
