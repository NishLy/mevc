import { create } from "zustand"
import { StreamVideoEntityType, StreamVideoState } from "../types/stream"

interface currentRoomState {
  roomId: string | null
  setRoomId: (id: string | null) => void
  videosStreams: StreamVideoState[]
  localStreams: StreamVideoState[]
  pinnedStreamIds: string[]
  setVideoStreams: (streams: StreamVideoState[]) => void
  setPinnedStreamIds: (ids: string[]) => void
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
  localStreams: [],
  pinnedStreamIds: [],
  setVideoStreams: (videoStreams) => {
    const localStreams = videoStreams.filter((s) => s.isLocal)

    set({
      videosStreams: videoStreams,
      localStreams,
    })
  },
  setPinnedStreamIds: (ids) => set({ pinnedStreamIds: ids }),
}))

export default useCurrentRoom
