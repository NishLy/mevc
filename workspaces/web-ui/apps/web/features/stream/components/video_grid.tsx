"use client"

import classNames from "classnames"
import useCurrentRoom from "../state"
import VideoTile from "./video_tile"
import { StreamVideoState } from "../types/stream"

const calculateGridColumns = (count: number) => {
  if (count === 1) return "grid-cols-1"
  if (count === 2) return "grid-cols-2"
  if (count <= 4) return "grid-cols-2"
  if (count <= 6) return "grid-cols-3"
  if (count <= 9) return "grid-cols-3"
  return "grid-cols-4"
}

export default function VideosGrid({
  streams,
}: {
  streams: StreamVideoState[]
}) {
  const pinnedStreamIds = useCurrentRoom((state) => state.pinnedStreamIds)
  const unpinnedStreams = streams.filter((s) => !pinnedStreamIds.includes(s.id))
  const pinnedStreams = streams.filter((s) => pinnedStreamIds.includes(s.id))

  console.log("Pinned streams:", pinnedStreams)
  return (
    <>
      <div
        className={classNames(
          "content-center gap-2 overflow-y-auto p-2",
          calculateGridColumns(unpinnedStreams.length),
          pinnedStreams.length > 0 &&
            "fixed top-0 left-0 z-10 grid h-44 w-fit justify-items-start overflow-x-auto overflow-y-hidden bg-transparent backdrop-blur-sm",
          pinnedStreams.length === 0 && "grid h-screen w-full rounded-lg p-4"
        )}
      >
        {unpinnedStreams.map((s) => (
          <VideoTile key={s.id} {...s} />
        ))}
      </div>

      {pinnedStreams.length > 0 && (
        <div
          className={classNames(
            "grid h-screen w-full content-center justify-items-center gap-2 overflow-y-auto p-2",
            calculateGridColumns(pinnedStreams.length),
            "z-10"
          )}
        >
          {pinnedStreams.map((s) => (
            <VideoTile key={s.id} {...s} />
          ))}
        </div>
      )}
    </>
  )
}
