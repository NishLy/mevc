import useSocketIo, { socketIoState } from "@/hooks/socket.io"
import { useRef, useEffect, use, useState } from "react"
import { RtcWsMessage } from "../types/stream"

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

const iceCandidateQueue: RTCIceCandidate[] = []

const bootstrapPeerConnection = async (
  chatRoomId: string,
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
      const candidateMessage: RtcWsMessage<RTCIceCandidateInit> = {
        clientId: useSocketIo.getState().socket?.id || "",
        type: "candidate",
        payload: event.candidate,
      }

      useSocketIo
        .getState()
        .socket?.emit("send_candidate", chatRoomId, candidateMessage)
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

  const offerMessage: RtcWsMessage<RTCSessionDescriptionInit> = {
    clientId: useSocketIo.getState().socket?.id || "",
    type: "offer",
    payload: offer,
  }

  useSocketIo.getState().socket?.emit("send_offer", chatRoomId, offerMessage)

  useSocketIo
    .getState()
    .socket?.on(
      "received_offer",
      async (data: RtcWsMessage<RTCSessionDescriptionInit>) => {
        console.log("Received offer message:", data)
        if (
          data.clientId === useSocketIo.getState().socket?.id ||
          data.type !== "offer"
        ) {
          return
        }

        const offer = new RTCSessionDescription(data.payload)
        await pc.setRemoteDescription(offer)
        const answer = await pc.createAnswer()
        await pc.setLocalDescription(answer)

        const answerMessage: RtcWsMessage<RTCSessionDescriptionInit> = {
          clientId: useSocketIo.getState().socket?.id || "",
          type: "answer",
          payload: answer,
        }

        useSocketIo
          .getState()
          .socket?.emit("send_answer", chatRoomId, answerMessage)
      }
    )

  useSocketIo
    .getState()
    .socket?.on(
      "received_answer",
      async (data: RtcWsMessage<RTCSessionDescriptionInit>) => {
        if (
          data.clientId === useSocketIo.getState().socket?.id ||
          data.type !== "answer"
        ) {
          return
        }

        const answer = new RTCSessionDescription(data.payload)
        await pc.setRemoteDescription(answer)

        // Now that the remote description is set, we can process any queued ICE candidates
        while (iceCandidateQueue.length > 0) {
          const candidate = iceCandidateQueue.shift()!
          await pc.addIceCandidate(candidate)
        }
      }
    )

  useSocketIo
    .getState()
    .socket?.on(
      "received_candidate",
      async (data: RtcWsMessage<RTCIceCandidateInit>) => {
        if (
          data.clientId === useSocketIo.getState().socket?.id ||
          data.type !== "candidate"
        ) {
          return
        }

        const candidate = new RTCIceCandidate(data.payload)

        if (
          peerRef.current?.remoteDescription &&
          peerRef.current?.remoteDescription.type
        ) {
          // If ready, add it immediately
          await peerRef.current?.addIceCandidate(candidate)
        } else {
          // If not ready, tuck it away for later
          iceCandidateQueue.push(candidate)
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

    if (!socket.socket || !props.roomId) {
      return
    }

    bootstrapPeerConnection(
      props.roomId,
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
  }, [props.localStreams.length, socket.socket, props.roomId])
}

export default usePeer
