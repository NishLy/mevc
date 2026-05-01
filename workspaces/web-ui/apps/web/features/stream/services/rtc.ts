import WSservice from "@/lib/ws"
import { MediaStreamItem } from "../types/service"

interface WebRTCServiceProps {
  onRemoteStream: (stream: MediaStreamItem) => void
}

export class WebRTCService {
  private peerConnection: RTCPeerConnection | null = null
  private wsService: WSservice | null = null
  private localStreams: MediaStreamItem[] = []
  private roomId: string = ""
  private remoteStreamsMetadata: {
    trackId: string
    kind: string
    clientId: string
    streamGroupId: string
  }[] = []
  private remoteStreams: Map<string, MediaStream> = new Map()

  private options: WebRTCServiceProps = {
    onRemoteStream: () => {},
  }

  constructor(
    private clientId: string,
    roomId: string,
    wsService: WSservice,
    localStreams: MediaStreamItem[],
    options?: WebRTCServiceProps
  ) {
    this.bindAllMethods()

    this.roomId = roomId
    this.wsService = wsService
    this.localStreams = localStreams
    if (options) {
      this.options = options
    }

    this.init()
      .then(() => {
        console.log("WebRTC service initialized successfully")
      })
      .catch((error) => {
        console.error("Failed to initialize WebRTC service:", error)
      })
  }

  private bindAllMethods() {
    this.init = this.init.bind(this)
    this.createPeerConnection = this.createPeerConnection.bind(this)
    this.sendOffer = this.sendOffer.bind(this)
    this.createOffer = this.createOffer.bind(this)
    this.emit = this.emit.bind(this)
    this.destroy = this.destroy.bind(this)
  }

  private async init() {
    this.createPeerConnection()
    this.sendOffer()

    // Listen for the answer from the remote peer
    this.wsService?.on(
      "receive_answer",
      async (clientId: string, answer: RTCSessionDescriptionInit) => {
        if (!this.peerConnection) {
          throw new Error("Peer connection not initialized")
        }

        await this.peerConnection.setRemoteDescription(answer)
      }
    )

    this.wsService?.on(
      "ice_candidate",
      async (clientId: string, candidate: RTCIceCandidateInit) => {
        if (!this.peerConnection) {
          throw new Error("Peer connection not initialized")
        }
        await this.peerConnection.addIceCandidate(candidate)
      }
    )

    this.peerConnection?.addEventListener("connectionstatechange", (event) => {
      if (this.peerConnection?.connectionState === "connected") {
        console.log("Peer connection established successfully")
      }
    })

    this.wsService?.on(
      "track_changed",
      (data: {
        trackId: string
        kind: string
        streamGroupId: string
        clientId: string
      }) => {
        if (this.clientId === data.clientId) {
          return
        }

        this.remoteStreamsMetadata.push({
          trackId: data.trackId,
          kind: data.kind,
          clientId: data.clientId,
          streamGroupId: data.streamGroupId,
        })
      }
    )
  }

  async createPeerConnection() {
    this.peerConnection = new RTCPeerConnection({
      iceServers: [
        {
          urls: "stun:stun.l.google.com:19302",
        },
      ],
    })

    // Add local streams to the peer connection
    this.localStreams.forEach((streamItem) => {
      streamItem.stream.getTracks().forEach((track) => {
        this.emit("track_changed", {
          trackId: track.id,
          kind: track.kind,
          streamGroupId: streamItem.id,
        })

        this.peerConnection?.addTrack(track, streamItem.stream)
      })
    })

    // Listen for remote streams
    this.peerConnection.ontrack = (event) => {
      const track = event.track
      const streamId = event.streams[0]?.id || "default"

      let trackMetadata = this.remoteStreamsMetadata.find(
        (s) => s.trackId === track.id
      )

      if (!trackMetadata) {
        return console.warn("Received track without metadata, ignoring", {
          trackId: track.id,
        })
      }

      if (trackMetadata.clientId === this.clientId) {
        return
      }

      let remoteStream = this.remoteStreams.get(trackMetadata.streamGroupId)

      if (!remoteStream) {
        remoteStream = new MediaStream()
        this.remoteStreams.set(trackMetadata.streamGroupId, remoteStream)
      } else {
        remoteStream = new MediaStream([...remoteStream.getTracks(), track])
        this.remoteStreams.set(trackMetadata.streamGroupId, remoteStream)
      }

      const remoteStreamItem: MediaStreamItem = {
        id: trackMetadata.streamGroupId,
        stream: remoteStream,
        type: "camera",
        isLocal: false,
      }

      if (this.options.onRemoteStream) {
        this.options.onRemoteStream(remoteStreamItem)
      }
    }

    // handle ice candidate
    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        this.emit("ice_candidate", event.candidate)
      }
    }
  }

  async setLocalStreams(newStreams: MediaStreamItem[]) {
    if (!this.peerConnection) {
      throw new Error("Peer connection not initialized")
    }

    for (const newStream of newStreams) {
      for (const oldStream of this.localStreams) {
        if (oldStream.id === newStream.id) {
          continue
        }

        newStream.stream.getTracks().forEach((track) => {
          this.emit("track_changed", {
            trackId: track.id,
            kind: track.kind,
            streamGroupId: newStream.id,
          })

          this.peerConnection?.addTrack(track, newStream.stream)
        })
      }
    }
  }

  async sendOffer() {
    if (!this.peerConnection) {
      throw new Error("Peer connection not initialized")
    }
    const offer = await this.createOffer()
    this.emit("send_offer", offer)
  }

  async createOffer() {
    if (!this.peerConnection) {
      throw new Error("Peer connection not initialized")
    }

    const offer = await this.peerConnection.createOffer()
    await this.peerConnection.setLocalDescription(offer)
    return offer
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
    this.wsService?.off("receive_answer")
    this.wsService?.off("ice_candidate")
  }
}
