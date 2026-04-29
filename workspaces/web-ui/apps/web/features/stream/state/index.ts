import { create } from "zustand"
import { StreamVideoEntityType, StreamVideoState } from "../types/stream"

interface currentRoomState {
  roomId: string | null
  setRoomId: (id: string | null) => void
  videosStreams: StreamVideoState[]
  pinnedStreamIds: string[]
  setVideoStreams: (streams: StreamVideoState[]) => void
  setPinnedStreamIds: (ids: string[]) => void
}

const useCurrentRoom = create<currentRoomState>((set) => ({
  roomId: null,
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
}))

export default useCurrentRoom
