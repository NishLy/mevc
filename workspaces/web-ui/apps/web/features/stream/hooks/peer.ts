import useSocketIo, { socketIoState } from "@/hooks/socket.io"
import { useRef, useEffect, use, useState } from "react"

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
  socket: socketIoState,
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
      socket.socket?.emit("send_candidate", event.candidate)
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

  socket.socket?.emit("send_offer", offer)

  socket.socket?.on(
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

  const socket = useSocketIo()

  useEffect(() => {
    if (!props.roomId || !socket.socket) return

    // Join the room via Socket.IO
    socket.socket?.emit("join_room", props.roomId)

    return () => {
      //   ioInstance.emit("leave_room", props.roomId)
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.roomId, socket.socket?.id])

  useEffect(() => {
    if (props.localStreams.length === 0) return
    // Initialize the RTCPeerConnection and set up event listeners

    if (!socket.socket) return

    bootstrapPeerConnection(
      socket,
      peerRef,
      props.localStreams,
      props.onRemoteStream
    )

    // Clean up the peer connection on unmount
    return () => {
      peerRef.current?.close()
      peerRef.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.localStreams.length, socket.socket])
}

export default usePeer
