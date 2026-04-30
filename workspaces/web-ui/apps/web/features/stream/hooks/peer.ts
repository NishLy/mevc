import ioInstance from "@/lib/socket.io"
import { useRef, useEffect } from "react"

interface UsePeerProps {
  roomId: string | null
  peerId: string
  localStreams: MediaStream[]
  onRemoteStream: (stream: MediaStream) => void
}

interface PeerHookReturn {
  state: {
    peerId: string
  }
  handlers: {
    // Define any handlers related to peer actions here
  }
}

const bootstrapPeerConnection = async (
  peerRef: React.MutableRefObject<RTCPeerConnection | null>,
  localStreams: MediaStream[],
  onRemoteStream?: (stream: MediaStream) => void
) => {
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }], // Always include a STUN server
  })
  peerRef.current = pc

  // 1. Set up listeners FIRST
  pc.onicecandidate = (event) => {
    if (event.candidate) {
      console.log("ICE Candidate found:", event.candidate)
      ioInstance.emit("send_candidate", event.candidate)
    }
  }

  pc.ontrack = (event) => {
    if (event.streams && event.streams[0]) {
      onRemoteStream?.(event.streams[0])
    }
  }

  localStreams.forEach((stream) => {
    stream.getTracks().forEach((track) => {
      pc.addTrack(track, stream)
    })
  })

  // 3. Now create the offer - this triggers ICE gathering
  const offer = await pc.createOffer()
  await pc.setLocalDescription(offer)

  ioInstance.emit("send_offer", offer)

  ioInstance.on(
    "received_candidate",
    async (candidate: RTCIceCandidateInit) => {
      console.log("Received ICE candidate:", candidate)
      try {
        await pc.addIceCandidate(new RTCIceCandidate(candidate))
      } catch (error) {
        console.error("Error adding received ICE candidate:", error)
      }
    }
  )
}

const usePeer = (props: UsePeerProps) => {
  const peerRef = useRef<RTCPeerConnection | null>(null)

  useEffect(() => {
    console.log("Peer hook initialized with roomId:", props.roomId)
    if (!props.roomId) return
    // Join the room via Socket.IO
    ioInstance.emit("join_room", props.roomId)
    return () => {
      //   ioInstance.emit("leave_room", props.roomId)
    }
  }, [props.roomId])

  useEffect(() => {
    if (props.localStreams.length === 0) return
    // Initialize the RTCPeerConnection and set up event listeners
    bootstrapPeerConnection(peerRef, props.localStreams, props.onRemoteStream)

    // Clean up the peer connection on unmount
    return () => {
      peerRef.current?.close()
      peerRef.current = null
    }
  }, [props.localStreams, props.onRemoteStream])
}

export default usePeer
