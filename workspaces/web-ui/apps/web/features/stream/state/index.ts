import { create } from "zustand"
import { StreamVideoEntityType, StreamVideoState } from "../types/stream"

interface currentRoomState {
  roomId: string | null
  setRoomId: (id: string | null) => void
  videosStreams: StreamVideoState[]
  pinnedStreamIds: string[]
  setVideoStreams: (streams: StreamVideoState[]) => void
  setPinnedStreamIds: (ids: string[]) => void
  getLocalStreams: () => MediaStream[]
}

const useCurrentRoom = create<currentRoomState>((set) => ({
  roomId: "default-room", // Default room for testing
  setRoomId: (id) => set({ roomId: id }),
  videosStreams: [
    // Placeholder for the local stream until it's initialized
    {
      id: "local-video-placeholder",
      stream: undefined,
      type: StreamVideoEntityType.SELF,
      isLocal: true,
    },
  ],
  pinnedStreamIds: [],
  setVideoStreams: (streams) => set({ videosStreams: streams }),
  setPinnedStreamIds: (ids) => set({ pinnedStreamIds: ids }),
  getLocalStreams: () => {
    const videosStreams = useCurrentRoom.getState()
      .videosStreams as StreamVideoState[]

    return videosStreams
      .filter((stream) => stream.isLocal && stream.stream)
      .map((stream) => stream.stream!)
  },
}))

export default useCurrentRoom
