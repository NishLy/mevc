import WSservice from "@/lib/ws"

interface WebRTCServiceProps {
  onRemoteStream: (stream: MediaStream) => void
}

export class WebRTCService {
  private peerConnection: RTCPeerConnection | null = null
  private wsService: WSservice | null = null
  private localStreams: MediaStream[] = []
  private roomId: string = ""

  private options: WebRTCServiceProps = {
    onRemoteStream: () => {},
  }

  constructor(
    roomId: string,
    wsService: WSservice,
    localStreams: MediaStream[],
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
      async (answer: RTCSessionDescriptionInit) => {
        if (!this.peerConnection) {
          throw new Error("Peer connection not initialized")
        }

        await this.peerConnection.setRemoteDescription(answer)
      }
    )

    this.wsService?.on(
      "ice_candidate",
      async (candidate: RTCIceCandidateInit) => {
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
    this.localStreams.forEach((stream) => {
      stream.getTracks().forEach((track) => {
        this.peerConnection?.addTrack(track, stream)
      })
    })

    // Listen for remote streams
    this.peerConnection.ontrack = (event) => {
      if (event.streams && event.streams[0]) {
        this.options?.onRemoteStream?.(event.streams[0])
      }
    }

    // handle ice candidate
    this.peerConnection.onicecandidate = (event) => {
      if (event.candidate) {
        this.emit("ice_candidate", event.candidate)
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
    this.wsService.emit(eventName, data)
  }

  destroy() {
    this.peerConnection?.close()
    this.peerConnection = null
    this.wsService?.off("receive_answer")
  }
}
